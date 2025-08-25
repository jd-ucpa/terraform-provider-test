package ssm

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &SendCommandResource{}

// NewSendCommandResource crée et retourne une nouvelle instance de la ressource
// SendCommandResource. Cette fonction est utilisée par le provider pour enregistrer
// la ressource dans Terraform.
func NewSendCommandResource() resource.Resource {
	return &SendCommandResource{}
}

// SendCommandResource gère l'envoi et le suivi de commandes SSM vers des instances EC2.
// Cette ressource permet d'exécuter des commandes sur des instances AWS en utilisant
// AWS Systems Manager (SSM) et surveille leur statut d'exécution.
type SendCommandResource struct {
	ssm *ssm.Client
}

// TargetResourceModel définit le modèle pour le bloc targets de la ressource.
// Il permet de cibler des instances EC2 par des critères spécifiques (tags, etc.)
// au lieu d'utiliser une liste directe d'instance IDs.
type TargetResourceModel struct {
	Key    types.String   `tfsdk:"key"`
	Values []types.String `tfsdk:"values"`
}

// SendCommandResourceModel définit le modèle de données pour la ressource SendCommand.
// Il contient tous les attributs de configuration et les données retournées par l'API SSM.
type SendCommandResourceModel struct {
	Id           types.String           `tfsdk:"id"`
	DocumentName types.String           `tfsdk:"document_name"`
	InstanceIds  types.List             `tfsdk:"instance_ids"`
	Targets      []TargetResourceModel  `tfsdk:"targets"`
	Parameters   types.Map              `tfsdk:"parameters"`
	Comment      types.String           `tfsdk:"comment"`
	CommandId    types.String           `tfsdk:"command_id"`
	Status       types.String           `tfsdk:"status"`
	Triggers     types.Map              `tfsdk:"triggers"`
}

// Metadata définit le nom du type de ressource utilisé dans les configurations Terraform.
// Ce nom est utilisé pour référencer cette ressource dans les fichiers .tf.
func (r *SendCommandResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "test_ssm_send_command"
}

// Configure initialise le client SSM à partir de la configuration du provider.
// Cette méthode est appelée par Terraform pour configurer la ressource avec
// les paramètres d'authentification AWS (région, credentials, etc.).
func (r *SendCommandResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Éviter le panic si le provider n'a pas été configuré
	if req.ProviderData == nil {
		return
	}

	// Vérifier que la configuration est du bon type
	config, ok := req.ProviderData.(aws.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *aws.Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}
	
	// Créer le client SSM à partir de la configuration AWS
	r.ssm = ssm.NewFromConfig(config)
}

// Schema définit la structure et la documentation de la ressource.
// Cette méthode décrit les attributs disponibles, leurs types, et leur documentation Markdown
// qui sera affichée dans la documentation Terraform.
func (r *SendCommandResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `test_ssm_send_command` resource allows you to send commands to EC2 instances using AWS Systems Manager (SSM). This resource supports targeting instances by instance IDs or by using target blocks for more flexible targeting options like EC2 tags.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: "Identifier",
			},
			"document_name": schema.StringAttribute{
				MarkdownDescription: "The name of the SSM document to use.",
				Required:            true,
			},
			"instance_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The list of instance IDs where the command should be executed. Either instance_ids or targets must be specified.",
				Optional:            true,
			},
			"parameters": schema.MapAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The parameters to pass to the SSM document.",
				Optional:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "A comment about the command.",
				Optional:            true,
			},
			"command_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the command that was sent.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of the command.",
			},
			"triggers": schema.MapAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "A map of arbitrary strings that, when changed, will force the resource to be recreated.",
				Optional:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"targets": schema.ListNestedBlock{
				MarkdownDescription: "The list of targets to send the command to. Either instance_ids or targets must be specified.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							MarkdownDescription: "The key of the target (e.g., 'InstanceIds', 'tag:Name', 'tag:Environment').",
							Required:            true,
						},
						"values": schema.ListAttribute{
							MarkdownDescription: "The values of the target.",
							Required:            true,
							ElementType:         types.StringType,
						},
					},
				},
			},
		},
	}
}

