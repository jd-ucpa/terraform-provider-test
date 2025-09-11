package codebuild

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/codebuild"
	codebuildtypes "github.com/aws/aws-sdk-go-v2/service/codebuild/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &StartBuildResource{}

// NewStartBuildResource crée et retourne une nouvelle instance de la ressource
// StartBuildResource. Cette fonction est utilisée par le provider pour enregistrer
// la ressource dans Terraform.
func NewStartBuildResource() resource.Resource {
	return &StartBuildResource{}
}

// StartBuildResource gère le démarrage de builds AWS CodeBuild.
// Cette ressource permet de démarrer un build CodeBuild et de surveiller son statut.
type StartBuildResource struct {
	codebuild *codebuild.Client
	ssm       *ssm.Client
	secrets   *secretsmanager.Client
}

// EnvironmentVariableResourceModel définit le modèle pour les variables d'environnement.
type EnvironmentVariableResourceModel struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
	Type  types.String `tfsdk:"type"`
}

// StartBuildResourceModel définit le modèle de données pour la ressource StartBuild.
// Il contient tous les attributs de configuration et les données retournées par l'API CodeBuild.
type StartBuildResourceModel struct {
	Id                   types.String                        `tfsdk:"id"`
	ProjectName          types.String                        `tfsdk:"project_name"`
	EnvironmentVariables []EnvironmentVariableResourceModel  `tfsdk:"environment_variables"`
	Triggers             types.Map                           `tfsdk:"triggers"`
	// Propriétés retournées directement (sans imbrication dans "build")
	BuildId              types.String                        `tfsdk:"build_id"`
	BuildArn             types.String                        `tfsdk:"build_arn"`
	BuildNumber          types.Int64                         `tfsdk:"build_number"`
	BuildProjectName     types.String                        `tfsdk:"build_project_name"`
	BuildImage           types.String                        `tfsdk:"build_image"`
	BuildEnvironmentVariables types.List `tfsdk:"build_environment_variables"`
}

// Metadata définit le nom du type de ressource utilisé dans les configurations Terraform.
// Ce nom est utilisé pour référencer cette ressource dans les fichiers .tf.
func (r *StartBuildResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "test_codebuild_start_build"
}

// Configure initialise le client CodeBuild à partir de la configuration du provider.
// Cette méthode est appelée par Terraform pour configurer la ressource avec
// les paramètres d'authentification AWS (région, credentials, etc.).
func (r *StartBuildResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Éviter le panic si le provider n'a pas été configuré
	if req.ProviderData == nil {
		return
	}

	// Vérifier que la configuration est du bon type
	config, ok := req.ProviderData.(aws.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Provider configuration error",
			fmt.Sprintf("Expected aws.Config for CodeBuild start build resource, got: %T. This indicates a provider configuration issue. Please verify your provider configuration and report this issue if it persists.", req.ProviderData),
		)
		return
	}
	
	// Créer les clients AWS à partir de la configuration
	r.codebuild = codebuild.NewFromConfig(config)
	r.ssm = ssm.NewFromConfig(config)
	r.secrets = secretsmanager.NewFromConfig(config)
}

// Schema définit la structure et la documentation de la ressource.
// Cette méthode décrit les attributs disponibles, leurs types, et leur documentation Markdown
// qui sera affichée dans la documentation Terraform.
func (r *StartBuildResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The `test_codebuild_start_build` resource allows you to start a build using AWS CodeBuild. This resource supports configuring environment variables.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: "Identifier",
			},
			"project_name": schema.StringAttribute{
				MarkdownDescription: "The name of the CodeBuild project to start a build for.",
				Required:            true,
			},
			"build_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the build.",
			},
			"build_arn": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ARN of the build.",
			},
			"build_number": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "The build number.",
			},
			"build_project_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The name of the project.",
			},
			"build_image": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The image used for the build environment.",
			},
			"build_environment_variables": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "The environment variables for the build.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the environment variable.",
						},
						"value": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The value of the environment variable.",
						},
						"type": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The type of the environment variable (PLAINTEXT, PARAMETER_STORE, or SECRETS_MANAGER).",
						},
					},
				},
			},
			"triggers": schema.MapAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "A map of arbitrary strings that, when changed, will force the resource to be recreated.",
				Optional:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"environment_variables": schema.ListNestedBlock{
				MarkdownDescription: "The environment variables to pass to the build.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the environment variable.",
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "The value of the environment variable.",
							Required:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "The type of the environment variable. Valid values are PLAINTEXT, PARAMETER_STORE, or SECRETS_MANAGER. Defaults to PLAINTEXT.",
							Optional:            true,
						},
					},
				},
			},
		},
	}
}

