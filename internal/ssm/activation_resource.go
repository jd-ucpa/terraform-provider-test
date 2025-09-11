package ssm

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &ActivationResource{}

// NewActivationResource crée et retourne une nouvelle instance de la ressource
// ActivationResource. Cette fonction est utilisée par le provider pour enregistrer
// la ressource dans Terraform.
func NewActivationResource() resource.Resource {
	return &ActivationResource{}
}

// ActivationResource gère la création et la gestion des activations SSM.
// Cette ressource permet de créer des activations SSM avec support du renouvellement
// automatique, de la gestion des secrets et de la configuration d'expiration.
type ActivationResource struct {
	ssm               *ssm.Client
	secretsManager    *secretsmanager.Client
}

// ExpirationDateModel définit le modèle pour le bloc expiration_date de la ressource.
// Il permet de configurer la durée d'expiration de l'activation SSM.
type ExpirationDateModel struct {
	Days    types.Int64 `tfsdk:"days"`
	Hours   types.Int64 `tfsdk:"hours"`
	Minutes types.Int64 `tfsdk:"minutes"`
}

// ActivationResourceModel définit le modèle de données pour la ressource Activation.
// Il contient tous les attributs de configuration et les données retournées par l'API SSM.
type ActivationResourceModel struct {
	Id              types.String         `tfsdk:"id"`
	Description     types.String         `tfsdk:"description"`
	IamRole         types.String         `tfsdk:"iam_role"`
	RegistrationLimit types.Int64        `tfsdk:"registration_limit"`
	Tags            types.Map            `tfsdk:"tags"`
	ActivationCode  types.String         `tfsdk:"activation_code"`
	ActivationId    types.String         `tfsdk:"activation_id"`
	Expired         types.Bool           `tfsdk:"expired"`
	ExpirationDate  *ExpirationDateModel `tfsdk:"expiration_date"`
	SecretName      types.String         `tfsdk:"secret_name"`
	SecretArn       types.String         `tfsdk:"secret_arn"`
	SecretVersion   types.String         `tfsdk:"secret_version"`
	Managed         types.Bool           `tfsdk:"managed"`
}

// Metadata définit le nom du type de ressource utilisé dans les configurations Terraform.
// Ce nom est utilisé pour référencer cette ressource dans les fichiers .tf.
func (r *ActivationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "test_ssm_activation"
}

// Configure initialise les clients SSM et Secrets Manager à partir de la configuration du provider.
// Cette méthode est appelée par Terraform pour configurer la ressource avec
// les paramètres d'authentification AWS (région, credentials, etc.).
func (r *ActivationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Éviter le panic si le provider n'a pas été configuré
	if req.ProviderData == nil {
		return
	}

	// Vérifier que la configuration est du bon type
	config, ok := req.ProviderData.(aws.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Provider configuration error",
			fmt.Sprintf("Expected aws.Config for SSM activation resource, got: %T. This indicates a provider configuration issue. Please verify your provider configuration and report this issue if it persists.", req.ProviderData),
		)
		return
	}
	
	// Créer les clients AWS à partir de la configuration
	r.ssm = ssm.NewFromConfig(config)
	r.secretsManager = secretsmanager.NewFromConfig(config)
}

