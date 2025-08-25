package internal

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	awsservice "github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/jd-ucpa/terraform-provider-test/internal/ssm"
	"github.com/jd-ucpa/terraform-provider-test/internal/sts"
)

// Ensure TestProvider satisfies various provider interfaces.
var _ provider.Provider = &TestProvider{}
var _ provider.ProviderWithFunctions = &TestProvider{}

// TestProvider est le provider Terraform principal qui gère l'authentification AWS
// et enregistre les ressources et data sources disponibles.
type TestProvider struct{}

// validRegions contient la liste des régions AWS valides supportées par ce provider.
// Cette liste est utilisée pour valider la région spécifiée dans la configuration.
var validRegions = []string{
	"us-east-1", // US East – Virginie du Nord
	"us-east-2", // US East – Ohio
	"us-west-1", // US West – Californie du Nord
	"us-west-2", // US West – Oregon
	"eu-west-1",  // Irlande
	"eu-west-2",  // Royaume‑Uni – Londres
	"eu-west-3",  // France – Paris
	"eu-central-1", // Allemagne – Francfort
	"eu-central-2", // Suisse – Zurich – opt‑in
	"eu-north-1",   // Suède – Stockholm
	"eu-south-1",   // Italie – Milan – opt‑in
	"eu-south-2",   // Espagne – Aragon – opt‑in
}

// TestProviderModel définit le modèle de configuration du provider.
// Il contient les paramètres globaux comme la région, le profil et la configuration assume_role.
type TestProviderModel struct {
	Region    types.String `tfsdk:"region"`
	Profile   types.String `tfsdk:"profile"`
	AssumeRole *AssumeRoleModel `tfsdk:"assume_role"`
}

// AssumeRoleModel définit la configuration pour l'assumption de rôle AWS.
// Il permet de spécifier un rôle IAM à assumer et un nom de session personnalisé.
type AssumeRoleModel struct {
	RoleArn     types.String `tfsdk:"role_arn"`
	SessionName types.String `tfsdk:"session_name"`
}

// Metadata définit le nom du provider utilisé dans les configurations Terraform.
// Ce nom est utilisé pour référencer ce provider dans les fichiers .tf.
func (p *TestProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "test"
	resp.Version = "0.0.2"
}

// Schema définit la structure et la documentation du provider.
// Cette méthode décrit les attributs de configuration disponibles et leur documentation
// qui sera affichée dans la documentation Terraform.
func (p *TestProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Test provider is used to interact with AWS services through SSM Send Command and STS Caller Identity. The provider needs to be configured with the proper credentials before it can be used.",
		Attributes: map[string]schema.Attribute{
			"region": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The region where operations will take place. This can also be set via the `AWS_REGION` environment variable.",
				Description:         "The region where operations will take place.",
			},
			"profile": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The AWS profile to use for authentication. This can also be set via the `AWS_PROFILE` environment variable.",
				Description:         "The AWS profile to use for authentication.",
			},
		},
		Blocks: map[string]schema.Block{
			"assume_role": schema.SingleNestedBlock{
				MarkdownDescription: "Configuration for assuming an IAM role.",
				Attributes: map[string]schema.Attribute{
					"role_arn": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The ARN of the role to assume. This can also be set via the `TF_VAR_assume_role_role_arn` environment variable.",
						Description:         "The ARN of the role to assume.",
					},
					"session_name": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "The session name to use when assuming the role. If not provided, defaults to `terraform-provider-test`.",
						Description:         "The session name to use when assuming the role.",
					},
				},
			},
		},
	}
}