// Create démarre un nouveau build CodeBuild.
// Cette méthode est appelée par Terraform lors de la création d'une ressource.
// Elle valide la configuration, démarre le build CodeBuild et retourne les informations du build.
func (r *StartBuildResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data StartBuildResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Valider l'existence des paramètres PARAMETER_STORE et SECRETS_MANAGER
	validationDiag := r.validateEnvironmentVariables(ctx, data)
	if validationDiag.HasError() {
		resp.Diagnostics.Append(validationDiag...)
		return
	}

	// Construire les variables d'environnement
	environmentVariables, diag := r.buildEnvironmentVariables(ctx, data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	// Construire l'input pour StartBuild
	input := &codebuild.StartBuildInput{
		ProjectName: aws.String(data.ProjectName.ValueString()),
	}

	// Ajouter les variables d'environnement si spécifiées
	if len(environmentVariables) > 0 {
		input.EnvironmentVariablesOverride = environmentVariables
	}

	// Démarrer le build
	output, err := r.codebuild.StartBuild(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to start CodeBuild project",
			fmt.Sprintf("Error calling AWS CodeBuild StartBuild API for project '%s': %s. Please verify your AWS credentials, permissions, and that the CodeBuild project exists and is accessible.", data.ProjectName.ValueString(), err),
		)
		return
	}

	// Mapper les données de retour
	data.Id = types.StringValue(*output.Build.Id)
	r.mapBuildToModel(ctx, output.Build, &data)

	// Normaliser les valeurs optionnelles
	r.normalizeOptionalValues(&data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read récupère l'état actuel de la ressource depuis l'état Terraform.
// Cette méthode est appelée par Terraform pour synchroniser l'état local avec l'état distant.
// Pour cette ressource, l'état est conservé tel quel car les builds CodeBuild sont statiques.
func (r *StartBuildResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StartBuildResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Dans une implémentation réelle, vous pourriez récupérer les détails du build
	// Pour l'instant, on garde l'état actuel
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update gère les modifications de la ressource.
// Cette méthode est appelée par Terraform lors de la modification d'une ressource existante.
// Elle vérifie si les triggers ont changé et exécute un nouveau build si nécessaire.
func (r *StartBuildResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data StartBuildResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Récupérer l'état actuel pour comparer les triggers
	var currentData StartBuildResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &currentData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Vérifier si les triggers ont changé
	triggersChanged := !data.Triggers.Equal(currentData.Triggers)

	if triggersChanged {

	// Valider l'existence des paramètres PARAMETER_STORE et SECRETS_MANAGER
	validationDiag := r.validateEnvironmentVariables(ctx, data)
	if validationDiag.HasError() {
		resp.Diagnostics.Append(validationDiag...)
		return
	}

	// Construire les variables d'environnement
	environmentVariables, diag := r.buildEnvironmentVariables(ctx, data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	// Construire l'input pour StartBuild
	input := &codebuild.StartBuildInput{
		ProjectName: aws.String(data.ProjectName.ValueString()),
	}

	// Ajouter les variables d'environnement si spécifiées
	if len(environmentVariables) > 0 {
		input.EnvironmentVariablesOverride = environmentVariables
	}

	// Démarrer le build
	output, err := r.codebuild.StartBuild(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to start CodeBuild project",
			fmt.Sprintf("Error calling AWS CodeBuild StartBuild API for project '%s': %s. Please verify your AWS credentials, permissions, and that the CodeBuild project exists and is accessible.", data.ProjectName.ValueString(), err),
		)
		return
	}

	// Mapper les données de retour
	data.Id = types.StringValue(*output.Build.Id)
	r.mapBuildToModel(ctx, output.Build, &data)

	// Normaliser les valeurs optionnelles
	r.normalizeOptionalValues(&data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	} else {
		// Si les triggers n'ont pas changé, préserver les valeurs calculées
		if data.Id.IsUnknown() || data.Id.IsNull() {
			data.Id = currentData.Id
		}
		if data.BuildId.IsUnknown() || data.BuildId.IsNull() {
			data.BuildId = currentData.BuildId
		}
		if data.BuildArn.IsUnknown() || data.BuildArn.IsNull() {
			data.BuildArn = currentData.BuildArn
		}
		if data.BuildNumber.IsUnknown() || data.BuildNumber.IsNull() {
			data.BuildNumber = currentData.BuildNumber
		}
		if data.BuildProjectName.IsUnknown() || data.BuildProjectName.IsNull() {
			data.BuildProjectName = currentData.BuildProjectName
		}
		if data.BuildImage.IsUnknown() || data.BuildImage.IsNull() {
			data.BuildImage = currentData.BuildImage
		}
		if data.BuildEnvironmentVariables.IsUnknown() || data.BuildEnvironmentVariables.IsNull() {
			data.BuildEnvironmentVariables = currentData.BuildEnvironmentVariables
		}

		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	}
}

// Delete gère la suppression de la ressource.
// Cette méthode est appelée par Terraform lors de la suppression d'une ressource.
// Pour cette ressource, on ne fait rien car les builds CodeBuild ne peuvent pas être supprimés.
func (r *StartBuildResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Les builds CodeBuild ne peuvent pas être supprimés, on ne fait rien
}

// buildEnvironmentVariables construit les variables d'environnement pour l'API CodeBuild
func (r *StartBuildResource) buildEnvironmentVariables(ctx context.Context, data StartBuildResourceModel) ([]codebuildtypes.EnvironmentVariable, diag.Diagnostics) {
	var diagnostics diag.Diagnostics
	
	environmentVariables := make([]codebuildtypes.EnvironmentVariable, len(data.EnvironmentVariables))
	
	for i, envVar := range data.EnvironmentVariables {
		envVarType := codebuildtypes.EnvironmentVariableTypePlaintext
		if !envVar.Type.IsNull() {
			switch envVar.Type.ValueString() {
			case "PARAMETER_STORE":
				envVarType = codebuildtypes.EnvironmentVariableTypeParameterStore
			case "SECRETS_MANAGER":
				envVarType = codebuildtypes.EnvironmentVariableTypeSecretsManager
			case "PLAINTEXT":
				envVarType = codebuildtypes.EnvironmentVariableTypePlaintext
			default:
				diagnostics.AddError(
					"Invalid environment variable type",
					fmt.Sprintf("Environment variable '%s' has invalid type '%s'. Valid values are PLAINTEXT, PARAMETER_STORE, or SECRETS_MANAGER.", envVar.Name.ValueString(), envVar.Type.ValueString()),
				)
				return nil, diagnostics
			}
		}
		
		environmentVariables[i] = codebuildtypes.EnvironmentVariable{
			Name:  aws.String(envVar.Name.ValueString()),
			Value: aws.String(envVar.Value.ValueString()),
			Type:  envVarType,
		}
	}
	
	return environmentVariables, diagnostics
}

// mapBuildToModel mappe les données de build de l'API vers le modèle Terraform
func (r *StartBuildResource) mapBuildToModel(ctx context.Context, build *codebuildtypes.Build, data *StartBuildResourceModel) {
	if build == nil {
		return
	}

	// Mapper les propriétés de base
	data.BuildId = types.StringValue(*build.Id)
	data.BuildArn = types.StringValue(*build.Arn)
	data.BuildNumber = types.Int64Value(int64(*build.BuildNumber))
	data.BuildProjectName = types.StringValue(*build.ProjectName)

	// Mapper l'environnement
	if build.Environment != nil {
		if build.Environment.Image != nil {
			data.BuildImage = types.StringValue(*build.Environment.Image)
		} else {
			data.BuildImage = types.StringNull()
		}
		
		// Mapper les variables d'environnement
		if len(build.Environment.EnvironmentVariables) > 0 {
			envVars := make([]EnvironmentVariableResourceModel, len(build.Environment.EnvironmentVariables))
			for i, envVar := range build.Environment.EnvironmentVariables {
				envVars[i] = EnvironmentVariableResourceModel{
					Name:  types.StringValue(*envVar.Name),
					Value: types.StringValue(*envVar.Value),
					Type:  types.StringValue(string(envVar.Type)),
				}
			}
			// Convertir en types.List
			envVarsList, diag := types.ListValueFrom(ctx, types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"name":  types.StringType,
					"value": types.StringType,
					"type":  types.StringType,
				},
			}, envVars)
			if !diag.HasError() {
				data.BuildEnvironmentVariables = envVarsList
			} else {
				data.BuildEnvironmentVariables = types.ListNull(types.ObjectType{
					AttrTypes: map[string]attr.Type{
						"name":  types.StringType,
						"value": types.StringType,
						"type":  types.StringType,
					},
				})
			}
		} else {
			data.BuildEnvironmentVariables = types.ListNull(types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"name":  types.StringType,
					"value": types.StringType,
					"type":  types.StringType,
				},
			})
		}
	} else {
		data.BuildImage = types.StringNull()
		data.BuildEnvironmentVariables = types.ListNull(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"name":  types.StringType,
				"value": types.StringType,
				"type":  types.StringType,
			},
		})
	}

}