// Schema définit la structure et la documentation de la ressource.
// Cette méthode décrit les attributs disponibles, leurs types, et leur documentation Markdown
// qui sera affichée dans la documentation Terraform.
func (r *ActivationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `test_ssm_activation` resource creates and manages SSM activations with automatic renewal, secret management, and expiration configuration. This resource automatically renews expired activations and can optionally store activation data in AWS Secrets Manager.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the SSM activation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "The description of the SSM activation.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"iam_role": schema.StringAttribute{
				MarkdownDescription: "The IAM role name to use for the SSM activation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"registration_limit": schema.Int64Attribute{
				MarkdownDescription: "The maximum number of managed instances that can be registered using this activation. Defaults to 1.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(1),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"activation_code": schema.StringAttribute{
				Computed:            true,
				Sensitive:           true,
				MarkdownDescription: "The activation code for the SSM activation.",
			},
			"activation_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The activation ID for the SSM activation.",
			},
			"expired": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the SSM activation has expired.",
			},
			"secret_name": schema.StringAttribute{
				MarkdownDescription: "The name of the AWS Secrets Manager secret to store activation data. Activation data will be automatically stored in this secret.",
				Required:            true,
			},
			"secret_arn": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ARN of the AWS Secrets Manager secret containing activation data.",
			},
			"secret_version": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The version ID of the AWS Secrets Manager secret containing activation data.",
			},
			"managed": schema.BoolAttribute{
				MarkdownDescription: "Whether the secret is managed by this resource. If true, the secret will be created and deleted by this resource. If false, the resource will only update an existing secret. Defaults to false.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"tags": schema.MapAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "A map of tags to assign to the SSM activation.",
				Optional:            true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"expiration_date": schema.SingleNestedBlock{
				MarkdownDescription: "Configuration for the expiration date of the SSM activation. The total duration cannot exceed 30 days.",
				Attributes: map[string]schema.Attribute{
					"days": schema.Int64Attribute{
						MarkdownDescription: "Number of days until expiration. Must be positive. Defaults to 30.",
						Optional:            true,
					},
					"hours": schema.Int64Attribute{
						MarkdownDescription: "Number of hours until expiration. Must be positive. Defaults to 0.",
						Optional:            true,
					},
					"minutes": schema.Int64Attribute{
						MarkdownDescription: "Number of minutes until expiration. Must be positive. Defaults to 0.",
						Optional:            true,
					},
				},
			},
		},
	}
}

// Create crée une nouvelle activation SSM.
// Cette méthode est appelée par Terraform lors de la création d'une ressource.
// Elle valide la configuration, crée l'activation SSM et gère le stockage des secrets.
func (r *ActivationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ActivationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Normaliser les valeurs optionnelles immédiatement
	r.normalizeOptionalValues(&data)

	// Valider que le secret existe si managed = false
	if !data.Managed.ValueBool() {
		if diag := r.validateSecretExists(ctx, data.SecretName.ValueString()); diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
	}

	// Valider la configuration d'expiration
	if diag := r.validateExpirationDate(data.ExpirationDate); diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	// Calculer la date d'expiration
	expirationDate := r.calculateExpirationDate(data.ExpirationDate, nil)

	// Préparer les tags
	tags, diag := r.convertTags(ctx, data.Tags)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	// Créer l'activation SSM
	createInput := &ssm.CreateActivationInput{
		IamRole: aws.String(data.IamRole.ValueString()),
	}

	if !data.Description.IsNull() {
		createInput.Description = aws.String(data.Description.ValueString())
	}

	if !data.RegistrationLimit.IsNull() {
		createInput.RegistrationLimit = aws.Int32(int32(data.RegistrationLimit.ValueInt64()))
	}

	if expirationDate != nil {
		createInput.ExpirationDate = expirationDate
	}

	if len(tags) > 0 {
		createInput.Tags = tags
	}

	createOutput, err := r.ssm.CreateActivation(ctx, createInput)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to create SSM activation",
			fmt.Sprintf("Error calling AWS SSM CreateActivation API: %s. Please verify your AWS credentials, permissions, and that you have the necessary IAM permissions to create SSM activations.", err),
		)
		return
	}

	// Mettre à jour le modèle avec les données retournées
	data.Id = types.StringValue(*createOutput.ActivationId)
	data.ActivationId = types.StringValue(*createOutput.ActivationId)
	data.ActivationCode = types.StringValue(*createOutput.ActivationCode)
	data.Expired = types.BoolValue(false)

	// Gérer le secret
	if diag := r.manageSecret(ctx, &data, true); diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	// S'assurer que les valeurs de secret sont toujours définies
	if data.SecretArn.IsNull() {
		data.SecretArn = types.StringValue("")
	}
	if data.SecretVersion.IsNull() {
		data.SecretVersion = types.StringValue("")
	}


	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read récupère l'état actuel de l'activation SSM depuis AWS.
