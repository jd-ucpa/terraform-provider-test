package ssm

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &SendFilesResource{}
var _ resource.ResourceWithImportState = &SendFilesResource{}

// stringvalidator.RegexMatches equivalent
type regexMatchesValidator struct {
	regexp *regexp.Regexp
	msg    string
}

func (v regexMatchesValidator) Description(ctx context.Context) string {
	return v.msg
}

func (v regexMatchesValidator) MarkdownDescription(ctx context.Context) string {
	return v.msg
}

func (v regexMatchesValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()
	if !v.regexp.MatchString(value) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid format",
			v.msg,
		)
	}
}

// stringvalidator.RegexMatches creates a regex validator
func stringvalidatorRegexMatches(regexp *regexp.Regexp, msg string) validator.String {
	return regexMatchesValidator{
		regexp: regexp,
		msg:    msg,
	}
}

// stringvalidator.StringLengthMin creates a minimum length validator
type stringLengthMinValidator struct {
	min int
	msg string
}

func (v stringLengthMinValidator) Description(ctx context.Context) string {
	return v.msg
}

func (v stringLengthMinValidator) MarkdownDescription(ctx context.Context) string {
	return v.msg
}

func (v stringLengthMinValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := strings.TrimSpace(req.ConfigValue.ValueString())
	if len(value) <= v.min {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid value",
			v.msg,
		)
	}
}

// stringvalidatorStringLengthMin creates a minimum length validator
func stringvalidatorStringLengthMin(min int, msg string) validator.String {
	return stringLengthMinValidator{
		min: min,
		msg: msg,
	}
}

func NewSendFilesResource() resource.Resource {
	return &SendFilesResource{}
}

// SendFilesResource defines the resource implementation.
type SendFilesResource struct {
	ssm *ssm.Client
}

// SendFilesResourceModel describes the resource data model.
type SendFilesResourceModel struct {
	Id                 types.String `tfsdk:"id"`
	CommandId          types.String `tfsdk:"command_id"`
	Status             types.String `tfsdk:"status"`
	Platform           types.String `tfsdk:"platform"`
	InstanceIds        types.List   `tfsdk:"instance_ids"`
	Targets            []Target     `tfsdk:"targets"`
	WorkingDirectory   types.String `tfsdk:"working_directory"`
	ScriptBeforeFiles  types.String `tfsdk:"script_before_files"`
	ScriptAfterFiles   types.String `tfsdk:"script_after_files"`
	Files              []File       `tfsdk:"file"`
	Triggers           types.Map    `tfsdk:"triggers"`
}

// Target represents a target for SSM command
type Target struct {
	Key    types.String `tfsdk:"key"`
	Values types.List   `tfsdk:"values"`
}

// File represents a file to be created
type File struct {
	Name             types.String `tfsdk:"name"`
	Content          types.String `tfsdk:"content"`
	Permissions      types.String `tfsdk:"permissions"`
	Owner            types.String `tfsdk:"owner"`
	Group            types.String `tfsdk:"group"`
	WorkingDirectory types.String `tfsdk:"-"` // Internal field for command generation, not exposed to Terraform
}

// PlatformRunner interface for different platforms
type PlatformRunner interface {
	DocumentName() string
	CommandScript(workingDirectory, script string) string
	CommandFile(file File) string
}

// PowerShell implementation
type PowerShell struct{}

func (p *PowerShell) DocumentName() string {
	return "AWS-RunPowerShellScript"
}

func (p *PowerShell) CommandScript(workingDirectory, script string) string {
	scriptBase64 := base64.StdEncoding.EncodeToString([]byte(script))
	return fmt.Sprintf(`Set-Location -Path "%s"
$c = [System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String("%s"))
Invoke-Expression "$c"`, workingDirectory, scriptBase64)
}

