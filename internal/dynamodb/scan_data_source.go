package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbtypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ScanDataSource satisfies various datasource interfaces.
var _ datasource.DataSource = &ScanDataSource{}

// NewScanDataSource crée et retourne une nouvelle instance du data source
// ScanDataSource. Cette fonction est utilisée par le provider pour enregistrer
// le data source dans Terraform.
func NewScanDataSource() datasource.DataSource {
	return &ScanDataSource{}
}

// ScanDataSource récupère les éléments d'une table DynamoDB via l'opération Scan.
type ScanDataSource struct {
	dynamodb *dynamodb.Client
}

// ScanDataSourceModel définit le modèle de données pour le data source
// Scan. Il contient les paramètres d'entrée et les attributs retournés.
type ScanDataSourceModel struct {
	TableName                 types.String            `tfsdk:"table_name"`
	IndexName                 types.String            `tfsdk:"index_name"`
	ProjectionExpression      types.String            `tfsdk:"projection_expression"`
	FilterExpression          types.String            `tfsdk:"filter_expression"`
	ExpressionAttributeNames  map[string]types.String `tfsdk:"expression_attribute_names"`
	ExpressionAttributeValues map[string]types.String `tfsdk:"expression_attribute_values"`
	ID                        types.String            `tfsdk:"id"`
	Items                     types.List              `tfsdk:"items"`
	ItemsCount                types.Int64             `tfsdk:"items_count"`
	ScannedCount              types.Int64             `tfsdk:"scanned_count"`
}

// Metadata définit le nom du type de data source utilisé dans les configurations Terraform.
func (d *ScanDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dynamodb_scan"
}

// Configure initialise le client DynamoDB à partir de la configuration du provider.
func (d *ScanDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Éviter le panic si le provider n'a pas été configuré
	if req.ProviderData == nil {
		return
	}

	// Vérifier que la configuration est du bon type
	config, ok := req.ProviderData.(aws.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Provider configuration error",
			fmt.Sprintf("Expected aws.Config for DynamoDB scan data source, got: %T. This indicates a provider configuration issue. Please verify your provider configuration and report this issue if it persists.", req.ProviderData),
		)
		return
	}
	
	// Créer le client DynamoDB à partir de la configuration AWS
	d.dynamodb = dynamodb.NewFromConfig(config)
}

// Schema définit la structure et la documentation du data source.
func (d *ScanDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to scan a DynamoDB table and retrieve items. This data source calls the AWS DynamoDB Scan API to retrieve items from a table based on specified criteria.",
		
		Attributes: map[string]schema.Attribute{
			"table_name": schema.StringAttribute{
				MarkdownDescription: "The name of the DynamoDB table to scan.",
				Required:            true,
			},
			"index_name": schema.StringAttribute{
				MarkdownDescription: "The name of the secondary index to scan. If not specified, the main table will be scanned.",
				Optional:            true,
			},
			"projection_expression": schema.StringAttribute{
				MarkdownDescription: "A string that identifies attributes to retrieve from the table. You can use attribute names without the # prefix, and the provider will automatically add them.",
				Optional:            true,
			},
			"filter_expression": schema.StringAttribute{
				MarkdownDescription: "A string that contains conditions that DynamoDB applies after the Scan operation, but before the data is returned to you. You can use attribute names without the # prefix, and the provider will automatically add them.",
				Optional:            true,
			},
			"expression_attribute_names": schema.MapAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "One or more substitution tokens for attribute names in an expression. You can specify attribute names without the # prefix, and the provider will automatically add them.",
				Optional:            true,
			},
			"expression_attribute_values": schema.MapAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "One or more values that can be substituted in an expression. Use the : (colon) character in an expression to dereference an attribute value.",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the data source.",
				Computed:            true,
			},
			"items": schema.ListAttribute{
				ElementType:         types.MapType{ElemType: types.StringType},
				MarkdownDescription: "List of items retrieved from the table.",
				Computed:            true,
			},
			"items_count": schema.Int64Attribute{
				MarkdownDescription: "The number of items in the response.",
				Computed:            true,
			},
			"scanned_count": schema.Int64Attribute{
				MarkdownDescription: "The number of items evaluated, before any ScanFilter is applied.",
				Computed:            true,
			},
		},
	}
}