// Cette méthode est appelée par Terraform pour synchroniser l'état local avec l'état distant.
// Elle vérifie si l'activation a expiré et la renouvelle automatiquement si nécessaire.
func (r *ActivationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ActivationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Normaliser les valeurs optionnelles immédiatement
	r.normalizeOptionalValues(&data)

	// Récupérer l'activation depuis AWS
	describeInput := &ssm.DescribeActivationsInput{
		Filters: []ssmtypes.DescribeActivationsFilter{
			{
				FilterKey:    ssmtypes.DescribeActivationsFilterKeysActivationIds,
				FilterValues: []string{data.ActivationId.ValueString()},
			},
		},
	}

	describeOutput, err := r.ssm.DescribeActivations(ctx, describeInput)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to retrieve SSM activation",
			fmt.Sprintf("Error calling AWS SSM DescribeActivations API for activation '%s': %s. Please verify your AWS credentials, permissions, and that the activation exists.", data.Id.ValueString(), err),
		)
		return
	}

	if len(describeOutput.ActivationList) == 0 {
		// L'activation n'existe plus, la marquer pour suppression
		resp.State.RemoveResource(ctx)
		return
	}

	activation := describeOutput.ActivationList[0]

	// Vérifier si l'activation a expiré
	expired := activation.Expired
	data.Expired = types.BoolValue(expired)

	// Si l'activation a expiré, la renouveler automatiquement
	if expired {
		
		// Créer une nouvelle activation
		createInput := &ssm.CreateActivationInput{
			IamRole: aws.String(data.IamRole.ValueString()),
		}

		if !data.Description.IsNull() {
			createInput.Description = aws.String(data.Description.ValueString())
		}

		if !data.RegistrationLimit.IsNull() {
			createInput.RegistrationLimit = aws.Int32(int32(data.RegistrationLimit.ValueInt64()))
		}

		// Recalculer la date d'expiration
		expirationDate := r.calculateExpirationDate(data.ExpirationDate, nil)
		if expirationDate != nil {
			createInput.ExpirationDate = expirationDate
		}

		// Préparer les tags
		tags, diag := r.convertTags(ctx, data.Tags)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
		if len(tags) > 0 {
			createInput.Tags = tags
		}

		createOutput, err := r.ssm.CreateActivation(ctx, createInput)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to renew expired SSM activation",
				fmt.Sprintf("Error calling AWS SSM CreateActivation API to renew activation '%s': %s. Please verify your AWS credentials, permissions, and that you have the necessary IAM permissions to create SSM activations.", data.Id.ValueString(), err),
			)
			return
		}

		// Mettre à jour avec la nouvelle activation
		data.Id = types.StringValue(*createOutput.ActivationId)
		data.ActivationId = types.StringValue(*createOutput.ActivationId)
		data.ActivationCode = types.StringValue(*createOutput.ActivationCode)
		data.Expired = types.BoolValue(false)

		// Mettre à jour le secret
		if diag := r.manageSecret(ctx, &data, false); diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
	}

	// S'assurer que les valeurs de secret sont toujours définies
	if data.SecretArn.IsNull() {
		data.SecretArn = types.StringValue("")
	}
	if data.SecretVersion.IsNull() {
		data.SecretVersion = types.StringValue("")
	}


	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update gère les modifications de l'activation SSM.