// Create envoie une nouvelle commande SSM vers les instances ciblées.
// Cette méthode est appelée par Terraform lors de la création d'une ressource.
// Elle valide la configuration, envoie la commande SSM et surveille son statut.
func (r *SendCommandResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SendCommandResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Valider et construire les targets
	targets, diag := r.validateAndBuildTargets(ctx, data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	// Convertir les paramètres
	parameters, diag := r.convertParameters(ctx, data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	// Exécuter la commande SSM
	data, diag = r.executeSSMCommand(ctx, data, targets, parameters)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	// Normaliser les valeurs optionnelles
	r.normalizeOptionalValues(&data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read récupère l'état actuel de la ressource depuis l'état Terraform.
// Cette méthode est appelée par Terraform pour synchroniser l'état local avec l'état distant.
// Pour cette ressource, l'état est conservé tel quel car les commandes SSM sont statiques.
func (r *SendCommandResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SendCommandResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Dans une implémentation réelle, vous vérifieriez le statut de la commande
	// Pour l'instant, on garde l'état actuel
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update gère les modifications de la ressource.
// Cette méthode est appelée par Terraform lors de la modification d'une ressource existante.
// Elle vérifie si les triggers ont changé et exécute une nouvelle commande si nécessaire.
func (r *SendCommandResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SendCommandResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Récupérer l'état actuel pour comparer les triggers
	var currentData SendCommandResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &currentData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Vérifier si les triggers ont changé
	triggersChanged := !data.Triggers.Equal(currentData.Triggers)

	if triggersChanged {
		// Si les triggers ont changé, exécuter une nouvelle commande SSM
		// Valider et construire les targets
		targets, diag := r.validateAndBuildTargets(ctx, data)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		// Convertir les paramètres
		parameters, diag := r.convertParameters(ctx, data)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		// Exécuter la commande SSM
		data, diag = r.executeSSMCommand(ctx, data, targets, parameters)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
	} else {
		// Si les triggers n'ont pas changé, préserver les valeurs calculées
		if data.Id.IsUnknown() || data.Id.IsNull() {
			data.Id = currentData.Id
		}
		if data.CommandId.IsUnknown() || data.CommandId.IsNull() {
			data.CommandId = currentData.CommandId
		}
		if data.Status.IsUnknown() || data.Status.IsNull() {
			data.Status = currentData.Status
		}
	}

	// S'assurer que les valeurs calculées sont toujours définies
	if data.Id.IsUnknown() || data.Id.IsNull() {
		data.Id = types.StringValue("")
	}
	if data.CommandId.IsUnknown() || data.CommandId.IsNull() {
		data.CommandId = types.StringValue("")
	}
	if data.Status.IsUnknown() || data.Status.IsNull() {
		data.Status = types.StringValue("")
	}

	// Normaliser les valeurs optionnelles
	r.normalizeOptionalValues(&data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SendCommandResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Les commandes SSM ne peuvent pas être supprimées, on ne fait rien
}

// PollCommandInvocation vérifie le statut d'une commande SSM
// Retourne des diagnostics avec :
// - Error : Erreur fatale (commande échouée)
// - Warning : Commande encore en cours (continuer polling)
// - Vide : Commande terminée avec succès
func PollCommandInvocation(ctx context.Context, client *ssm.Client, command *ssm.SendCommandOutput) diag.Diagnostics {
	diagnostics := diag.Diagnostics{}
	fmt.Printf("DEBUG: Polling command %s\n", *command.Command.CommandId)

	// Récupérer les détails de la commande
	listCommandInvocationsOutput, err := client.ListCommandInvocations(ctx, &ssm.ListCommandInvocationsInput{
		CommandId: command.Command.CommandId,
		Details:   true,
	})
	if err != nil {
		diagnostics.AddError("AWS Client Error", fmt.Sprintf("Unable to ListCommandInvocations, got error: %s", err))
		return diagnostics
	}

	invocationCount := int32(len(listCommandInvocationsOutput.CommandInvocations))
	fmt.Printf("DEBUG: Found %d invocations\n", invocationCount)
	
	if invocationCount == 0 {
		// Continue polling as the command invocation is not yet available and the api is eventually consistent
		diagnostics.AddWarning("Command Invocations Not Found", fmt.Sprintf("No command invocations found for command ID: %s", *command.Command.CommandId))
		return diagnostics
	}

	// Vérifier le statut de chaque invocation
	allCompleted := true
	hasErrors := false
	
	for _, invocation := range listCommandInvocationsOutput.CommandInvocations {
		fmt.Printf("DEBUG: Invocation %s status: %s\n", *invocation.InstanceId, invocation.Status)
		
		// Vérifier si la commande est encore en cours
		if invocation.Status == ssmtypes.CommandInvocationStatusPending ||
			invocation.Status == ssmtypes.CommandInvocationStatusInProgress ||
			invocation.Status == ssmtypes.CommandInvocationStatusDelayed ||
			invocation.Status == ssmtypes.CommandInvocationStatusCancelling {
			allCompleted = false
			diagnostics.AddWarning("Command Invocation In Progress", fmt.Sprintf("Command invocation is still in progress for command ID: %s, instance ID: %s", *command.Command.CommandId, *invocation.InstanceId))
		}

		// Vérifier les erreurs d'invocation
		if invocation.Status == ssmtypes.CommandInvocationStatusFailed ||
			invocation.Status == ssmtypes.CommandInvocationStatusTimedOut ||
			invocation.Status == ssmtypes.CommandInvocationStatusCancelled {
			hasErrors = true
			diagnostics.AddError("Command Invocation Failed", fmt.Sprintf("Command invocation failed for instance %s: %s", *invocation.InstanceId, *invocation.StatusDetails))
		}

		// Vérifier les plugins pour détecter les erreurs
		for _, plugin := range invocation.CommandPlugins {
			fmt.Printf("DEBUG: Plugin %s status: %s\n", *plugin.Name, plugin.Status)
			if plugin.Status == ssmtypes.CommandPluginStatusInProgress ||
				plugin.Status == ssmtypes.CommandPluginStatusPending {
				// Ces états indiquent que la commande est encore en cours
				allCompleted = false
			} else if plugin.Status != ssmtypes.CommandPluginStatusSuccess {
				// Seuls les états autres que Success, InProgress, Pending, Delayed sont des erreurs
				hasErrors = true
				diagnostics.AddError(
					fmt.Sprintf("Plugin %s failed: %s", *plugin.Name, *plugin.StatusDetails),
					fmt.Sprintf("Command: %s, Instance: %s, Output: %s", *command.Command.CommandId, *invocation.InstanceId, *plugin.Output),
				)
			}
		}
	}

	// Si il y a des erreurs, retourner immédiatement
	if hasErrors {
		fmt.Printf("DEBUG: Command has errors, returning error diagnostics\n")
		return diagnostics
	}

	// Si toutes les invocations ne sont pas terminées, continuer le polling
	if !allCompleted {
		fmt.Printf("DEBUG: Command still in progress, returning warning diagnostics\n")
		return diagnostics
	}

	// Si on arrive ici, toutes les invocations sont terminées avec succès
	fmt.Printf("DEBUG: All invocations completed successfully, returning empty diagnostics\n")
	return diag.Diagnostics{} // Retourner des diagnostics vides pour indiquer le succès
}

// getActualCommandStatus récupère le statut réel de la commande SSM depuis AWS.
// Cette fonction analyse les invocations de commande et leurs plugins pour déterminer
// le statut final de la commande (Success, Failed, TimedOut, Cancelled, Unknown).
func getActualCommandStatus(ctx context.Context, client *ssm.Client, command *ssm.SendCommandOutput) string {
	fmt.Printf("DEBUG: Getting actual status for command %s\n", *command.Command.CommandId)
	
	// Récupérer les détails de la commande
	listCommandInvocationsOutput, err := client.ListCommandInvocations(ctx, &ssm.ListCommandInvocationsInput{
		CommandId: command.Command.CommandId,
		Details:   true,
	})
	if err != nil {
		fmt.Printf("DEBUG: Error getting command invocations: %s\n", err)
		return "Unknown"
	}

	fmt.Printf("DEBUG: Found %d command invocations\n", len(listCommandInvocationsOutput.CommandInvocations))
	
	// Vérifier le statut de chaque invocation
	for _, invocation := range listCommandInvocationsOutput.CommandInvocations {
		fmt.Printf("DEBUG: Invocation status: %s\n", invocation.Status)
		
		// Vérifier les plugins pour détecter les erreurs
		for _, plugin := range invocation.CommandPlugins {
			fmt.Printf("DEBUG: Plugin %s status: %s\n", *plugin.Name, plugin.Status)
			if plugin.Status != ssmtypes.CommandPluginStatusSuccess {
				fmt.Printf("DEBUG: Plugin failed, returning Failed\n")
				return "Failed"
			}
		}

		// Vérifier le statut de l'invocation
		switch invocation.Status {
		case ssmtypes.CommandInvocationStatusSuccess:
			fmt.Printf("DEBUG: Invocation success, returning Success\n")
			return "Success"
		case ssmtypes.CommandInvocationStatusFailed:
			fmt.Printf("DEBUG: Invocation failed, returning Failed\n")
			return "Failed"
		case ssmtypes.CommandInvocationStatusTimedOut:
			fmt.Printf("DEBUG: Invocation timed out, returning TimedOut\n")
			return "TimedOut"
		case ssmtypes.CommandInvocationStatusCancelled:
			fmt.Printf("DEBUG: Invocation cancelled, returning Cancelled\n")
			return "Cancelled"
		default:
			fmt.Printf("DEBUG: Unknown invocation status, returning Unknown\n")
			return "Unknown"
		}
	}

	fmt.Printf("DEBUG: No invocations found, returning Unknown\n")
	return "Unknown"
}

// validateAndBuildTargets valide les paramètres et construit les targets pour l'API SSM
func (r *SendCommandResource) validateAndBuildTargets(ctx context.Context, data SendCommandResourceModel) ([]ssmtypes.Target, diag.Diagnostics) {
	var diagnostics diag.Diagnostics
	
	// Vérifier qu'exactement l'un des deux (instance_ids ou targets) est spécifié
	hasInstanceIds := !data.InstanceIds.IsNull()
	hasTargets := len(data.Targets) > 0
	
	if !hasInstanceIds && !hasTargets {
		diagnostics.AddError("Validation Error", "Either instance_ids or targets must be specified")
		return nil, diagnostics
	}
	
	if hasInstanceIds && hasTargets {
		diagnostics.AddError("Validation Error", "Cannot specify both instance_ids and targets. Use either instance_ids or targets, not both")
		return nil, diagnostics
	}

	var targets []ssmtypes.Target

	// Si instance_ids est spécifié, l'utiliser
	if !data.InstanceIds.IsNull() {
		instanceIds := make([]string, 0, len(data.InstanceIds.Elements()))
		diagnostics.Append(data.InstanceIds.ElementsAs(ctx, &instanceIds, false)...)
		if diagnostics.HasError() {
			return nil, diagnostics
		}

		targets = []ssmtypes.Target{
			{
				Key:    aws.String("InstanceIds"),
				Values: instanceIds,
			},
		}
	} else {
		// Sinon, utiliser les targets spécifiés
		targets = make([]ssmtypes.Target, len(data.Targets))
		for i, target := range data.Targets {
			targets[i] = ssmtypes.Target{
				Key:    target.Key.ValueStringPointer(),
				Values: make([]string, len(target.Values)),
			}
			for j, value := range target.Values {
				targets[i].Values[j] = value.ValueString()
			}
		}
	}

	return targets, diagnostics
}

// convertParameters convertit les paramètres Terraform en format attendu par l'API SSM
func (r *SendCommandResource) convertParameters(ctx context.Context, data SendCommandResourceModel) (map[string][]string, diag.Diagnostics) {
	var diagnostics diag.Diagnostics
	
	parameters := make(map[string][]string)
	if !data.Parameters.IsNull() {
		unparsed := make(map[string]string)
		err := data.Parameters.ElementsAs(ctx, &unparsed, false)
		if err != nil {
			diagnostics.AddError("Client Error", fmt.Sprintf("Unable to convert parameters, got error: %s", err))
			return nil, diagnostics
		}
		for k, v := range unparsed {
			parameters[k] = []string{v}
		}
	}
	
	return parameters, diagnostics
}

// executeSSMCommand exécute une commande SSM et gère le polling
func (r *SendCommandResource) executeSSMCommand(ctx context.Context, data SendCommandResourceModel, targets []ssmtypes.Target, parameters map[string][]string) (SendCommandResourceModel, diag.Diagnostics) {
	var diagnostics diag.Diagnostics
	
	// Envoyer la commande SSM
	command, err := r.ssm.SendCommand(ctx, &ssm.SendCommandInput{
		DocumentName: aws.String(data.DocumentName.ValueString()),
		Targets:      targets,
		Parameters:   parameters,
		Comment:      data.Comment.ValueStringPointer(),
	})
	if err != nil {
		diagnostics.AddError("Client Error", fmt.Sprintf("Unable to send command, got error: %s", err))
		return data, diagnostics
	}

	data.Id = types.StringValue(*command.Command.CommandId)
	data.CommandId = types.StringValue(*command.Command.CommandId)
	data.Status = types.StringValue("InProgress")

	// Polling de la commande avec timeout
	createTimeout := 5 * time.Minute
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	backoff := time.Second
	maxAttempts := 10 // Limiter le nombre de tentatives
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if ctx.Err() != nil {
			diagnostics.AddError("Operation Cancelled", fmt.Sprintf("Context cancelled before attempt %d: %s", attempt+1, ctx.Err()))
			return data, diagnostics
		}

		// Diagnostics with Severity warnings are treated as retriable errors
		attemptDiag := PollCommandInvocation(ctx, r.ssm, command)
		if attemptDiag.HasError() {
			// La commande SSM a échoué, mais on ne fait pas échouer terraform apply
			// On met juste le statut à "Failed" et on continue
			fmt.Printf("DEBUG: SSM command failed, setting status to Failed\n")
			data.Status = types.StringValue("Failed")
			return data, diagnostics
		} else if attemptDiag.WarningsCount() > 0 {
			// Retry with exponential backoff
			select {
			case <-time.After(backoff):
				// Exponential backoff with a maximum of 30 seconds
				backoff *= 2
				if backoff > 30*time.Second {
					backoff = 30 * time.Second
				}
				continue
			case <-ctx.Done():
				if errors.Is(ctx.Err(), context.DeadlineExceeded) {
					diagnostics.AddError("Timeout while waiting for SSM Command to complete", fmt.Sprintf("Timeout occurred while waiting on command %s (polled %d times over %s)", *command.Command.CommandId, attempt+1, createTimeout))
					return data, diagnostics
				} else {
					diagnostics.AddError("Operation Cancelled", fmt.Sprintf("Context cancelled before attempt %d: %s", attempt+1, ctx.Err()))
					return data, diagnostics
				}
			}
		} else {
			// Command completed - get the actual status from AWS
			fmt.Printf("DEBUG: Polling completed, getting actual status\n")
			actualStatus := getActualCommandStatus(ctx, r.ssm, command)
			fmt.Printf("DEBUG: Actual status: %s\n", actualStatus)
			data.Status = types.StringValue(actualStatus)
			return data, diagnostics
		}
	}
	
	// Si on arrive ici, on a dépassé le nombre max de tentatives
	fmt.Printf("DEBUG: Max attempts reached, getting final status\n")
	actualStatus := getActualCommandStatus(ctx, r.ssm, command)
	fmt.Printf("DEBUG: Final status: %s\n", actualStatus)
	data.Status = types.StringValue(actualStatus)
	return data, diagnostics
}

// normalizeOptionalValues s'assure que les valeurs optionnelles sont définies
func (r *SendCommandResource) normalizeOptionalValues(data *SendCommandResourceModel) {
	if data.Parameters.IsNull() {
		data.Parameters = types.MapNull(types.StringType)
	}
	if data.Comment.IsNull() {
		data.Comment = types.StringValue("")
	}
	if data.Triggers.IsNull() {
		data.Triggers = types.MapNull(types.StringType)
	}
}
