package functions

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure TimestampDataSource satisfies various datasource interfaces.
var _ datasource.DataSource = &TimestampDataSource{}

// NewTimestampDataSource crée et retourne une nouvelle instance du data source
// TimestampDataSource. Cette fonction est utilisée par le provider pour enregistrer
// le data source dans Terraform.
func NewTimestampDataSource() datasource.DataSource {
	return &TimestampDataSource{}
}

// TimestampDataSource retourne la date et l'heure actuelle avec support des fuseaux horaires
// et de l'ajout de temps. Par défaut, il retourne le timestamp UTC au format yyyy-mm-ddThh:mm:ss.
type TimestampDataSource struct{}

// TimestampDataSourceModel définit le modèle de données pour le data source
// Timestamp. Il contient les attributs de configuration et le résultat.
type TimestampDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	Result   types.String `tfsdk:"result"`
	TimeZone types.String `tfsdk:"time_zone"`
	TimeAdd  *TimeAddModel `tfsdk:"time_add"`
}

// TimeAddModel définit le bloc de configuration pour l'ajout de temps.
type TimeAddModel struct {
	Days    types.Int64 `tfsdk:"days"`
	Hours   types.Int64 `tfsdk:"hours"`
	Minutes types.Int64 `tfsdk:"minutes"`
	Seconds types.Int64 `tfsdk:"seconds"`
}

// Metadata définit le nom du type de data source utilisé dans les configurations Terraform.
// Ce nom est utilisé pour référencer ce data source dans les fichiers .tf.
func (d *TimestampDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = "test_timestamp"
}

// Schema définit la structure et la documentation du data source.
// Cette méthode décrit les attributs disponibles et leur documentation Markdown
// qui sera affichée dans la documentation Terraform.
func (d *TimestampDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to get the current timestamp with optional timezone and time addition. Returns the timestamp in the format yyyy-mm-ddThh:mm:ss (without the 'Z' suffix).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the timestamp data source.",
				Computed:            true,
			},
			"result": schema.StringAttribute{
				MarkdownDescription: "The current timestamp in the format yyyy-mm-ddThh:mm:ss.",
				Computed:            true,
			},
			"time_zone": schema.StringAttribute{
				MarkdownDescription: "The timezone to use for the timestamp. If not specified, defaults to UTC. Example: 'Europe/Paris', 'America/New_York'.",
				Optional:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"time_add": schema.SingleNestedBlock{
				MarkdownDescription: "Configuration for adding or subtracting time from the current timestamp.",
				Attributes: map[string]schema.Attribute{
					"days": schema.Int64Attribute{
						MarkdownDescription: "Number of days to add or subtract. Use negative values to subtract. Defaults to 0.",
						Optional:            true,
					},
					"hours": schema.Int64Attribute{
						MarkdownDescription: "Number of hours to add or subtract. Use negative values to subtract. Defaults to 0.",
						Optional:            true,
					},
					"minutes": schema.Int64Attribute{
						MarkdownDescription: "Number of minutes to add or subtract. Use negative values to subtract. Defaults to 0.",
						Optional:            true,
					},
					"seconds": schema.Int64Attribute{
						MarkdownDescription: "Number of seconds to add or subtract. Use negative values to subtract. Defaults to 0.",
						Optional:            true,
					},
				},
			},
		},
	}
}