// Configure initialise le provider avec la configuration fournie par l'utilisateur.
// Cette méthode valide les paramètres, configure l'authentification AWS et prépare
// les clients AWS pour les ressources et data sources.
func (p *TestProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config TestProviderModel
	
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Configuration AWS de base - utilise les credentials par défaut du système
	var cfg aws.Config
	var err error

	// Charger la configuration AWS par défaut (utilise les credentials AWS standard)
	cfg, err = awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		// Vérifier si l'erreur est liée à un profil inexistant
		if strings.Contains(err.Error(), "failed to get shared config profile") {
			resp.Diagnostics.AddError("failed to get shared config profile", err.Error())
		} else {
			resp.Diagnostics.AddError("AWS Configuration Error", fmt.Sprintf("Unable to load AWS config: %s", err))
		}
		return
	}

	// Appliquer la région si spécifiée
	if !config.Region.IsNull() {
		region := config.Region.ValueString()
		
		// Valider la région
		if !isValidRegion(region) {
			resp.Diagnostics.AddError(
				fmt.Sprintf("invalid AWS Region: %s", region),
				"",
			)
			return
		}
		
		cfg.Region = region
	}

	// Appliquer le profil si spécifié
	if !config.Profile.IsNull() {
		profile := config.Profile.ValueString()
		
		// Charger la configuration AWS avec le profil spécifié
		cfg, err = awsconfig.LoadDefaultConfig(ctx, awsconfig.WithSharedConfigProfile(profile))
		if err != nil {
			resp.Diagnostics.AddError("failed to get shared config profile: "+profile, "")
      // resp.Diagnostics.AddError("failed to get shared config profile", err.Error())
			return
		}
		
		// Réappliquer la région si elle était spécifiée (car LoadDefaultConfig peut la réinitialiser)
		if !config.Region.IsNull() {
			cfg.Region = config.Region.ValueString()
		}
	}

	// Configuration de l'assume role si spécifiée
	if config.AssumeRole != nil && !config.AssumeRole.RoleArn.IsNull() {
		roleArn := config.AssumeRole.RoleArn.ValueString()
		
		// Valider le format de l'ARN
		if !isValidRoleARN(roleArn) {
			resp.Diagnostics.AddError(
				fmt.Sprintf(`"assume_role.0.role_arn" (%s) is an invalid ARN: invalid account ID value (expecting to match regular expression: ^(aws|aws-managed|third-party|aws-marketplace|\\d{12}|cw.{10})$)`, roleArn),
				"",
			)
			return
		}
		
		stsClient := awsservice.NewFromConfig(cfg)
		
		// Préparer les options pour l'assume role
		assumeRoleOptions := []func(*stscreds.AssumeRoleOptions){
			func(options *stscreds.AssumeRoleOptions) {
				options.RoleARN = roleArn
			},
		}
		
		// Ajouter le session name (par défaut ou spécifié)
		sessionName := "terraform-provider-test"
		if !config.AssumeRole.SessionName.IsNull() {
			sessionName = config.AssumeRole.SessionName.ValueString()
		}
		assumeRoleOptions = append(assumeRoleOptions, func(options *stscreds.AssumeRoleOptions) {
			options.RoleSessionName = sessionName
		})
		

		
		cfg.Credentials = stscreds.NewAssumeRoleProvider(stsClient, config.AssumeRole.RoleArn.ValueString(), assumeRoleOptions...)
	}

	// Partager la configuration AWS avec les ressources
	resp.DataSourceData = cfg
	resp.ResourceData = cfg
}

// Resources enregistre toutes les ressources disponibles dans ce provider.
// Cette méthode retourne une liste de constructeurs de ressources qui seront
// disponibles dans les configurations Terraform.
func (p *TestProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		ssm.NewSendCommandResource,
	}
}

// DataSources enregistre tous les data sources disponibles dans ce provider.
// Cette méthode retourne une liste de constructeurs de data sources qui seront
// disponibles dans les configurations Terraform.
func (p *TestProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		sts.NewCallerIdentityDataSource,
	}
}

// Functions enregistre toutes les fonctions disponibles dans ce provider.
// Cette méthode retourne une liste de constructeurs de fonctions qui seront
// disponibles dans les configurations Terraform.
func (p *TestProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

// Provider crée et retourne une nouvelle instance du provider TestProvider.
// Cette fonction est utilisée par Terraform pour initialiser le provider.
func Provider() provider.Provider {
	return &TestProvider{}
}

// isValidRegion vérifie si une région AWS est valide en la comparant avec la liste
// des régions supportées. Cette fonction est utilisée pour valider la configuration
// du provider.
func isValidRegion(region string) bool {
	for _, validRegion := range validRegions {
		if strings.EqualFold(region, validRegion) {
			return true
		}
	}
	return false
}

// isValidRoleARN vérifie si un ARN de rôle IAM est valide en utilisant une expression
// régulière. Cette fonction valide le format arn:aws:iam::ACCOUNT_ID:role/ROLE_NAME
// et s'assure que l'ACCOUNT_ID contient exactement 12 chiffres.
func isValidRoleARN(roleArn string) bool {
	// Regex pour valider un ARN de rôle IAM
	// Format: arn:aws:iam::ACCOUNT_ID:role/ROLE_NAME
	arnRegex := regexp.MustCompile(`^arn:aws:iam::(\d{12}):role/[a-zA-Z0-9+=,.@_-]+$`)
	return arnRegex.MatchString(roleArn)
}
