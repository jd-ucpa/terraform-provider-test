package cloudwatch

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cwtypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure CloudWatchAlarmsDataSource satisfies various datasource interfaces.
var _ datasource.DataSource = &CloudWatchAlarmsDataSource{}

// stringvalidator.OneOf equivalent for alarm types
type alarmTypeValidator struct {
	validValues []string
	msg         string
}

func (v alarmTypeValidator) Description(ctx context.Context) string {
	return v.msg
}

func (v alarmTypeValidator) MarkdownDescription(ctx context.Context) string {
	return v.msg
}

func (v alarmTypeValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	for _, validValue := range v.validValues {
		if value == validValue {
			return
		}
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid alarm type",
		v.msg,
	)
}

// stringvalidatorOneOf creates a one-of validator
func stringvalidatorOneOf(validValues []string, msg string) validator.String {
	return alarmTypeValidator{
		validValues: validValues,
		msg:         msg,
	}
}

// NewCloudWatchAlarmsDataSource crée et retourne une nouvelle instance du data source
// CloudWatchAlarmsDataSource. Cette fonction est utilisée par le provider pour enregistrer
// le data source dans Terraform.
func NewCloudWatchAlarmsDataSource() datasource.DataSource {
	return &CloudWatchAlarmsDataSource{}
}

// CloudWatchAlarmsDataSource récupère les alarmes CloudWatch selon les critères spécifiés
// via l'API CloudWatch DescribeAlarms.
type CloudWatchAlarmsDataSource struct {
	cloudwatch *cloudwatch.Client
}

// CloudWatchAlarmsDataSourceModel définit le modèle de données pour le data source
// CloudWatchAlarms. Il contient les paramètres d'entrée et les attributs retournés.
type CloudWatchAlarmsDataSourceModel struct {
	AlarmType               types.String `tfsdk:"alarm_type"`
	AlarmNamePrefix         types.String `tfsdk:"alarm_name_prefix"`
	AlarmNames              types.List   `tfsdk:"alarm_names"`
	StateValue              types.String `tfsdk:"state_value"`
	IgnoreAutoscalingAlarms types.Bool   `tfsdk:"ignore_autoscaling_alarms"`
	ID                      types.String `tfsdk:"id"`
	Alarms                  []Alarm      `tfsdk:"alarms"`
}

// Alarm représente une alarme CloudWatch avec toutes ses propriétés
type Alarm struct {
	AlarmName                types.String   `tfsdk:"alarm_name"`
	AlarmArn                 types.String   `tfsdk:"alarm_arn"`
	ActionsEnabled           types.Bool     `tfsdk:"actions_enabled"`
	OKActions                []types.String `tfsdk:"ok_actions"`
	AlarmActions             []types.String `tfsdk:"alarm_actions"`
	InsufficientDataActions  []types.String `tfsdk:"insufficient_data_actions"`
	StateValue               types.String   `tfsdk:"state_value"`
	MetricName               types.String   `tfsdk:"metric_name"`
	Namespace                types.String   `tfsdk:"namespace"`
	Statistic                types.String   `tfsdk:"statistic"`
	Dimensions               []Dimension    `tfsdk:"dimensions"`
	Period                   types.Int64    `tfsdk:"period"`
	EvaluationPeriods        types.Int64    `tfsdk:"evaluation_periods"`
	DatapointsToAlarm        types.Int64    `tfsdk:"datapoints_to_alarm"`
	Threshold                types.Float64  `tfsdk:"threshold"`
	ComparisonOperator       types.String   `tfsdk:"comparison_operator"`
	TreatMissingData         types.String   `tfsdk:"treat_missing_data"`
}

// Dimension représente une dimension d'une métrique CloudWatch
type Dimension struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

// Metadata définit le nom du type de data source utilisé dans les configurations Terraform.
func (d *CloudWatchAlarmsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloudwatch_alarms"
}

// Configure initialise le client CloudWatch à partir de la configuration du provider.
func (d *CloudWatchAlarmsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Éviter le panic si le provider n'a pas été configuré
	if req.ProviderData == nil {
		return
	}

	// Vérifier que la configuration est du bon type
	config, ok := req.ProviderData.(aws.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Provider configuration error",
			fmt.Sprintf("Expected aws.Config for CloudWatch alarms data source, got: %T. This indicates a provider configuration issue. Please verify your provider configuration and report this issue if it persists.", req.ProviderData),
		)
		return
	}
	
	// Créer le client CloudWatch à partir de la configuration AWS
	d.cloudwatch = cloudwatch.NewFromConfig(config)
}

