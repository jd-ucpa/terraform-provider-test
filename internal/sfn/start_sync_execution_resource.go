package sfn

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &StartSyncExecutionResource{}

// NewStartSyncExecutionResource crée et retourne une nouvelle instance de la ressource
// StartSyncExecutionResource. Cette fonction est utilisée par le provider pour enregistrer
// la ressource dans Terraform.
func NewStartSyncExecutionResource() resource.Resource {
	return &StartSyncExecutionResource{}
}

// StartSyncExecutionResource gère l'exécution synchrone d'une state machine AWS Step Functions.
// Cette ressource permet d'exécuter une state machine et d'attendre le résultat de manière synchrone.
type StartSyncExecutionResource struct {
	sfn *sfn.Client
}

// StartSyncExecutionResourceModel définit le modèle de données pour la ressource StartSyncExecution.
// Il contient tous les attributs de configuration et les données retournées par l'API SFN.
type StartSyncExecutionResourceModel struct {
	Id               types.String `tfsdk:"id"`
	StateMachineArn  types.String `tfsdk:"state_machine_arn"`
	Name             types.String `tfsdk:"name"`
	Input            types.String `tfsdk:"input"`
	Triggers         types.Map    `tfsdk:"triggers"`
	ExecutionArn     types.String `tfsdk:"execution_arn"`
	Status           types.String `tfsdk:"status"`
	Output           types.String `tfsdk:"output"`
	Error            types.String `tfsdk:"error"`
	Cause            types.String `tfsdk:"cause"`
	BillingDetails   types.Object `tfsdk:"billing_details"`
	StartDate        types.String `tfsdk:"start_date"`
	StopDate         types.String `tfsdk:"stop_date"`
}

// BillingDetailsModel définit le modèle pour les détails de facturation.
type BillingDetailsModel struct {
	BilledDurationInMilliseconds types.Int64 `tfsdk:"billed_duration_in_milliseconds"`
	BilledMemoryUsedInMB         types.Int64 `tfsdk:"billed_memory_used_in_mb"`
}

// Metadata définit le nom du type de ressource utilisé dans les configurations Terraform.
// Ce nom est utilisé pour référencer cette ressource dans les fichiers .tf.
func (r *StartSyncExecutionResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "test_sfn_start_sync_execution"
}

// Configure initialise le client SFN à partir de la configuration du provider.
// Cette méthode est appelée par Terraform pour configurer la ressource avec
// les paramètres d'authentification AWS (région, credentials, etc.).
func (r *StartSyncExecutionResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Éviter le panic si le provider n'a pas été configuré
	if req.ProviderData == nil {
		return
	}

	// Vérifier que la configuration est du bon type
	config, ok := req.ProviderData.(aws.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Provider configuration error",
			fmt.Sprintf("Expected aws.Config for SFN start sync execution resource, got: %T. This indicates a provider configuration issue. Please verify your provider configuration and report this issue if it persists.", req.ProviderData),
		)
		return
	}
	
	// Créer le client SFN à partir de la configuration AWS
	r.sfn = sfn.NewFromConfig(config)
}

// Schema définit la structure et la documentation de la ressource.
// Cette méthode décrit les attributs disponibles, leurs types, et leur documentation Markdown
// qui sera affichée dans la documentation Terraform.
func (r *StartSyncExecutionResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `test_sfn_start_sync_execution` resource allows you to start a synchronous execution of an AWS Step Functions state machine. This resource waits for the execution to complete and returns the result.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The unique identifier for this execution.",
			},
			"state_machine_arn": schema.StringAttribute{
				MarkdownDescription: "The ARN of the state machine to execute.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the execution. If not provided, AWS will generate a unique name.",
				Optional:            true,
			},
			"input": schema.StringAttribute{
				MarkdownDescription: "The JSON input data for the execution. Defaults to `{}` if not provided.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("{}"),
			},
			"triggers": schema.MapAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "A map of arbitrary strings that, when changed, will force the resource to be recreated.",
				Optional:            true,
			},
			"execution_arn": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ARN of the execution.",
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of the execution (SUCCEEDED, FAILED, TIMED_OUT, ABORTED).",
			},
			"output": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The JSON output data of the execution.",
			},
			"error": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The error name if the execution failed (e.g., States.DataLimitExceeded).",
			},
			"cause": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The error cause if the execution failed, providing detailed information about the failure.",
			},
			"billing_details": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The billing details of the execution.",
				Attributes: map[string]schema.Attribute{
					"billed_duration_in_milliseconds": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "The billed duration of the execution in milliseconds.",
					},
					"billed_memory_used_in_mb": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "The billed memory used in MB.",
					},
				},
			},
			"start_date": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time when the execution started.",
			},
			"stop_date": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time when the execution stopped.",
			},
		},
	}
}