// Cette méthode est appelée par Terraform lors de la modification d'une ressource existante.
// Pour les activations SSM, la plupart des modifications nécessitent une recréation.
func (r *ActivationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ActivationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Normaliser les valeurs optionnelles immédiatement
	r.normalizeOptionalValues(&data)

	// Récupérer l'état actuel
	var currentData ActivationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &currentData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Vérifier si des attributs qui nécessitent une recréation ont changé
	needsRecreation := false

	if !data.IamRole.Equal(currentData.IamRole) {
		needsRecreation = true
	}

	if !data.Description.Equal(currentData.Description) {
		needsRecreation = true
	}

	if !data.RegistrationLimit.Equal(currentData.RegistrationLimit) {
		needsRecreation = true
	}

	// Vérifier si expiration_date a changé
	if !r.expirationDateEqual(data.ExpirationDate, currentData.ExpirationDate) {
		needsRecreation = true
	}

	// Vérifier si les tags ont changé
	if !data.Tags.Equal(currentData.Tags) {
		needsRecreation = true
	}

	// Vérifier si secret_name a changé
	if !data.SecretName.Equal(currentData.SecretName) {
		needsRecreation = true
	}

	// Vérifier si managed a changé
	if !data.Managed.Equal(currentData.Managed) {
		needsRecreation = true
	}

	if needsRecreation {
		// Supprimer l'ancienne activation
		if diag := r.deleteActivation(ctx, currentData.ActivationId.ValueString()); diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		// Supprimer l'ancien secret si géré
		if !currentData.SecretName.IsNull() && currentData.Managed.ValueBool() {
			if diag := r.deleteSecret(ctx, currentData.SecretName.ValueString()); diag.HasError() {
				resp.Diagnostics.Append(diag...)
				return
			}
		}

		// Créer une nouvelle activation
		// Valider que le secret existe si managed = false
		if !data.Managed.ValueBool() {
			if diag := r.validateSecretExists(ctx, data.SecretName.ValueString()); diag.HasError() {
				resp.Diagnostics.Append(diag...)
				return
			}
		}

		// Valider la configuration d'expiration
		if diag := r.validateExpirationDate(data.ExpirationDate); diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		// Calculer la date d'expiration
		expirationDate := r.calculateExpirationDate(data.ExpirationDate, nil)

		// Préparer les tags
		tags, diag := r.convertTags(ctx, data.Tags)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		// Créer l'activation SSM
		createInput := &ssm.CreateActivationInput{
			IamRole: aws.String(data.IamRole.ValueString()),
		}

		if !data.Description.IsNull() {
			createInput.Description = aws.String(data.Description.ValueString())
		}

		if !data.RegistrationLimit.IsNull() {
			createInput.RegistrationLimit = aws.Int32(int32(data.RegistrationLimit.ValueInt64()))
		}

		if expirationDate != nil {
			createInput.ExpirationDate = expirationDate
		}

		if len(tags) > 0 {
			createInput.Tags = tags
		}

		createOutput, err := r.ssm.CreateActivation(ctx, createInput)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to create new SSM activation",
				fmt.Sprintf("Error calling AWS SSM CreateActivation API: %s. Please verify your AWS credentials, permissions, and that you have the necessary IAM permissions to create SSM activations.", err),
			)
			return
		}

		// Mettre à jour le modèle avec les données retournées
		data.Id = types.StringValue(*createOutput.ActivationId)
		data.ActivationId = types.StringValue(*createOutput.ActivationId)
		data.ActivationCode = types.StringValue(*createOutput.ActivationCode)
		data.Expired = types.BoolValue(false)

		// Gérer le secret
		if diag := r.manageSecret(ctx, &data, true); diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		// La ressource a été recréée - c'est normal pour les activations SSM
	} else {
		// Seuls secret_name et managed peuvent être modifiés sans recréation
		// Mettre à jour le secret
		if diag := r.manageSecret(ctx, &data, false); diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}

		// Préserver les valeurs calculées
		data.Id = currentData.Id
		data.ActivationId = currentData.ActivationId
		data.ActivationCode = currentData.ActivationCode
		data.Expired = currentData.Expired
	}

	// S'assurer que les valeurs de secret sont toujours définies
	if data.SecretArn.IsNull() {
		data.SecretArn = types.StringValue("")
	}
	if data.SecretVersion.IsNull() {
		data.SecretVersion = types.StringValue("")
	}


	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete supprime l'activation SSM et le secret associé si géré.
// Cette méthode est appelée par Terraform lors de la suppression d'une ressource.
func (r *ActivationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ActivationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Supprimer l'activation SSM
	if diag := r.deleteActivation(ctx, data.ActivationId.ValueString()); diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	// Supprimer le secret si géré
	if data.Managed.ValueBool() {
		if diag := r.deleteSecret(ctx, data.SecretName.ValueString()); diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
	}
}