// Schema définit la structure et la documentation du data source.
func (d *CloudWatchAlarmsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to get information about CloudWatch alarms. This data source calls the AWS CloudWatch DescribeAlarms API to retrieve alarm information based on specified criteria.",
		
		Attributes: map[string]schema.Attribute{
			"alarm_type": schema.StringAttribute{
				MarkdownDescription: "The type of alarm to retrieve. Valid values are 'MetricAlarm' and 'CompositeAlarm'. Defaults to 'MetricAlarm'.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidatorOneOf(
						[]string{"MetricAlarm", "CompositeAlarm"},
						"alarm_type must be either 'MetricAlarm' or 'CompositeAlarm'",
					),
				},
			},
			"alarm_name_prefix": schema.StringAttribute{
				MarkdownDescription: "The prefix of the alarm name to filter by.",
				Optional:            true,
			},
			"alarm_names": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "The list of alarm names to retrieve.",
				Optional:            true,
			},
			"state_value": schema.StringAttribute{
				MarkdownDescription: "The state value to filter by. Valid values are 'OK', 'ALARM', and 'INSUFFICIENT_DATA'.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidatorOneOf(
						[]string{"OK", "ALARM", "INSUFFICIENT_DATA"},
						"state_value must be either 'OK', 'ALARM', or 'INSUFFICIENT_DATA'",
					),
				},
			},
			"ignore_autoscaling_alarms": schema.BoolAttribute{
				MarkdownDescription: "If true, filters out alarms whose names start with 'TargetTracking-'. These are typically Auto Scaling target tracking alarms. Defaults to true.",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the data source.",
				Computed:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"alarms": schema.ListNestedBlock{
				MarkdownDescription: "List of CloudWatch alarms matching the criteria.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"alarm_name": schema.StringAttribute{
							MarkdownDescription: "The name of the alarm.",
							Computed:            true,
						},
						"alarm_arn": schema.StringAttribute{
							MarkdownDescription: "The ARN of the alarm.",
							Computed:            true,
						},
						"actions_enabled": schema.BoolAttribute{
							MarkdownDescription: "Whether the actions for this alarm are enabled.",
							Computed:            true,
						},
						"ok_actions": schema.ListAttribute{
							ElementType:         types.StringType,
							MarkdownDescription: "The list of actions to execute when this alarm transitions to an OK state.",
							Computed:            true,
						},
						"alarm_actions": schema.ListAttribute{
							ElementType:         types.StringType,
							MarkdownDescription: "The list of actions to execute when this alarm transitions to an ALARM state.",
							Computed:            true,
						},
						"insufficient_data_actions": schema.ListAttribute{
							ElementType:         types.StringType,
							MarkdownDescription: "The list of actions to execute when this alarm transitions to an INSUFFICIENT_DATA state.",
							Computed:            true,
						},
						"state_value": schema.StringAttribute{
							MarkdownDescription: "The current state of the alarm.",
							Computed:            true,
						},
						"metric_name": schema.StringAttribute{
							MarkdownDescription: "The name of the metric associated with the alarm.",
							Computed:            true,
						},
						"namespace": schema.StringAttribute{
							MarkdownDescription: "The namespace of the metric associated with the alarm.",
							Computed:            true,
						},
						"statistic": schema.StringAttribute{
							MarkdownDescription: "The statistic for the metric associated with the alarm.",
							Computed:            true,
						},
						"dimensions": schema.ListNestedAttribute{
							MarkdownDescription: "The dimensions for the metric associated with the alarm.",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										MarkdownDescription: "The name of the dimension.",
										Computed:            true,
									},
									"value": schema.StringAttribute{
										MarkdownDescription: "The value of the dimension.",
										Computed:            true,
									},
								},
							},
						},
						"period": schema.Int64Attribute{
							MarkdownDescription: "The period in seconds over which the specified statistic is applied.",
							Computed:            true,
						},
						"evaluation_periods": schema.Int64Attribute{
							MarkdownDescription: "The number of periods over which data is compared to the specified threshold.",
							Computed:            true,
						},
						"datapoints_to_alarm": schema.Int64Attribute{
							MarkdownDescription: "The number of datapoints that must be breaching to trigger the alarm.",
							Computed:            true,
						},
						"threshold": schema.Float64Attribute{
							MarkdownDescription: "The value to compare with the specified statistic.",
							Computed:            true,
						},
						"comparison_operator": schema.StringAttribute{
							MarkdownDescription: "The arithmetic operation to use when comparing the specified statistic and threshold.",
							Computed:            true,
						},
						"treat_missing_data": schema.StringAttribute{
							MarkdownDescription: "Sets how this alarm is to handle missing data points.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// Read récupère les alarmes CloudWatch en appelant l'API CloudWatch DescribeAlarms.
func (d *CloudWatchAlarmsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CloudWatchAlarmsDataSourceModel

	// Récupérer la configuration
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Définir les valeurs par défaut
	if data.AlarmType.IsNull() || data.AlarmType.IsUnknown() {
		data.AlarmType = types.StringValue("MetricAlarm")
	}
	
	if data.IgnoreAutoscalingAlarms.IsNull() || data.IgnoreAutoscalingAlarms.IsUnknown() {
		data.IgnoreAutoscalingAlarms = types.BoolValue(true)
	}

	// Construire les paramètres pour l'API
	input := &cloudwatch.DescribeAlarmsInput{
		MaxRecords: aws.Int32(100),
	}

	// Ajouter le type d'alarme
	if data.AlarmType.ValueString() == "CompositeAlarm" {
		input.AlarmTypes = []cwtypes.AlarmType{cwtypes.AlarmTypeCompositeAlarm}
	} else {
		input.AlarmTypes = []cwtypes.AlarmType{cwtypes.AlarmTypeMetricAlarm}
	}

	// Ajouter le préfixe du nom d'alarme
	if !data.AlarmNamePrefix.IsNull() && !data.AlarmNamePrefix.IsUnknown() {
		input.AlarmNamePrefix = aws.String(data.AlarmNamePrefix.ValueString())
	}

	// Ajouter les noms d'alarmes spécifiques
	if !data.AlarmNames.IsNull() && !data.AlarmNames.IsUnknown() {
		var alarmNames []string
		resp.Diagnostics.Append(data.AlarmNames.ElementsAs(ctx, &alarmNames, false)...)
		if resp.Diagnostics.HasError() {
			resp.Diagnostics.AddError(
				"Invalid alarm names configuration",
				"Failed to parse alarm_names parameter. Please ensure all alarm names are valid strings.",
			)
			return
		}
		input.AlarmNames = alarmNames
	}

	// Ajouter la valeur d'état
	if !data.StateValue.IsNull() && !data.StateValue.IsUnknown() {
		input.StateValue = cwtypes.StateValue(data.StateValue.ValueString())
	}

	// Variables pour la pagination
	var allAlarms []Alarm
	var nextToken *string

	// Boucle de pagination
	for {
		// Ajouter le NextToken si disponible
		if nextToken != nil {
			input.NextToken = nextToken
		}

		// Appeler l'API CloudWatch
		result, err := d.cloudwatch.DescribeAlarms(ctx, input)
		if err != nil {
			resp.Diagnostics.AddError(
				"Unable to retrieve CloudWatch alarms",
				fmt.Sprintf("Error calling AWS CloudWatch DescribeAlarms API: %s. Please verify your AWS credentials, permissions, and that the CloudWatch service is available in your region.", err),
			)
			return
		}

		// Convertir les résultats selon le type d'alarme
		var alarms []Alarm
		if data.AlarmType.ValueString() == "CompositeAlarm" {
			alarms = d.convertCompositeAlarms(result.CompositeAlarms)
		} else {
			alarms = d.convertMetricAlarms(result.MetricAlarms)
		}

		// Ajouter les alarmes à la liste complète
		allAlarms = append(allAlarms, alarms...)

		// Vérifier s'il y a plus de pages
		nextToken = result.NextToken
		if nextToken == nil {
			break
		}
	}

	// Filtrer les alarmes Auto Scaling si nécessaire
	if data.IgnoreAutoscalingAlarms.ValueBool() {
		allAlarms = d.filterAutoscalingAlarms(allAlarms)
	}

	// Mettre à jour le modèle de données
	data.ID = types.StringValue("cloudwatch_alarms")
	data.Alarms = allAlarms

	// Sauvegarder les données
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// convertMetricAlarms convertit les alarmes métriques AWS en modèle Terraform
func (d *CloudWatchAlarmsDataSource) convertMetricAlarms(metricAlarms []cwtypes.MetricAlarm) []Alarm {
	alarms := make([]Alarm, len(metricAlarms))
	
	for i, alarm := range metricAlarms {
		// Convertir les actions
		okActions := make([]types.String, len(alarm.OKActions))
		for j, action := range alarm.OKActions {
			okActions[j] = types.StringValue(action)
		}

		alarmActions := make([]types.String, len(alarm.AlarmActions))
		for j, action := range alarm.AlarmActions {
			alarmActions[j] = types.StringValue(action)
		}

		insufficientDataActions := make([]types.String, len(alarm.InsufficientDataActions))
		for j, action := range alarm.InsufficientDataActions {
			insufficientDataActions[j] = types.StringValue(action)
		}

		// Convertir les dimensions
		dimensions := make([]Dimension, len(alarm.Dimensions))
		for j, dim := range alarm.Dimensions {
			dimensions[j] = Dimension{
				Name:  types.StringValue(aws.ToString(dim.Name)),
				Value: types.StringValue(aws.ToString(dim.Value)),
			}
		}

		alarms[i] = Alarm{
			AlarmName:               types.StringValue(aws.ToString(alarm.AlarmName)),
			AlarmArn:                types.StringValue(aws.ToString(alarm.AlarmArn)),
			ActionsEnabled:          types.BoolValue(aws.ToBool(alarm.ActionsEnabled)),
			OKActions:               okActions,
			AlarmActions:            alarmActions,
			InsufficientDataActions: insufficientDataActions,
			StateValue:              types.StringValue(string(alarm.StateValue)),
			MetricName:              types.StringValue(aws.ToString(alarm.MetricName)),
			Namespace:               types.StringValue(aws.ToString(alarm.Namespace)),
			Statistic:               types.StringValue(string(alarm.Statistic)),
			Dimensions:              dimensions,
			Period:                  types.Int64Value(int64(aws.ToInt32(alarm.Period))),
			EvaluationPeriods:       types.Int64Value(int64(aws.ToInt32(alarm.EvaluationPeriods))),
			DatapointsToAlarm:       types.Int64Value(int64(aws.ToInt32(alarm.DatapointsToAlarm))),
			Threshold:               types.Float64Value(aws.ToFloat64(alarm.Threshold)),
			ComparisonOperator:      types.StringValue(string(alarm.ComparisonOperator)),
			TreatMissingData:        types.StringValue(aws.ToString(alarm.TreatMissingData)),
		}
	}

	return alarms
}

// filterAutoscalingAlarms filtre les alarmes dont le nom commence par 'TargetTracking-'
func (d *CloudWatchAlarmsDataSource) filterAutoscalingAlarms(alarms []Alarm) []Alarm {
	var filteredAlarms []Alarm
	
	for _, alarm := range alarms {
		alarmName := alarm.AlarmName.ValueString()
		if !strings.HasPrefix(alarmName, "TargetTracking-") {
			filteredAlarms = append(filteredAlarms, alarm)
		}
	}
	
	return filteredAlarms
}

// convertCompositeAlarms convertit les alarmes composites AWS en modèle Terraform
func (d *CloudWatchAlarmsDataSource) convertCompositeAlarms(compositeAlarms []cwtypes.CompositeAlarm) []Alarm {
	alarms := make([]Alarm, len(compositeAlarms))
	
	for i, alarm := range compositeAlarms {
		// Convertir les actions
		okActions := make([]types.String, len(alarm.OKActions))
		for j, action := range alarm.OKActions {
			okActions[j] = types.StringValue(action)
		}

		alarmActions := make([]types.String, len(alarm.AlarmActions))
		for j, action := range alarm.AlarmActions {
			alarmActions[j] = types.StringValue(action)
		}

		insufficientDataActions := make([]types.String, len(alarm.InsufficientDataActions))
		for j, action := range alarm.InsufficientDataActions {
			insufficientDataActions[j] = types.StringValue(action)
		}

		alarms[i] = Alarm{
			AlarmName:               types.StringValue(aws.ToString(alarm.AlarmName)),
			AlarmArn:                types.StringValue(aws.ToString(alarm.AlarmArn)),
			ActionsEnabled:          types.BoolValue(aws.ToBool(alarm.ActionsEnabled)),
			OKActions:               okActions,
			AlarmActions:            alarmActions,
			InsufficientDataActions: insufficientDataActions,
			StateValue:              types.StringValue(string(alarm.StateValue)),
			// Les alarmes composites n'ont pas de métriques associées
			MetricName:         types.StringNull(),
			Namespace:          types.StringNull(),
			Statistic:          types.StringNull(),
			Dimensions:         []Dimension{},
			Period:             types.Int64Null(),
			EvaluationPeriods:  types.Int64Null(),
			DatapointsToAlarm:  types.Int64Null(),
			Threshold:          types.Float64Null(),
			ComparisonOperator: types.StringNull(),
			TreatMissingData:   types.StringNull(),
		}
	}

	return alarms
}