func (p *PowerShell) CommandFile(file File) string {
	contentBase64 := base64.StdEncoding.EncodeToString([]byte(file.Content.ValueString()))
	return fmt.Sprintf(`if (Test-Path -Path "%s") {
  Set-Location -Path "%s"
} else {
  Throw "PathNotFound %s"
  Exit 1
}
Remove-Item "%s" -Force -ErrorAction SilentlyContinue
[System.Text.Encoding]::UTF8.GetString([System.Convert]::FromBase64String("%s")) > "%s"`,
		file.WorkingDirectory.ValueString(), file.WorkingDirectory.ValueString(), file.WorkingDirectory.ValueString(),
		file.Name.ValueString(), contentBase64, file.Name.ValueString())
}

// Bash implementation
type Bash struct{}

func (b *Bash) DocumentName() string {
	return "AWS-RunShellScript"
}

func (b *Bash) CommandScript(workingDirectory, script string) string {
	scriptBase64 := base64.StdEncoding.EncodeToString([]byte(script))
	return fmt.Sprintf(`cd "%s"
echo %s | base64 -d | bash`, workingDirectory, scriptBase64)
}

func (b *Bash) CommandFile(file File) string {
	contentBase64 := base64.StdEncoding.EncodeToString([]byte(file.Content.ValueString()))
	commands := []string{
		fmt.Sprintf(`cd "%s"`, file.WorkingDirectory.ValueString()),
		fmt.Sprintf(`rm -f "%s"`, file.Name.ValueString()),
		fmt.Sprintf(`echo "%s" | base64 -d > "%s"`, contentBase64, file.Name.ValueString()),
	}

	// Add permissions if specified
	if !file.Permissions.IsNull() && !file.Permissions.IsUnknown() {
		commands = append(commands, fmt.Sprintf(`chmod %s "%s"`, file.Permissions.ValueString(), file.Name.ValueString()))
	}

	// Add owner/group if specified
	if (!file.Owner.IsNull() && !file.Owner.IsUnknown()) || (!file.Group.IsNull() && !file.Group.IsUnknown()) {
		chown := ""
		if !file.Owner.IsNull() && !file.Owner.IsUnknown() {
			chown = strings.TrimSpace(file.Owner.ValueString())
		}
		if !file.Group.IsNull() && !file.Group.IsUnknown() {
			if chown != "" {
				chown += ":"
			}
			chown += strings.TrimSpace(file.Group.ValueString())
		}
		if chown != "" {
			commands = append(commands, fmt.Sprintf(`chown %s "%s"`, chown, file.Name.ValueString()))
		}
	}

	return strings.Join(commands, "\n")
}

func (r *SendFilesResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssm_send_files"
}

func (r *SendFilesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "The `test_ssm_send_files` resource allows you to send files to EC2 instances using AWS Systems Manager (SSM). This resource supports creating files with custom permissions, owner, and group settings on Linux instances, and basic file creation on Windows instances. You can execute scripts before and after file creation for additional setup or verification tasks.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Unique identifier for the resource",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"command_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The ID of the SSM command",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The status of the SSM command",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"platform": schema.StringAttribute{
				MarkdownDescription: "The platform (linux or windows)",
				Required:            true,
			},
			"instance_ids": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of instance IDs to target",
				Optional:            true,
			},
			"working_directory": schema.StringAttribute{
				MarkdownDescription: "Working directory for the commands",
				Required:            true,
			},
			"script_before_files": schema.StringAttribute{
				MarkdownDescription: "Script to execute before creating files",
				Optional:            true,
			},
			"script_after_files": schema.StringAttribute{
				MarkdownDescription: "Script to execute after creating files",
				Optional:            true,
			},
			"triggers": schema.MapAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Triggers to force recreation",
				Optional:            true,
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.RequiresReplace(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"targets": schema.ListNestedBlock{
				MarkdownDescription: "Targets for the SSM command",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"key": schema.StringAttribute{
							MarkdownDescription: "Target key",
							Required:            true,
						},
						"values": schema.ListAttribute{
							ElementType:         types.StringType,
							MarkdownDescription: "Target values",
							Required:            true,
						},
					},
				},
			},
			"file": schema.ListNestedBlock{
				MarkdownDescription: "Files to create",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "File name",
							Required:            true,
						},
						"content": schema.StringAttribute{
							MarkdownDescription: "File content",
							Required:            true,
						},
						"permissions": schema.StringAttribute{
							MarkdownDescription: "File permissions (Linux only)",
							Optional:            true,
							Validators: []validator.String{
								stringvalidatorRegexMatches(
									regexp.MustCompile(`^[0-7]{3}$`),
									"must be a 3-digit octal string between 000 and 777",
								),
							},
						},
						"owner": schema.StringAttribute{
							MarkdownDescription: "File owner (Linux only)",
							Optional:            true,
							Validators: []validator.String{
								stringvalidatorStringLengthMin(0, "owner cannot be empty or contain only whitespace"),
							},
						},
						"group": schema.StringAttribute{
							MarkdownDescription: "File group (Linux only)",
							Optional:            true,
							Validators: []validator.String{
								stringvalidatorStringLengthMin(0, "group cannot be empty or contain only whitespace"),
							},
						},
					},
				},
			},
		},
	}
}

