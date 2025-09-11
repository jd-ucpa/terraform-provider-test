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
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &ActivationsDataSource{}

// NewActivationsDataSource crée et retourne une nouvelle instance du data source
// ActivationsDataSource. Cette fonction est utilisée par le provider pour enregistrer
// le data source dans Terraform.
func NewActivationsDataSource() datasource.DataSource {
	return &ActivationsDataSource{}
}

// ActivationsDataSource gère la récupération d'informations sur plusieurs activations SSM.
// Ce data source permet de récupérer les détails de plusieurs activations SSM
// en utilisant des filtres et en gérant la pagination.
type ActivationsDataSource struct {
	ssm *ssm.Client
}

// ActivationsDataSourceModel définit le modèle de données pour le data source Activations.
// Il contient tous les attributs de configuration et les données retournées par l'API SSM.
type ActivationsDataSourceModel struct {
	Id         types.String           `tfsdk:"id"`
	Expired    types.Bool             `tfsdk:"expired"`
	Filters    []FilterModel          `tfsdk:"filter"`
	Activations []ActivationModel     `tfsdk:"activations"`
}

// FilterModel définit le modèle pour les filtres d'activation.
type FilterModel struct {
	Name   types.String   `tfsdk:"name"`
	Values []types.String `tfsdk:"values"`
}

// ActivationModel définit le modèle pour une activation individuelle.
type ActivationModel struct {
	ActivationId       types.String `tfsdk:"activation_id"`
	IamRole            types.String `tfsdk:"iam_role"`
	RegistrationLimit  types.Int64  `tfsdk:"registration_limit"`
	RegistrationsCount types.Int64  `tfsdk:"registrations_count"`
	ExpirationDate     types.String `tfsdk:"expiration_date"`
	Expired            types.Bool   `tfsdk:"expired"`
	CreatedDate        types.String `tfsdk:"created_date"`
}

// filterValuesValidator valide que le tableau de valeurs de filtre n'est pas vide.
type filterValuesValidator struct{}

// Description retourne la description du validateur.
func (v filterValuesValidator) Description(ctx context.Context) string {
	return "Filter values must not be empty"
}

// MarkdownDescription retourne la description Markdown du validateur.
func (v filterValuesValidator) MarkdownDescription(ctx context.Context) string {
	return "Filter values must not be empty"
}

// ValidateList valide que la liste n'est pas vide.
func (v filterValuesValidator) ValidateList(ctx context.Context, req validator.ListRequest, resp *validator.ListResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if len(req.ConfigValue.Elements()) == 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid filter values configuration",
			"Filter requires at least one value",
		)
	}
}

// Metadata définit le nom du type de data source utilisé dans les configurations Terraform.
// Ce nom est utilisé pour référencer ce data source dans les fichiers .tf.
func (d *ActivationsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "test_ssm_activations"
}

// Configure initialise le client SSM à partir de la configuration du provider.
// Cette méthode est appelée par Terraform pour configurer le data source avec
// les paramètres d'authentification AWS (région, credentials, etc.).
func (d *ActivationsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Éviter le panic si le provider n'a pas été configuré
	if req.ProviderData == nil {
		return
	}

	// Vérifier que la configuration est du bon type
	config, ok := req.ProviderData.(aws.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Provider configuration error",
			fmt.Sprintf("Expected aws.Config for SSM activations data source, got: %T. This indicates a provider configuration issue. Please verify your provider configuration and report this issue if it persists.", req.ProviderData),
		)
		return
	}
	
	// Créer le client SSM à partir de la configuration AWS
	d.ssm = ssm.NewFromConfig(config)
}

// Schema définit la structure et la documentation du data source.
// Cette méthode décrit les attributs disponibles, leurs types, et leur documentation Markdown
// qui sera affichée dans la documentation Terraform.
func (d *ActivationsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `test_ssm_activations` data source retrieves information about multiple SSM activations using filters and pagination. This data source calls the AWS SSM DescribeActivations API to retrieve activation details including IAM role, registration limits, expiration date, and other metadata.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the data source (always 'ssm_activations').",
			},
			"expired": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Filter activations by expiration status. If not specified, returns both expired and non-expired activations.",
			},
			"activations": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of SSM activations matching the specified filters.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"activation_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The ID of the SSM activation.",
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
				},
			},
		},
		Blocks: map[string]schema.Block{
			"filter": schema.ListNestedBlock{
				MarkdownDescription: "Filter activations by specific criteria. Multiple filters are cumulative (AND operation).",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:            true,
							MarkdownDescription: "The name of the filter. Valid values are: 'activation-ids', 'iam-role', 'default-instance-name'.",
						},
						"values": schema.ListAttribute{
							Required:            true,
							ElementType:         types.StringType,
							MarkdownDescription: "The values for the filter.",
							Validators: []validator.List{
								filterValuesValidator{},
							},
						},
					},
				},
			},
		},
	}
}