// validateExpirationDate valide la configuration d'expiration.
func (r *ActivationResource) validateExpirationDate(expirationDate *ExpirationDateModel) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	if expirationDate == nil {
		return diagnostics
	}

	// Récupérer les valeurs avec des valeurs par défaut
	days := int64(30)
	hours := int64(0)
	minutes := int64(0)

	if !expirationDate.Days.IsNull() && !expirationDate.Days.IsUnknown() {
		days = expirationDate.Days.ValueInt64()
	}
	if !expirationDate.Hours.IsNull() && !expirationDate.Hours.IsUnknown() {
		hours = expirationDate.Hours.ValueInt64()
	}
	if !expirationDate.Minutes.IsNull() && !expirationDate.Minutes.IsUnknown() {
		minutes = expirationDate.Minutes.ValueInt64()
	}

	// Vérifier que les valeurs sont positives
	if days < 0 {
		diagnostics.AddError("Invalid expiration_date configuration", "Days must be a positive number. Please specify a valid number of days.")
	}
	if hours < 0 {
		diagnostics.AddError("Invalid expiration_date configuration", "Hours must be a positive number. Please specify a valid number of hours.")
	}
	if minutes < 0 {
		diagnostics.AddError("Invalid expiration_date configuration", "Minutes must be a positive number. Please specify a valid number of minutes.")
	}

	// Vérifier que la durée totale ne dépasse pas 30 jours
	totalMinutes := days*24*60 + hours*60 + minutes
	maxMinutes := int64(30 * 24 * 60) // 30 jours en minutes

	if totalMinutes > maxMinutes {
		diagnostics.AddError(
			"Invalid expiration_date configuration",
			fmt.Sprintf("Total duration cannot exceed 30 days (got %d days, %d hours, %d minutes). Please reduce the expiration period to be within the 30-day limit.", days, hours, minutes),
		)
	}

	return diagnostics
}

// calculateExpirationDate calcule la date d'expiration basée sur la configuration.
// Si baseTime est fourni, l'expiration est calculée à partir de cette date.
// Sinon, elle est calculée à partir de time.Now().
func (r *ActivationResource) calculateExpirationDate(expirationDate *ExpirationDateModel, baseTime *time.Time) *time.Time {
	// Utiliser baseTime si fourni, sinon time.Now()
	var startTime time.Time
	if baseTime != nil {
		startTime = *baseTime
	} else {
		startTime = time.Now().UTC()
	}

	if expirationDate == nil {
		// Par défaut, 30 jours
		expiration := startTime.Add(30 * 24 * time.Hour)
		return &expiration
	}

	// Récupérer les valeurs avec des valeurs par défaut
	days := int64(30)
	hours := int64(0)
	minutes := int64(0)

	if !expirationDate.Days.IsNull() && !expirationDate.Days.IsUnknown() {
		days = expirationDate.Days.ValueInt64()
	}
	if !expirationDate.Hours.IsNull() && !expirationDate.Hours.IsUnknown() {
		hours = expirationDate.Hours.ValueInt64()
	}
	if !expirationDate.Minutes.IsNull() && !expirationDate.Minutes.IsUnknown() {
		minutes = expirationDate.Minutes.ValueInt64()
	}

	// Calculer la durée totale
	duration := time.Duration(days)*24*time.Hour +
		time.Duration(hours)*time.Hour +
		time.Duration(minutes)*time.Minute

	// Calculer la date d'expiration
	expiration := startTime.Add(duration)
	return &expiration
}

// expirationDateEqual compare deux configurations d'expiration.
func (r *ActivationResource) expirationDateEqual(a, b *ExpirationDateModel) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	return a.Days.Equal(b.Days) && a.Hours.Equal(b.Hours) && a.Minutes.Equal(b.Minutes)
}