// Create démarre une nouvelle exécution synchrone de la state machine.
// Cette méthode est appelée par Terraform lors de la création d'une ressource.
// Elle valide la configuration, démarre l'exécution SFN et attend le résultat.
func (r *StartSyncExecutionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data StartSyncExecutionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Utiliser l'input (la valeur par défaut "{}" est gérée par le schéma)
	input := data.Input.ValueString()

	// Préparer les paramètres pour l'API SFN
	inputParams := &sfn.StartSyncExecutionInput{
		StateMachineArn: aws.String(data.StateMachineArn.ValueString()),
		Input:           aws.String(input),
	}

	// Ajouter le nom si spécifié
	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		inputParams.Name = aws.String(data.Name.ValueString())
	}

	// Démarrer l'exécution synchrone
	result, err := r.sfn.StartSyncExecution(ctx, inputParams)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to start SFN sync execution",
			fmt.Sprintf("Error calling AWS Step Functions StartSyncExecution API for state machine '%s': %s. Please verify your AWS credentials, permissions, and that the state machine exists and is accessible.", data.StateMachineArn.ValueString(), err),
		)
		return
	}

	// Remplir les données de réponse
	data.Id = types.StringValue(*result.ExecutionArn)
	data.ExecutionArn = types.StringValue(*result.ExecutionArn)
	data.Status = types.StringValue(string(result.Status))
	
	if result.Output != nil {
		data.Output = types.StringValue(*result.Output)
	} else {
		data.Output = types.StringValue("")
	}

	// Gérer les champs d'erreur
	if result.Error != nil {
		data.Error = types.StringValue(*result.Error)
	} else {
		data.Error = types.StringNull()
	}

	if result.Cause != nil {
		data.Cause = types.StringValue(*result.Cause)
	} else {
		data.Cause = types.StringNull()
	}

	// Préserver la valeur d'input du plan (ne pas la modifier)
	// data.Input reste inchangé car il vient du plan


	// Remplir les détails de facturation
	if result.BillingDetails != nil {
		billingDetails := BillingDetailsModel{
			BilledDurationInMilliseconds: types.Int64Value(result.BillingDetails.BilledDurationInMilliseconds),
			BilledMemoryUsedInMB:         types.Int64Value(result.BillingDetails.BilledMemoryUsedInMB),
		}
		billingDetailsObj, diag := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"billed_duration_in_milliseconds": types.Int64Type,
			"billed_memory_used_in_mb":         types.Int64Type,
		}, billingDetails)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
		data.BillingDetails = billingDetailsObj
	} else {
		// Créer un objet vide pour les détails de facturation
		billingDetailsObj, diag := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"billed_duration_in_milliseconds": types.Int64Type,
			"billed_memory_used_in_mb":         types.Int64Type,
		}, BillingDetailsModel{
			BilledDurationInMilliseconds: types.Int64Value(0),
			BilledMemoryUsedInMB:         types.Int64Value(0),
		})
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
		data.BillingDetails = billingDetailsObj
	}

	// Remplir les dates
	if result.StartDate != nil {
		data.StartDate = types.StringValue(result.StartDate.Format("2006-01-02T15:04:05.000Z"))
	} else {
		data.StartDate = types.StringValue("")
	}

	if result.StopDate != nil {
		data.StopDate = types.StringValue(result.StopDate.Format("2006-01-02T15:04:05.000Z"))
	} else {
		data.StopDate = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read récupère l'état actuel de la ressource depuis l'état Terraform.
// Cette méthode est appelée par Terraform pour synchroniser l'état local avec l'état distant.
// Pour cette ressource, l'état est conservé tel quel car les exécutions SFN sont statiques.
func (r *StartSyncExecutionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StartSyncExecutionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Dans une implémentation réelle, vous pourriez vérifier le statut de l'exécution
	// Pour l'instant, on garde l'état actuel
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update gère les modifications de la ressource.
// Cette méthode est appelée par Terraform lors de la modification d'une ressource existante.
// Elle vérifie si les triggers ont changé et exécute une nouvelle exécution si nécessaire.
func (r *StartSyncExecutionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data StartSyncExecutionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Récupérer l'état actuel pour comparer les triggers
	var currentData StartSyncExecutionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &currentData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Vérifier si les triggers ont changé
	triggersChanged := !data.Triggers.Equal(currentData.Triggers)

	if triggersChanged {

	// Utiliser l'input (la valeur par défaut "{}" est gérée par le schéma)
	input := data.Input.ValueString()

	// Préparer les paramètres pour l'API SFN
	inputParams := &sfn.StartSyncExecutionInput{
		StateMachineArn: aws.String(data.StateMachineArn.ValueString()),
		Input:           aws.String(input),
	}

	// Ajouter le nom si spécifié
	if !data.Name.IsNull() && !data.Name.IsUnknown() {
		inputParams.Name = aws.String(data.Name.ValueString())
	}

	// Démarrer l'exécution synchrone
	result, err := r.sfn.StartSyncExecution(ctx, inputParams)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to start SFN sync execution",
			fmt.Sprintf("Error calling AWS Step Functions StartSyncExecution API for state machine '%s': %s. Please verify your AWS credentials, permissions, and that the state machine exists and is accessible.", data.StateMachineArn.ValueString(), err),
		)
		return
	}

	// Mettre à jour les données avec le résultat
	data.Id = types.StringValue(*result.ExecutionArn)
	data.ExecutionArn = types.StringValue(*result.ExecutionArn)
	data.Status = types.StringValue(string(result.Status))
	
	// Gérer le cas où Output peut être nil (ex: quand la Step Function échoue)
	if result.Output != nil {
		data.Output = types.StringValue(*result.Output)
	} else {
		data.Output = types.StringValue("")
	}

	// Gérer les champs d'erreur
	if result.Error != nil {
		data.Error = types.StringValue(*result.Error)
	} else {
		data.Error = types.StringNull()
	}

	if result.Cause != nil {
		data.Cause = types.StringValue(*result.Cause)
	} else {
		data.Cause = types.StringNull()
	}

	// Gérer les détails de facturation
	if result.BillingDetails != nil {
		billingDetails := BillingDetailsModel{
			BilledDurationInMilliseconds: types.Int64Value(result.BillingDetails.BilledDurationInMilliseconds),
			BilledMemoryUsedInMB:         types.Int64Value(result.BillingDetails.BilledMemoryUsedInMB),
		}

		billingDetailsObj, diag := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"billed_duration_in_milliseconds": types.Int64Type,
			"billed_memory_used_in_mb":         types.Int64Type,
		}, billingDetails)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
		data.BillingDetails = billingDetailsObj
	} else {
		// Créer un objet vide pour les détails de facturation
		billingDetailsObj, diag := types.ObjectValueFrom(ctx, map[string]attr.Type{
			"billed_duration_in_milliseconds": types.Int64Type,
			"billed_memory_used_in_mb":         types.Int64Type,
		}, BillingDetailsModel{
			BilledDurationInMilliseconds: types.Int64Value(0),
			BilledMemoryUsedInMB:         types.Int64Value(0),
		})
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
		data.BillingDetails = billingDetailsObj
	}

	// Remplir les dates
	if result.StartDate != nil {
		data.StartDate = types.StringValue(result.StartDate.Format("2006-01-02T15:04:05.000Z"))
	} else {
		data.StartDate = types.StringValue("")
	}

	if result.StopDate != nil {
		data.StopDate = types.StringValue(result.StopDate.Format("2006-01-02T15:04:05.000Z"))
	} else {
		data.StopDate = types.StringValue("")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	} else {
		// Si les triggers n'ont pas changé, préserver les valeurs calculées
		if data.Id.IsUnknown() || data.Id.IsNull() {
			data.Id = currentData.Id
		}
		if data.ExecutionArn.IsUnknown() || data.ExecutionArn.IsNull() {
			data.ExecutionArn = currentData.ExecutionArn
		}
		if data.Status.IsUnknown() || data.Status.IsNull() {
			data.Status = currentData.Status
		}
		if data.Output.IsUnknown() || data.Output.IsNull() {
			data.Output = currentData.Output
		}
		if data.Error.IsUnknown() || data.Error.IsNull() {
			data.Error = currentData.Error
		}
		if data.Cause.IsUnknown() || data.Cause.IsNull() {
			data.Cause = currentData.Cause
		}
		if data.BillingDetails.IsUnknown() || data.BillingDetails.IsNull() {
			data.BillingDetails = currentData.BillingDetails
		}
		if data.StartDate.IsUnknown() || data.StartDate.IsNull() {
			data.StartDate = currentData.StartDate
		}
		if data.StopDate.IsUnknown() || data.StopDate.IsNull() {
			data.StopDate = currentData.StopDate
		}

		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	}
}

// Delete gère la suppression de la ressource.
// Cette méthode est appelée par Terraform lors de la suppression d'une ressource.
// Les exécutions SFN ne peuvent pas être supprimées, on ne fait rien.
func (r *StartSyncExecutionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Les exécutions SFN ne peuvent pas être supprimées, on ne fait rien
}

