package ssm

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ActivationDataSource{}

// NewActivationDataSource crée et retourne une nouvelle instance du data source
// ActivationDataSource. Cette fonction est utilisée par le provider pour enregistrer
// le data source dans Terraform.
func NewActivationDataSource() datasource.DataSource {
	return &ActivationDataSource{}
}

// ActivationDataSource gère la récupération d'informations sur les activations SSM.
// Ce data source permet de récupérer les détails d'une activation SSM spécifique
// en utilisant son ID d'activation.
type ActivationDataSource struct {
	ssm *ssm.Client
}

// ActivationDataSourceModel définit le modèle de données pour le data source Activation.
// Il contient tous les attributs de configuration et les données retournées par l'API SSM.
type ActivationDataSourceModel struct {
	Id                  types.String `tfsdk:"id"`
	ActivationId        types.String `tfsdk:"activation_id"`
	IamRole             types.String `tfsdk:"iam_role"`
	RegistrationLimit   types.Int64  `tfsdk:"registration_limit"`
	RegistrationsCount  types.Int64  `tfsdk:"registrations_count"`
	ExpirationDate      types.String `tfsdk:"expiration_date"`
	Expired             types.Bool   `tfsdk:"expired"`
	CreatedDate         types.String `tfsdk:"created_date"`
}

// Metadata définit le nom du type de data source utilisé dans les configurations Terraform.
// Ce nom est utilisé pour référencer ce data source dans les fichiers .tf.
func (d *ActivationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "test_ssm_activation"
}

// Configure initialise le client SSM à partir de la configuration du provider.
// Cette méthode est appelée par Terraform pour configurer le data source avec
// les paramètres d'authentification AWS (région, credentials, etc.).
func (d *ActivationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Éviter le panic si le provider n'a pas été configuré
	if req.ProviderData == nil {
		return
	}

	// Vérifier que la configuration est du bon type
	config, ok := req.ProviderData.(aws.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Provider configuration error",
			fmt.Sprintf("Expected aws.Config for SSM activation data source, got: %T. This indicates a provider configuration issue. Please verify your provider configuration and report this issue if it persists.", req.ProviderData),
		)
		return
	}
	
	// Créer le client SSM à partir de la configuration AWS
	d.ssm = ssm.NewFromConfig(config)
}

// Schema définit la structure et la documentation du data source.
// Cette méthode décrit les attributs disponibles, leurs types, et leur documentation Markdown
// qui sera affichée dans la documentation Terraform.
func (d *ActivationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `test_ssm_activation` data source retrieves information about a specific SSM activation using its activation ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The activation ID (same as activation_id).",
			},
			"activation_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the SSM activation to retrieve information for.",
				Required:            true,
			},
			"iam_role": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The IAM role associated with the activation.",
			},
			"registration_limit": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The maximum number of managed instances that can be registered with this activation.",
			},
			"registrations_count": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The number of managed instances currently registered with this activation.",
			},
			"expiration_date": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time when the activation expires.",
			},
			"expired": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "Whether the activation has expired.",
			},
			"created_date": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The date and time when the activation was created.",
			},
		},
	}
}

// Read récupère les informations d'activation SSM en appelant l'API SSM DescribeActivations.
func (d *ActivationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ActivationDataSourceModel

	// Récupérer la configuration depuis la requête
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	activationId := data.ActivationId.ValueString()
	if activationId == "" {
		resp.Diagnostics.AddError(
			"Invalid activation_id configuration",
			"activation_id cannot be empty. Please provide a valid SSM activation ID.",
		)
		return
	}

	// Construire l'input pour l'API DescribeActivations
	input := &ssm.DescribeActivationsInput{
		Filters: []ssmtypes.DescribeActivationsFilter{
			{
				FilterKey:    ssmtypes.DescribeActivationsFilterKeysActivationIds,
				FilterValues: []string{activationId},
			},
		},
	}

	// Appeler l'API SSM DescribeActivations
	output, err := d.ssm.DescribeActivations(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to retrieve SSM activation",
			fmt.Sprintf("Error calling AWS SSM DescribeActivations API for activation '%s': %s. Please verify your AWS credentials, permissions, and that the activation exists.", activationId, err),
		)
		return
	}

	// Chercher l'activation avec l'ID spécifié
	var foundActivation *ssmtypes.Activation
	for _, activation := range output.ActivationList {
		if aws.ToString(activation.ActivationId) == activationId {
			foundActivation = &activation
			break
		}
	}

	// Vérifier si l'activation a été trouvée
	if foundActivation == nil {
		resp.Diagnostics.AddError(
			"SSM activation not found",
			fmt.Sprintf("No SSM activation found with ID '%s'. Please verify the activation ID is correct and that the activation exists in your AWS account.", activationId),
		)
		return
	}

	// Remplir le modèle avec les données de l'activation
	data.Id = types.StringValue(aws.ToString(foundActivation.ActivationId))
	data.ActivationId = types.StringValue(aws.ToString(foundActivation.ActivationId))
	data.IamRole = types.StringValue(aws.ToString(foundActivation.IamRole))
	data.RegistrationLimit = types.Int64Value(int64(aws.ToInt32(foundActivation.RegistrationLimit)))
	data.RegistrationsCount = types.Int64Value(int64(aws.ToInt32(foundActivation.RegistrationsCount)))
	
	// Formater les dates
	if foundActivation.ExpirationDate != nil {
		data.ExpirationDate = types.StringValue(foundActivation.ExpirationDate.Format(time.RFC3339))
	} else {
		data.ExpirationDate = types.StringNull()
	}
	
	if foundActivation.CreatedDate != nil {
		data.CreatedDate = types.StringValue(foundActivation.CreatedDate.Format(time.RFC3339))
	} else {
		data.CreatedDate = types.StringNull()
	}
	
	// Déterminer si l'activation a expiré
	if foundActivation.ExpirationDate != nil {
		data.Expired = types.BoolValue(time.Now().After(*foundActivation.ExpirationDate))
	} else {
		data.Expired = types.BoolValue(false)
	}

	// Sauvegarder les données dans l'état
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