// convertTags convertit les tags Terraform en format attendu par l'API SSM.
func (r *ActivationResource) convertTags(ctx context.Context, tags types.Map) ([]ssmtypes.Tag, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	if tags.IsNull() || tags.IsUnknown() {
		return []ssmtypes.Tag{}, diagnostics
	}

	tagMap := make(map[string]string)
	diagnostics.Append(tags.ElementsAs(ctx, &tagMap, false)...)
	if diagnostics.HasError() {
		return nil, diagnostics
	}

	ssmTags := make([]ssmtypes.Tag, 0, len(tagMap))
	for key, value := range tagMap {
		ssmTags = append(ssmTags, ssmtypes.Tag{
			Key:   aws.String(key),
			Value: aws.String(value),
		})
	}

	return ssmTags, diagnostics
}

// manageSecret gère la création, mise à jour ou suppression du secret.
func (r *ActivationResource) manageSecret(ctx context.Context, data *ActivationResourceModel, isCreate bool) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	secretName := data.SecretName.ValueString()
	managed := data.Managed.ValueBool()

	// Préparer les données du secret
	secretData := map[string]string{
		"activation_id":   data.ActivationId.ValueString(),
		"activation_code": data.ActivationCode.ValueString(),
	}

	secretJSON, err := json.Marshal(secretData)
	if err != nil {
		diagnostics.AddError(
			"Unable to serialize secret data",
			fmt.Sprintf("Error marshaling secret data to JSON: %s. Please verify the secret data format is valid.", err),
		)
		return diagnostics
	}

	if isCreate && managed {
		// Vérifier si le secret existe déjà
		_, err := r.secretsManager.DescribeSecret(ctx, &secretsmanager.DescribeSecretInput{
			SecretId: aws.String(secretName),
		})
		
		if err != nil {
			// Le secret n'existe pas, le créer
			createInput := &secretsmanager.CreateSecretInput{
				Name:         aws.String(secretName),
				SecretString: aws.String(string(secretJSON)),
				Description:  aws.String("SSM Activation data managed by terraform-provider-test"),
			}

			createOutput, err := r.secretsManager.CreateSecret(ctx, createInput)
			if err != nil {
				diagnostics.AddError(
					"Unable to create secret in Secrets Manager",
					fmt.Sprintf("Error calling AWS Secrets Manager CreateSecret API for secret '%s': %s. Please verify your AWS credentials, permissions, and that you have the necessary IAM permissions to create secrets.", secretName, err),
				)
				return diagnostics
			}

			data.SecretArn = types.StringValue(*createOutput.ARN)
			if createOutput.VersionId != nil {
				data.SecretVersion = types.StringValue(*createOutput.VersionId)
			}
		} else {
			// Le secret existe, le mettre à jour
			putInput := &secretsmanager.PutSecretValueInput{
				SecretId:     aws.String(secretName),
				SecretString: aws.String(string(secretJSON)),
			}

			putOutput, err := r.secretsManager.PutSecretValue(ctx, putInput)
			if err != nil {
				diagnostics.AddError(
					"Unable to update secret in Secrets Manager",
					fmt.Sprintf("Error calling AWS Secrets Manager PutSecretValue API for secret '%s': %s. Please verify your AWS credentials, permissions, and that the secret exists and is accessible.", secretName, err),
				)
				return diagnostics
			}

			// Récupérer les métadonnées du secret
			describeOutput, err := r.secretsManager.DescribeSecret(ctx, &secretsmanager.DescribeSecretInput{
				SecretId: aws.String(secretName),
			})
			if err != nil {
				diagnostics.AddError(
					"Unable to retrieve secret metadata",
					fmt.Sprintf("Error calling AWS Secrets Manager DescribeSecret API for secret '%s': %s. Please verify your AWS credentials, permissions, and that the secret exists.", secretName, err),
				)
				return diagnostics
			}

			data.SecretArn = types.StringValue(*describeOutput.ARN)
			if putOutput.VersionId != nil {
				data.SecretVersion = types.StringValue(*putOutput.VersionId)
			}
		}
	} else {
		// Mettre à jour le secret existant (managed = false ou update)
		putInput := &secretsmanager.PutSecretValueInput{
			SecretId:     aws.String(secretName),
			SecretString: aws.String(string(secretJSON)),
		}

		putOutput, err := r.secretsManager.PutSecretValue(ctx, putInput)
		if err != nil {
			diagnostics.AddError(
				"Unable to update secret in Secrets Manager",
				fmt.Sprintf("Error calling AWS Secrets Manager PutSecretValue API for secret '%s': %s. Please verify your AWS credentials, permissions, and that the secret exists and is accessible.", secretName, err),
			)
			return diagnostics
		}

		// Récupérer les métadonnées du secret
		describeOutput, err := r.secretsManager.DescribeSecret(ctx, &secretsmanager.DescribeSecretInput{
			SecretId: aws.String(secretName),
		})
		if err != nil {
			diagnostics.AddError(
				"Unable to retrieve secret metadata",
				fmt.Sprintf("Error calling AWS Secrets Manager DescribeSecret API for secret '%s': %s. Please verify your AWS credentials, permissions, and that the secret exists.", secretName, err),
			)
			return diagnostics
		}

		data.SecretArn = types.StringValue(*describeOutput.ARN)
		if putOutput.VersionId != nil {
			data.SecretVersion = types.StringValue(*putOutput.VersionId)
		}
	}

	return diagnostics
}