// Read récupère les informations d'activations SSM en appelant l'API SSM DescribeActivations.
func (d *ActivationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ActivationsDataSourceModel

	// Récupérer la configuration depuis la requête
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Construire les filtres pour l'API SSM
	filters, diag := d.buildFilters(data.Filters)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	// Boucle de pagination pour récupérer toutes les activations
	var nextToken *string
	var allActivations []ssmtypes.Activation

	for {
		// Construire l'input pour l'API DescribeActivations
		input := &ssm.DescribeActivationsInput{
			Filters: filters,
		}

		// Ajouter le NextToken si disponible
		if nextToken != nil {
			input.NextToken = nextToken
		}

		// Appeler l'API SSM DescribeActivations
		output, err := d.ssm.DescribeActivations(ctx, input)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to retrieve SSM activations",
				fmt.Sprintf("Error calling AWS SSM DescribeActivations API: %s. Please verify your AWS credentials, permissions, and that the SSM service is available in your region.", err),
			)
			return
		}

		// Ajouter les activations à la liste
		allActivations = append(allActivations, output.ActivationList...)

		// Vérifier s'il y a plus de pages
		if output.NextToken == nil {
			break
		}
		nextToken = output.NextToken
	}

	// Convertir les activations en modèle Terraform
	activations := d.convertActivations(allActivations)

	// Appliquer le filtre expired si spécifié
	if !data.Expired.IsNull() {
		activations = d.filterByExpired(activations, data.Expired.ValueBool())
	}

	// S'assurer que activations est toujours un array (même vide)
	if activations == nil {
		activations = []ActivationModel{}
	}

	// Remplir le modèle avec les données
	data.Id = types.StringValue("ssm_activations")
	data.Activations = activations

	// Sauvegarder les données dans l'état
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// buildFilters construit les filtres pour l'API SSM à partir des filtres Terraform.
func (d *ActivationsDataSource) buildFilters(terraformFilters []FilterModel) ([]ssmtypes.DescribeActivationsFilter, diag.Diagnostics) {
	var diagnostics diag.Diagnostics
	var filters []ssmtypes.DescribeActivationsFilter

	for _, tfFilter := range terraformFilters {
		// Valider le nom du filtre
		filterKey, err := d.validateFilterName(tfFilter.Name.ValueString())
		if err != nil {
			diagnostics.AddError(
				"Invalid filter name configuration",
				fmt.Sprintf("Filter name '%s' is invalid: %s. Please use a valid filter name.", tfFilter.Name.ValueString(), err.Error()),
			)
			continue
		}

		// Convertir les valeurs
		values := make([]string, len(tfFilter.Values))
		for i, value := range tfFilter.Values {
			values[i] = value.ValueString()
		}

		filters = append(filters, ssmtypes.DescribeActivationsFilter{
			FilterKey:    filterKey,
			FilterValues: values,
		})
	}

	return filters, diagnostics
}

// validateFilterName valide et convertit le nom du filtre Terraform en clé de filtre SSM.
func (d *ActivationsDataSource) validateFilterName(name string) (ssmtypes.DescribeActivationsFilterKeys, error) {
	switch name {
	case "activation-ids":
		return ssmtypes.DescribeActivationsFilterKeysActivationIds, nil
	case "iam-role":
		return ssmtypes.DescribeActivationsFilterKeysIamRole, nil
	case "default-instance-name":
		return ssmtypes.DescribeActivationsFilterKeysDefaultInstanceName, nil
	default:
		return "", fmt.Errorf("invalid filter name '%s'. Valid values are: 'activation-ids', 'iam-role', 'default-instance-name'", name)
	}
}

// convertActivations convertit les activations SSM en modèle Terraform.
func (d *ActivationsDataSource) convertActivations(activations []ssmtypes.Activation) []ActivationModel {
	if activations == nil {
		return []ActivationModel{}
	}
	
	result := make([]ActivationModel, len(activations))

	for i, activation := range activations {
		// Formater les dates
		var expirationDate, createdDate string
		if activation.ExpirationDate != nil {
			expirationDate = activation.ExpirationDate.Format(time.RFC3339)
		}
		if activation.CreatedDate != nil {
			createdDate = activation.CreatedDate.Format(time.RFC3339)
		}

		// Déterminer si l'activation a expiré
		expired := false
		if activation.ExpirationDate != nil {
			expired = time.Now().After(*activation.ExpirationDate)
		}

		result[i] = ActivationModel{
			ActivationId:       types.StringValue(aws.ToString(activation.ActivationId)),
			IamRole:            types.StringValue(aws.ToString(activation.IamRole)),
			RegistrationLimit:  types.Int64Value(int64(aws.ToInt32(activation.RegistrationLimit))),
			RegistrationsCount: types.Int64Value(int64(aws.ToInt32(activation.RegistrationsCount))),
			ExpirationDate:     types.StringValue(expirationDate),
			Expired:            types.BoolValue(expired),
			CreatedDate:        types.StringValue(createdDate),
		}
	}

	return result
}

// filterByExpired filtre les activations par statut d'expiration.
func (d *ActivationsDataSource) filterByExpired(activations []ActivationModel, expired bool) []ActivationModel {
	if activations == nil {
		return []ActivationModel{}
	}
	
	var result []ActivationModel

	for _, activation := range activations {
		if activation.Expired.ValueBool() == expired {
			result = append(result, activation)
		}
	}

	return result
}