// Read génère le timestamp actuel selon la configuration spécifiée.
// Cette méthode est appelée par Terraform pour obtenir les données du data source.
func (d *TimestampDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TimestampDataSourceModel

	// Parser la configuration du data source
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Commencer avec l'heure actuelle en UTC
	now := time.Now().UTC()

	// Appliquer le fuseau horaire si spécifié
	if !data.TimeZone.IsNull() && !data.TimeZone.IsUnknown() {
		timezoneStr := data.TimeZone.ValueString()
		if timezoneStr != "" && timezoneStr != "UTC" {
			loc, err := time.LoadLocation(timezoneStr)
			if err != nil {
			resp.Diagnostics.AddError(
				"Invalid timezone configuration",
				fmt.Sprintf("Unable to load timezone '%s': %s. Please verify the timezone string is valid (e.g., 'UTC', 'Europe/Paris', 'America/New_York').", timezoneStr, err.Error()),
			)
				return
			}
			now = now.In(loc)
		}
	}

	// Appliquer l'ajout de temps si configuré
	if data.TimeAdd != nil {
		// Récupérer les valeurs avec des valeurs par défaut
		days := int64(0)
		hours := int64(0)
		minutes := int64(0)
		seconds := int64(0)

		if !data.TimeAdd.Days.IsNull() && !data.TimeAdd.Days.IsUnknown() {
			days = data.TimeAdd.Days.ValueInt64()
		}
		if !data.TimeAdd.Hours.IsNull() && !data.TimeAdd.Hours.IsUnknown() {
			hours = data.TimeAdd.Hours.ValueInt64()
		}
		if !data.TimeAdd.Minutes.IsNull() && !data.TimeAdd.Minutes.IsUnknown() {
			minutes = data.TimeAdd.Minutes.ValueInt64()
		}
		if !data.TimeAdd.Seconds.IsNull() && !data.TimeAdd.Seconds.IsUnknown() {
			seconds = data.TimeAdd.Seconds.ValueInt64()
		}

		// Calculer la durée totale
		duration := time.Duration(days)*24*time.Hour +
			time.Duration(hours)*time.Hour +
			time.Duration(minutes)*time.Minute +
			time.Duration(seconds)*time.Second

		// Appliquer la durée
		now = now.Add(duration)
	}

	// Formater le timestamp au format yyyy-mm-ddThh:mm:ss (sans le 'Z')
	timestamp := now.Format("2006-01-02T15:04:05")

	// Générer un ID unique basé sur les paramètres
	var idInput string
	if data.TimeAdd != nil {
		days := int64(0)
		hours := int64(0)
		minutes := int64(0)
		seconds := int64(0)
		
		if !data.TimeAdd.Days.IsNull() && !data.TimeAdd.Days.IsUnknown() {
			days = data.TimeAdd.Days.ValueInt64()
		}
		if !data.TimeAdd.Hours.IsNull() && !data.TimeAdd.Hours.IsUnknown() {
			hours = data.TimeAdd.Hours.ValueInt64()
		}
		if !data.TimeAdd.Minutes.IsNull() && !data.TimeAdd.Minutes.IsUnknown() {
			minutes = data.TimeAdd.Minutes.ValueInt64()
		}
		if !data.TimeAdd.Seconds.IsNull() && !data.TimeAdd.Seconds.IsUnknown() {
			seconds = data.TimeAdd.Seconds.ValueInt64()
		}
		
		idInput = fmt.Sprintf("%s|%d|%d|%d|%d", data.TimeZone.ValueString(), days, hours, minutes, seconds)
	} else {
		idInput = data.TimeZone.ValueString()
	}
	
	hash := sha1.Sum([]byte(idInput))
	id := hex.EncodeToString(hash[:])

	// Mettre à jour le modèle avec les résultats
	data.ID = types.StringValue(id)
	data.Result = types.StringValue(timestamp)

	// Gérer les valeurs nulles de manière cohérente
	if data.TimeZone.IsNull() || data.TimeZone.IsUnknown() {
		data.TimeZone = types.StringNull()
	}
	
	if data.TimeAdd == nil {
		data.TimeAdd = &TimeAddModel{
			Days:    types.Int64Null(),
			Hours:   types.Int64Null(),
			Minutes: types.Int64Null(),
			Seconds: types.Int64Null(),
		}
	} else {
		// S'assurer que tous les champs sont définis (même à null)
		if data.TimeAdd.Days.IsNull() || data.TimeAdd.Days.IsUnknown() {
			data.TimeAdd.Days = types.Int64Null()
		}
		if data.TimeAdd.Hours.IsNull() || data.TimeAdd.Hours.IsUnknown() {
			data.TimeAdd.Hours = types.Int64Null()
		}
		if data.TimeAdd.Minutes.IsNull() || data.TimeAdd.Minutes.IsUnknown() {
			data.TimeAdd.Minutes = types.Int64Null()
		}
		if data.TimeAdd.Seconds.IsNull() || data.TimeAdd.Seconds.IsUnknown() {
			data.TimeAdd.Seconds = types.Int64Null()
		}
	}

	// Sauvegarder les données dans l'état Terraform
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