// Read récupère les éléments DynamoDB en appelant l'API DynamoDB Scan.
func (d *ScanDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ScanDataSourceModel

	// Récupérer la configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Construire les paramètres pour l'API
	input := &dynamodb.ScanInput{
		TableName: aws.String(data.TableName.ValueString()),
	}

	// Ajouter l'index si spécifié
	if !data.IndexName.IsNull() && !data.IndexName.IsUnknown() {
		input.IndexName = aws.String(data.IndexName.ValueString())
	}

	// Traiter les expression attribute names
	expressionAttributeNames := make(map[string]string)
	if data.ExpressionAttributeNames != nil {
		for key, value := range data.ExpressionAttributeNames {
			// Ajouter automatiquement le préfixe # si pas présent
			if !strings.HasPrefix(key, "#") {
				expressionAttributeNames["#"+key] = value.ValueString()
			} else {
				expressionAttributeNames[key] = value.ValueString()
			}
		}
	}

	// Traiter la projection expression
	if !data.ProjectionExpression.IsNull() && !data.ProjectionExpression.IsUnknown() {
		projectionExpr := data.ProjectionExpression.ValueString()
		
		// Remplacer les alias sans # par des alias avec #
		for key := range data.ExpressionAttributeNames {
			if !strings.HasPrefix(key, "#") {
				projectionExpr = strings.ReplaceAll(projectionExpr, key, "#"+key)
			}
		}
		
		input.ProjectionExpression = aws.String(projectionExpr)
	}

	// Traiter le filter expression
	if !data.FilterExpression.IsNull() && !data.FilterExpression.IsUnknown() {
		filterExpr := data.FilterExpression.ValueString()
		
		// Remplacer les alias sans # par des alias avec #
		for key := range data.ExpressionAttributeNames {
			if !strings.HasPrefix(key, "#") {
				filterExpr = strings.ReplaceAll(filterExpr, key, "#"+key)
			}
		}
		
		// Remplacer les valeurs sans : par des valeurs avec :
		for key := range data.ExpressionAttributeValues {
			if !strings.HasPrefix(key, ":") {
				filterExpr = strings.ReplaceAll(filterExpr, key, ":"+key)
			}
		}
		
		input.FilterExpression = aws.String(filterExpr)
	}

	// Traiter les expression attribute values
	expressionAttributeValues := make(map[string]dynamodbtypes.AttributeValue)
	if data.ExpressionAttributeValues != nil {
		for key, value := range data.ExpressionAttributeValues {
			// Ajouter automatiquement le préfixe : si pas présent
			attrKey := key
			if !strings.HasPrefix(key, ":") {
				attrKey = ":" + key
			}
			
			// Créer l'AttributeValue pour DynamoDB
			expressionAttributeValues[attrKey] = &dynamodbtypes.AttributeValueMemberS{
				Value: value.ValueString(),
			}
		}
	}

	// Ajouter les expression attribute names si présents
	if len(expressionAttributeNames) > 0 {
		input.ExpressionAttributeNames = expressionAttributeNames
	}

	// Ajouter les expression attribute values si présents
	if len(expressionAttributeValues) > 0 {
		input.ExpressionAttributeValues = expressionAttributeValues
	}

	// Appeler l'API DynamoDB
	result, err := d.dynamodb.Scan(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to scan DynamoDB table",
			fmt.Sprintf("Error calling AWS DynamoDB Scan API for table '%s': %s. Please verify your AWS credentials, permissions, and that the DynamoDB table exists and is accessible.", data.TableName.ValueString(), err),
		)
		return
	}

	// Convertir les éléments DynamoDB en liste Terraform
	items := make([]attr.Value, len(result.Items))
	for i, item := range result.Items {
		// Convertir chaque item en map[string]string
		itemMap := make(map[string]attr.Value)
		for key, value := range item {
			// Convertir la valeur DynamoDB en string
			itemMap[key] = d.convertDynamoDBValueToString(value)
		}
		items[i] = types.MapValueMust(types.StringType, itemMap)
	}

	// Mettre à jour le modèle de données
	data.ID = types.StringValue(fmt.Sprintf("dynamodb_scan_%s", data.TableName.ValueString()))
	data.Items = types.ListValueMust(types.MapType{ElemType: types.StringType}, items)
	data.ItemsCount = types.Int64Value(int64(result.Count))
	data.ScannedCount = types.Int64Value(int64(result.ScannedCount))

	// Sauvegarder les données
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// convertDynamoDBValueToString convertit une valeur DynamoDB en string
func (d *ScanDataSource) convertDynamoDBValueToString(value interface{}) types.String {
	switch v := value.(type) {
	case *dynamodbtypes.AttributeValueMemberS:
		return types.StringValue(v.Value)
	case *dynamodbtypes.AttributeValueMemberN:
		return types.StringValue(v.Value)
	case *dynamodbtypes.AttributeValueMemberB:
		return types.StringValue(string(v.Value))
	case *dynamodbtypes.AttributeValueMemberBOOL:
		return types.StringValue(fmt.Sprintf("%t", v.Value))
	case *dynamodbtypes.AttributeValueMemberNULL:
		return types.StringValue("null")
	case *dynamodbtypes.AttributeValueMemberL:
		// Convertir la liste en JSON string
		jsonBytes, _ := json.Marshal(v.Value)
		return types.StringValue(string(jsonBytes))
	case *dynamodbtypes.AttributeValueMemberM:
		// Convertir la map en JSON string
		jsonBytes, _ := json.Marshal(v.Value)
		return types.StringValue(string(jsonBytes))
	default:
		return types.StringValue(fmt.Sprintf("%v", value))
	}
}
