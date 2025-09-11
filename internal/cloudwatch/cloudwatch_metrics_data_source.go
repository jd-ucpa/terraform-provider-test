package cloudwatch

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &CloudWatchMetricsDataSource{}

func NewCloudWatchMetricsDataSource() datasource.DataSource {
	return &CloudWatchMetricsDataSource{}
}

// CloudWatchMetricsDataSource defines the data source implementation.
type CloudWatchMetricsDataSource struct {
	cloudwatch *cloudwatch.Client
}

// CloudWatchMetricsDataSourceModel describes the data source data model.
type CloudWatchMetricsDataSourceModel struct {
	Namespace   basetypes.StringValue `tfsdk:"namespace"`
	MetricName  basetypes.StringValue `tfsdk:"metric_name"`
	Dimensions  []DimensionModel      `tfsdk:"dimension"`
	Metrics     []MetricModel         `tfsdk:"metrics"`
	Id          basetypes.StringValue `tfsdk:"id"`
}

// DimensionModel describes the dimension data model.
type DimensionModel struct {
	Name  basetypes.StringValue `tfsdk:"name"`
	Value basetypes.StringValue `tfsdk:"value"`
}

// MetricModel describes the metric data model.
type MetricModel struct {
	Namespace   basetypes.StringValue `tfsdk:"namespace"`
	MetricName  basetypes.StringValue `tfsdk:"metric_name"`
	Dimensions  []DimensionModel      `tfsdk:"dimensions"`
}

func (d *CloudWatchMetricsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cloudwatch_metrics"
}

func (d *CloudWatchMetricsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Use this data source to get information about CloudWatch metrics. This data source calls the AWS CloudWatch ListMetrics API to retrieve metric information based on specified criteria.",

		Attributes: map[string]schema.Attribute{
			"namespace": schema.StringAttribute{
				MarkdownDescription: "The namespace of the metric to filter by.",
				Optional:            true,
			},
			"metric_name": schema.StringAttribute{
				MarkdownDescription: "The name of the metric to filter by.",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The ID of the data source.",
				Computed:            true,
			},
			"metrics": schema.ListNestedAttribute{
				MarkdownDescription: "List of CloudWatch metrics matching the criteria.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"namespace": schema.StringAttribute{
							MarkdownDescription: "The namespace of the metric.",
							Computed:            true,
						},
						"metric_name": schema.StringAttribute{
							MarkdownDescription: "The name of the metric.",
							Computed:            true,
						},
						"dimensions": schema.ListNestedAttribute{
							MarkdownDescription: "The dimensions of the metric.",
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
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"dimension": schema.ListNestedBlock{
				MarkdownDescription: "The dimensions to filter by.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the dimension.",
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "The value of the dimension.",
							Required:            true,
						},
					},
				},
			},
		},
	}
}

func (d *CloudWatchMetricsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	// Vérifier que la configuration est du bon type
	config, ok := req.ProviderData.(aws.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Provider configuration error",
			fmt.Sprintf("Expected aws.Config for CloudWatch metrics data source, got: %T. This indicates a provider configuration issue. Please verify your provider configuration and report this issue if it persists.", req.ProviderData),
		)
		return
	}
	
	// Créer le client CloudWatch à partir de la configuration AWS
	d.cloudwatch = cloudwatch.NewFromConfig(config)
}

func (d *CloudWatchMetricsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CloudWatchMetricsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Build the input parameters for ListMetrics
	input := &cloudwatch.ListMetricsInput{}

	if !data.Namespace.IsNull() && !data.Namespace.IsUnknown() {
		input.Namespace = aws.String(data.Namespace.ValueString())
	}

	if !data.MetricName.IsNull() && !data.MetricName.IsUnknown() {
		input.MetricName = aws.String(data.MetricName.ValueString())
	}

	// Add dimensions if provided
	if len(data.Dimensions) > 0 {
		input.Dimensions = make([]types.DimensionFilter, len(data.Dimensions))
		for i, dim := range data.Dimensions {
			input.Dimensions[i] = types.DimensionFilter{
				Name:  aws.String(dim.Name.ValueString()),
				Value: aws.String(dim.Value.ValueString()),
			}
		}
	}

	// Call the AWS API
	result, err := d.cloudwatch.ListMetrics(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to retrieve CloudWatch metrics",
			fmt.Sprintf("Error calling AWS CloudWatch ListMetrics API: %s. Please verify your AWS credentials, permissions, and that the CloudWatch service is available in your region.", err),
		)
		return
	}

	// Convert the result to our model
	metrics := make([]MetricModel, len(result.Metrics))
	for i, metric := range result.Metrics {
		// Convert dimensions
		dimensions := make([]DimensionModel, len(metric.Dimensions))
		for j, dim := range metric.Dimensions {
			dimensions[j] = DimensionModel{
				Name:  basetypes.NewStringValue(aws.ToString(dim.Name)),
				Value: basetypes.NewStringValue(aws.ToString(dim.Value)),
			}
		}

		metrics[i] = MetricModel{
			Namespace:  basetypes.NewStringValue(aws.ToString(metric.Namespace)),
			MetricName: basetypes.NewStringValue(aws.ToString(metric.MetricName)),
			Dimensions: dimensions,
		}
	}

	// Set the data
	data.Metrics = metrics
	data.Id = basetypes.NewStringValue("cloudwatch_metrics")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
