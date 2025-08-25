package sts

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure CallerIdentityDataSource satisfies various datasource interfaces.
var _ datasource.DataSource = &CallerIdentityDataSource{}

// NewCallerIdentityDataSource crée et retourne une nouvelle instance du data source
// CallerIdentityDataSource. Cette fonction est utilisée par le provider pour enregistrer
// le data source dans Terraform.
func NewCallerIdentityDataSource() datasource.DataSource {
	return &CallerIdentityDataSource{}
}

// CallerIdentityDataSource récupère les informations d'identité AWS actuelles
// via l'API STS GetCallerIdentity. Il fournit l'Account ID, User ID et ARN
// de l'utilisateur/rôle actuellement authentifié.
type CallerIdentityDataSource struct {
	sts *sts.Client
}

// CallerIdentityDataSourceModel définit le modèle de données pour le data source
// CallerIdentity. Il contient les attributs retournés par l'API STS GetCallerIdentity.
type CallerIdentityDataSourceModel struct {
	AccountID types.String `tfsdk:"account_id"`
	ARN       types.String `tfsdk:"arn"`
	ID        types.String `tfsdk:"id"`
	UserID    types.String `tfsdk:"user_id"`
}

// Metadata définit le nom du type de data source utilisé dans les configurations Terraform.
// Ce nom est utilisé pour référencer ce data source dans les fichiers .tf.
func (d *CallerIdentityDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "test_caller_identity"
}

// Configure initialise le client STS à partir de la configuration du provider.
// Cette méthode est appelée par Terraform pour configurer le data source avec
// les paramètres d'authentification AWS (région, credentials, etc.).
func (d *CallerIdentityDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Éviter le panic si le provider n'a pas été configuré
	if req.ProviderData == nil {
		return
	}

	// Vérifier que la configuration est du bon type
	config, ok := req.ProviderData.(aws.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			"Expected aws.Config, got: %T. Please report this issue to the provider developers.",
		)
		return
	}
	
	// Créer le client STS à partir de la configuration AWS
	d.sts = sts.NewFromConfig(config)
}

// Schema définit la structure et la documentation du data source.
// Cette méthode décrit les attributs disponibles et leur documentation Markdown
// qui sera affichée dans la documentation Terraform.
func (d *CallerIdentityDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to get the access to the effective Account ID, User ID, and ARN in which Terraform is authorized. This data source calls the AWS STS GetCallerIdentity API to retrieve information about the current AWS identity.",
		Attributes: map[string]schema.Attribute{
			"account_id": schema.StringAttribute{
				MarkdownDescription: "The ID of the AWS account that the identity belongs to.",
				Computed:            true,
			},
			"arn": schema.StringAttribute{
				MarkdownDescription: "The ARN of the identity.",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the identity.",
				Computed:            true,
			},
			"user_id": schema.StringAttribute{
				MarkdownDescription: "The unique ID of the identity.",
				Computed:            true,
			},
		},
	}
}

// Read récupère les informations d'identité AWS en appelant l'API STS GetCallerIdentity.
// Cette méthode est appelée par Terraform pour obtenir les données du data source.
// Elle valide la configuration, appelle l'API AWS et mappe les résultats.
func (d *CallerIdentityDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CallerIdentityDataSourceModel

	// Parser la configuration du data source
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Vérifier si le client STS est configuré
	if d.sts == nil {
		resp.Diagnostics.AddError(
			"STS Client Not Configured",
			"Unable to get caller identity: STS client is not configured.",
		)
		return
	}

	// Appeler l'API STS GetCallerIdentity pour récupérer les informations d'identité
	output, err := d.sts.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get caller identity",
			"An error occurred while calling STS GetCallerIdentity: "+err.Error(),
		)
		return
	}

	// Mapper les données de réponse vers le modèle Terraform
	if output.Account != nil {
		data.AccountID = types.StringValue(aws.ToString(output.Account))
		data.ID = types.StringValue(aws.ToString(output.Account)) // ID = AccountID pour ce data source
	}
	if output.Arn != nil {
		data.ARN = types.StringValue(aws.ToString(output.Arn))
	}
	if output.UserId != nil {
		data.UserID = types.StringValue(aws.ToString(output.UserId))
	}

	// Sauvegarder les données dans l'état Terraform
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

