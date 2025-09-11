package functions

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure JSONPrettyDataSource satisfies various datasource interfaces.
var _ datasource.DataSource = &JSONPrettyDataSource{}

// NewJSONPrettyDataSource crée et retourne une nouvelle instance du data source
// JSONPrettyDataSource. Cette fonction est utilisée par le provider pour enregistrer
// le data source dans Terraform.
func NewJSONPrettyDataSource() datasource.DataSource {
	return &JSONPrettyDataSource{}
}

// JSONPrettyDataSource formate une chaîne JSON avec indentation personnalisable
type JSONPrettyDataSource struct{}

// JSONPrettyDataSourceModel définit le modèle de données pour le data source
// JSONPretty. Il contient les attributs de configuration et le résultat.
type JSONPrettyDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	JSON    types.String `tfsdk:"json"`
	Indent  types.Int64  `tfsdk:"indent"`
	Newline types.Bool   `tfsdk:"newline"`
	Result  types.String `tfsdk:"result"`
}

// Metadata définit le nom du type de data source utilisé dans les configurations Terraform.
// Ce nom est utilisé pour référencer ce data source dans les fichiers .tf.
func (d *JSONPrettyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "test_json_pretty"
}

// Schema définit la structure et la documentation du data source.
func (d *JSONPrettyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Formats a JSON string with proper indentation and optional newline.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the data source.",
			},
			"json": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "The JSON string to format.",
			},
			"indent": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Number of spaces to use for indentation. Default is 2.",
			},
			"newline": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Whether to append a newline character at the end. Default is false.",
			},
			"result": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The formatted JSON string.",
			},
		},
	}
}

// Read lit les données du data source et les retourne.
func (d *JSONPrettyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data JSONPrettyDataSourceModel

	// Lire la configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Utiliser les valeurs par défaut si non fournies
	indentValue := int64(2)
	if !data.Indent.IsNull() {
		indentValue = data.Indent.ValueInt64()
		if indentValue < 0 {
			indentValue = 2 // fallback to default if negative
		}
	}

	newlineValue := false
	if !data.Newline.IsNull() {
		newlineValue = data.Newline.ValueBool()
	}

	// Valider et parser le JSON
	var parsed interface{}
	if err := json.Unmarshal([]byte(data.JSON.ValueString()), &parsed); err != nil {
		resp.Diagnostics.AddError(
			"Invalid JSON string",
			"Unable to parse JSON string: "+err.Error()+". Please verify the JSON syntax is correct.",
		)
		return
	}

	// Formater le JSON avec indentation
	indentStr := strings.Repeat(" ", int(indentValue))
	prettyBytes, err := json.MarshalIndent(parsed, "", indentStr)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to format JSON",
			"Unable to format JSON string: "+err.Error()+". Please verify the JSON content is valid.",
		)
		return
	}

	result := string(prettyBytes)
	if newlineValue {
		result += "\n"
	}

	// Générer un ID unique basé sur le contenu et les paramètres
	idInput := fmt.Sprintf("%s|%d|%t", data.JSON.ValueString(), indentValue, newlineValue)
	hash := sha1.Sum([]byte(idInput))
	id := hex.EncodeToString(hash[:])

	// Mettre à jour le modèle avec les résultats
	data.ID = types.StringValue(id)
	data.Result = types.StringValue(result)

	// Sauvegarder les données dans l'état
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