func (r *SendFilesResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Éviter le panic si le provider n'a pas été configuré
	if req.ProviderData == nil {
		return
	}

	// Vérifier que la configuration est du bon type
	config, ok := req.ProviderData.(aws.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Provider configuration error",
			fmt.Sprintf("Expected aws.Config for SSM send files resource, got: %T. This indicates a provider configuration issue. Please verify your provider configuration and report this issue if it persists.", req.ProviderData),
		)
		return
	}
	
	// Créer le client SSM à partir de la configuration AWS
	r.ssm = ssm.NewFromConfig(config)
}

func (r *SendFilesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SendFilesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Execute the common logic for creating/updating the resource
	data, diag := r.createOrUpdateResource(ctx, data)
	if diag.HasError() {
		resp.Diagnostics.Append(diag...)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SendFilesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SendFilesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// In a real implementation, you would check the command status
	// For now, we keep the current state
	
	// Normalize optional values to ensure consistency
	r.normalizeOptionalValues(&data)
	
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SendFilesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SendFilesResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state to compare triggers
	var currentData SendFilesResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &currentData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check if triggers have changed
	triggersChanged := !data.Triggers.Equal(currentData.Triggers)

	if triggersChanged {
		// Execute the common logic for creating/updating the resource
		var diag diag.Diagnostics
		data, diag = r.createOrUpdateResource(ctx, data)
		if diag.HasError() {
			resp.Diagnostics.Append(diag...)
			return
		}
	} else {
		// If triggers haven't changed, preserve computed values
		if data.Id.IsUnknown() || data.Id.IsNull() {
			data.Id = currentData.Id
		}
		if data.CommandId.IsUnknown() || data.CommandId.IsNull() {
			data.CommandId = currentData.CommandId
		}
		if data.Status.IsUnknown() || data.Status.IsNull() {
			data.Status = currentData.Status
		}
	}

	// Ensure computed values are always defined
	r.ensureComputedValues(&data)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SendFilesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// SSM commands cannot be deleted, we do nothing
}

func (r *SendFilesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// validateAndBuildTargets validates parameters and builds targets for SSM API
func (r *SendFilesResource) validateAndBuildTargets(ctx context.Context, data SendFilesResourceModel) ([]ssmtypes.Target, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	// Check that exactly one of the two (instance_ids or targets) is specified
	hasInstanceIds := !data.InstanceIds.IsNull()
	hasTargets := len(data.Targets) > 0

	if !hasInstanceIds && !hasTargets {
		diagnostics.AddError(
			"Invalid target configuration",
			"Either instance_ids or targets must be specified. Please provide at least one target for the SSM command.",
		)
		return nil, diagnostics
	}

	if hasInstanceIds && hasTargets {
		diagnostics.AddError(
			"Conflicting target configuration",
			"Cannot specify both instance_ids and targets. Use either instance_ids or targets, not both. Please choose one targeting method.",
		)
		return nil, diagnostics
	}

	var targets []ssmtypes.Target

	// If instance_ids is specified, use it
	if !data.InstanceIds.IsNull() {
		instanceIds := make([]string, 0, len(data.InstanceIds.Elements()))
		diagnostics.Append(data.InstanceIds.ElementsAs(ctx, &instanceIds, false)...)
		if diagnostics.HasError() {
			return nil, diagnostics
		}

		targets = []ssmtypes.Target{
			{
				Key:    aws.String("InstanceIds"),
				Values: instanceIds,
			},
		}
	} else {
		// Otherwise, use the specified targets
		targets = make([]ssmtypes.Target, len(data.Targets))
		for i, target := range data.Targets {
			targetValues := make([]string, len(target.Values.Elements()))
			diagnostics.Append(target.Values.ElementsAs(ctx, &targetValues, false)...)
			if diagnostics.HasError() {
				return nil, diagnostics
			}

			targets[i] = ssmtypes.Target{
				Key:    target.Key.ValueStringPointer(),
				Values: targetValues,
			}
		}
	}

	return targets, diagnostics
}

// buildCommands builds the commands array for SSM
func (r *SendFilesResource) buildCommands(data SendFilesResourceModel) ([]string, diag.Diagnostics) {
	var diagnostics diag.Diagnostics
	var commands []string

	// Get platform runner
	var runner PlatformRunner
	if data.Platform.ValueString() == "windows" {
		runner = &PowerShell{}
	} else {
		runner = &Bash{}
	}

	// Add script before files if specified and not empty
	if !data.ScriptBeforeFiles.IsNull() && !data.ScriptBeforeFiles.IsUnknown() && strings.TrimSpace(data.ScriptBeforeFiles.ValueString()) != "" {
		commands = append(commands, runner.CommandScript(data.WorkingDirectory.ValueString(), data.ScriptBeforeFiles.ValueString()))
	}

	// Add file commands
	for _, file := range data.Files {
		// Add working directory to file for command generation
		file.WorkingDirectory = data.WorkingDirectory
		commands = append(commands, runner.CommandFile(file))
	}

	// Add script after files if specified and not empty
	if !data.ScriptAfterFiles.IsNull() && !data.ScriptAfterFiles.IsUnknown() && strings.TrimSpace(data.ScriptAfterFiles.ValueString()) != "" {
		commands = append(commands, runner.CommandScript(data.WorkingDirectory.ValueString(), data.ScriptAfterFiles.ValueString()))
	}

	return commands, diagnostics
}

// executeSSMCommand executes an SSM command and handles polling
func (r *SendFilesResource) executeSSMCommand(ctx context.Context, data SendFilesResourceModel, targets []ssmtypes.Target, commands []string) (SendFilesResourceModel, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	// Get platform runner for document name
	var runner PlatformRunner
	if data.Platform.ValueString() == "windows" {
		runner = &PowerShell{}
	} else {
		runner = &Bash{}
	}

	// Convert commands to parameters format
	parameters := map[string][]string{
		"workingDirectory": {data.WorkingDirectory.ValueString()},
		"commands":         commands,
	}

	// Send SSM command
	command, err := r.ssm.SendCommand(ctx, &ssm.SendCommandInput{
		DocumentName: aws.String(runner.DocumentName()),
		Targets:      targets,
		Parameters:   parameters,
	})
	if err != nil {
		diagnostics.AddError(
			"Unable to send SSM command",
			fmt.Sprintf("Error calling AWS SSM SendCommand API: %s. Please verify your AWS credentials, permissions, and that the target instances are accessible via SSM.", err),
		)
		return data, diagnostics
	}

	data.Id = types.StringValue(*command.Command.CommandId)
	data.CommandId = types.StringValue(*command.Command.CommandId)
	data.Status = types.StringValue("InProgress")

	// Polling with timeout
	createTimeout := 5 * time.Minute
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	backoff := time.Second
	maxAttempts := 10
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if ctx.Err() != nil {
			diagnostics.AddError(
				"Operation cancelled",
				fmt.Sprintf("Context cancelled before attempt %d: %s. The operation was interrupted before completion.", attempt+1, ctx.Err()),
			)
			return data, diagnostics
		}

		attemptDiag := PollCommandInvocation(ctx, r.ssm, command)
		if attemptDiag.HasError() {
			// fmt.Printf("DEBUG: SSM command failed, setting status to Failed\n")
			data.Status = types.StringValue("Failed")
			return data, diagnostics
		} else if attemptDiag.WarningsCount() > 0 {
			select {
			case <-time.After(backoff):
				backoff *= 2
				if backoff > 30*time.Second {
					backoff = 30 * time.Second
				}
				continue
			case <-ctx.Done():
				if errors.Is(ctx.Err(), context.DeadlineExceeded) {
					diagnostics.AddError(
						"Timeout while waiting for SSM command to complete",
						fmt.Sprintf("Timeout occurred while waiting on command '%s' (polled %d times over %s). The command may still be running on the target instances.", *command.Command.CommandId, attempt+1, createTimeout),
					)
					return data, diagnostics
				} else {
					diagnostics.AddError(
						"Operation cancelled",
						fmt.Sprintf("Context cancelled before attempt %d: %s. The operation was interrupted before completion.", attempt+1, ctx.Err()),
					)
					return data, diagnostics
				}
			}
		} else {
			// fmt.Printf("DEBUG: Polling completed, getting actual status\n")
			actualStatus := getActualCommandStatus(ctx, r.ssm, command)
			// fmt.Printf("DEBUG: Actual status: %s\n", actualStatus)
			data.Status = types.StringValue(actualStatus)
			return data, diagnostics
		}
	}

	// fmt.Printf("DEBUG: Max attempts reached, getting final status\n")
	actualStatus := getActualCommandStatus(ctx, r.ssm, command)
	// fmt.Printf("DEBUG: Final status: %s\n", actualStatus)
	data.Status = types.StringValue(actualStatus)
	return data, diagnostics
}

// normalizeOptionalValues ensures optional values are defined
func (r *SendFilesResource) normalizeOptionalValues(data *SendFilesResourceModel) {
	// Only normalize triggers, not the script fields
	if data.Triggers.IsNull() && !data.Triggers.IsUnknown() {
		data.Triggers = types.MapNull(types.StringType)
	}
}

// createOrUpdateResource contains the common logic for creating or updating the resource
func (r *SendFilesResource) createOrUpdateResource(ctx context.Context, data SendFilesResourceModel) (SendFilesResourceModel, diag.Diagnostics) {
	var diagnostics diag.Diagnostics

	// Validate platform
	if data.Platform.ValueString() != "linux" && data.Platform.ValueString() != "windows" {
		diagnostics.AddError(
			"Invalid platform configuration",
			fmt.Sprintf("Platform must be either 'linux' or 'windows', got '%s'. Please specify a valid platform.", data.Platform.ValueString()),
		)
		return data, diagnostics
	}

	// Validate that at least one file is specified
	if len(data.Files) == 0 {
		diagnostics.AddError(
			"Invalid file configuration",
			"At least one file block must be specified. Please provide at least one file to send to the target instances.",
		)
		return data, diagnostics
	}

	// Validate and build targets
	targets, diag := r.validateAndBuildTargets(ctx, data)
	if diag.HasError() {
		diagnostics.Append(diag...)
		return data, diagnostics
	}

	// Build commands
	commands, diag := r.buildCommands(data)
	if diag.HasError() {
		diagnostics.Append(diag...)
		return data, diagnostics
	}

	// Execute SSM command
	data, diag = r.executeSSMCommand(ctx, data, targets, commands)
	if diag.HasError() {
		diagnostics.Append(diag...)
		return data, diagnostics
	}

	// Normalize optional values
	r.normalizeOptionalValues(&data)

	return data, diagnostics
}

// ensureComputedValues ensures computed values are always defined
func (r *SendFilesResource) ensureComputedValues(data *SendFilesResourceModel) {
	if data.Id.IsUnknown() || data.Id.IsNull() {
		data.Id = types.StringValue("")
	}
	if data.CommandId.IsUnknown() || data.CommandId.IsNull() {
		data.CommandId = types.StringValue("")
	}
	if data.Status.IsUnknown() || data.Status.IsNull() {
		data.Status = types.StringValue("")
	}
}