// deleteActivation supprime une activation SSM.
func (r *ActivationResource) deleteActivation(ctx context.Context, activationId string) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	deleteInput := &ssm.DeleteActivationInput{
		ActivationId: aws.String(activationId),
	}

	_, err := r.ssm.DeleteActivation(ctx, deleteInput)
	if err != nil {
		diagnostics.AddError(
			"Unable to delete SSM activation",
			fmt.Sprintf("Error calling AWS SSM DeleteActivation API for activation '%s': %s. Please verify your AWS credentials, permissions, and that the activation exists.", activationId, err),
		)
		return diagnostics
	}

	return diagnostics
}

// deleteSecret supprime un secret AWS Secrets Manager.
func (r *ActivationResource) deleteSecret(ctx context.Context, secretName string) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	deleteInput := &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(secretName),
		ForceDeleteWithoutRecovery: aws.Bool(true),
	}

	_, err := r.secretsManager.DeleteSecret(ctx, deleteInput)
	if err != nil {
		diagnostics.AddError(
			"Unable to delete secret from Secrets Manager",
			fmt.Sprintf("Error calling AWS Secrets Manager DeleteSecret API for secret '%s': %s. Please verify your AWS credentials, permissions, and that the secret exists.", secretName, err),
		)
		return diagnostics
	}

	return diagnostics
}

// validateSecretExists vérifie qu'un secret existe dans AWS Secrets Manager.
func (r *ActivationResource) validateSecretExists(ctx context.Context, secretName string) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	_, err := r.secretsManager.DescribeSecret(ctx, &secretsmanager.DescribeSecretInput{
		SecretId: aws.String(secretName),
	})
	if err != nil {
		diagnostics.AddError(
			"Secret not found in Secrets Manager",
			fmt.Sprintf("The secret '%s' does not exist in AWS Secrets Manager. When 'managed = false', the secret must already exist. Either create the secret first in AWS Secrets Manager or set 'managed = true' to let the provider create it.", secretName),
		)
		return diagnostics
	}

	return diagnostics
}

// normalizeOptionalValues s'assure que les valeurs optionnelles sont définies.
func (r *ActivationResource) normalizeOptionalValues(data *ActivationResourceModel) {
	if data.SecretName.IsNull() {
		data.SecretName = types.StringValue("")
	}
	if data.SecretArn.IsNull() {
		data.SecretArn = types.StringValue("")
	}
	if data.SecretVersion.IsNull() {
		data.SecretVersion = types.StringValue("")
	}
	if data.Tags.IsNull() {
		data.Tags = types.MapNull(types.StringType)
	}
	if data.Managed.IsNull() {
		data.Managed = types.BoolValue(false)
	}
}