// validateEnvironmentVariables valide l'existence des paramètres PARAMETER_STORE et SECRETS_MANAGER
func (r *StartBuildResource) validateEnvironmentVariables(ctx context.Context, data StartBuildResourceModel) diag.Diagnostics {
	var diagnostics diag.Diagnostics
	
	for _, envVar := range data.EnvironmentVariables {
		if envVar.Type.IsNull() || envVar.Type.ValueString() == "PLAINTEXT" {
			continue
		}
		
		switch envVar.Type.ValueString() {
		case "PARAMETER_STORE":
			// Vérifier que le paramètre SSM existe
			_, err := r.ssm.GetParameter(ctx, &ssm.GetParameterInput{
				Name:           aws.String(envVar.Value.ValueString()),
				WithDecryption: aws.Bool(false), // Pas besoin de décrypter pour vérifier l'existence
			})
			if err != nil {
				diagnostics.AddError(
					"Parameter Store validation failed",
					fmt.Sprintf("Parameter Store parameter '%s' does not exist or is not accessible: %v. Please verify the parameter exists and you have the necessary permissions to access it.", 
						envVar.Value.ValueString(), err),
				)
			}
			
		case "SECRETS_MANAGER":
			// Vérifier que le secret existe
			_, err := r.secrets.DescribeSecret(ctx, &secretsmanager.DescribeSecretInput{
				SecretId: aws.String(envVar.Value.ValueString()),
			})
			if err != nil {
				diagnostics.AddError(
					"Secrets Manager validation failed",
					fmt.Sprintf("Secrets Manager secret '%s' does not exist or is not accessible: %v. Please verify the secret exists and you have the necessary permissions to access it.",
						envVar.Value.ValueString(), err),
				)
			}
		}
	}
	
	return diagnostics
}

// normalizeOptionalValues s'assure que les valeurs optionnelles d'input sont définies
func (r *StartBuildResource) normalizeOptionalValues(data *StartBuildResourceModel) {
	// Ne pas normaliser les valeurs d'input car elles sont aussi des attributs de sortie
	// et doivent rester null si non définies pour éviter les erreurs d'incohérence
	
	// Normaliser les triggers
	if data.Triggers.IsNull() {
		data.Triggers = types.MapNull(types.StringType)
	}
}
