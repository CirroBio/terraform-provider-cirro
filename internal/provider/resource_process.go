package provider

import (
	"context"
	"fmt"

	cirroclient "github.com/cirro-bio/terraform-provider-cirro/internal/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ resource.Resource = &ProcessResource{}
var _ resource.ResourceWithImportState = &ProcessResource{}

func NewProcessResource() resource.Resource {
	return &ProcessResource{}
}

type ProcessResource struct {
	client *cirroclient.Client
}

type ProcessResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	Description          types.String `tfsdk:"description"`
	Executor             types.String `tfsdk:"executor"`
	DataType             types.String `tfsdk:"data_type"`
	Category             types.String `tfsdk:"category"`
	DocumentationURL     types.String `tfsdk:"documentation_url"`
	FileRequirementsMsg  types.String `tfsdk:"file_requirements_message"`
	ChildProcessIDs      types.List   `tfsdk:"child_process_ids"`
	ParentProcessIDs     types.List   `tfsdk:"parent_process_ids"`
	LinkedProjectIDs     types.List   `tfsdk:"linked_project_ids"`
	IsTenantWide         types.Bool   `tfsdk:"is_tenant_wide"`
	AllowMultipleSources types.Bool   `tfsdk:"allow_multiple_sources"`
	UsesSampleSheet      types.Bool   `tfsdk:"uses_sample_sheet"`
	PipelineCode         types.Object `tfsdk:"pipeline_code"`
	CustomSettings       types.Object `tfsdk:"custom_settings"`
	Maintainers          types.List   `tfsdk:"maintainers"`
	ExecutionModes       types.List   `tfsdk:"execution_modes"`
	Tags                 types.List   `tfsdk:"tags"`
	ConfigSource         types.String `tfsdk:"config_source"`
	FileMappingRules     types.List   `tfsdk:"file_mapping_rules"`
	// Computed
	Owner      types.String `tfsdk:"owner"`
	IsArchived types.Bool   `tfsdk:"is_archived"`
	CreatedAt  types.String `tfsdk:"created_at"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
}

type PipelineCodeModel struct {
	RepositoryPath  types.String `tfsdk:"repository_path"`
	Version         types.String `tfsdk:"version"`
	RepositoryType  types.String `tfsdk:"repository_type"`
	EntryPoint      types.String `tfsdk:"entry_point"`
	ExecutorVersion types.String `tfsdk:"executor_version"`
}

type CustomSettingsModel struct {
	Repository     types.String `tfsdk:"repository"`
	Branch         types.String `tfsdk:"branch"`
	Folder         types.String `tfsdk:"folder"`
	RepositoryType types.String `tfsdk:"repository_type"`
	// Computed
	LastSync   types.String `tfsdk:"last_sync"`
	SyncStatus types.String `tfsdk:"sync_status"`
	CommitHash types.String `tfsdk:"commit_hash"`
}

type FileMappingRuleModel struct {
	Description      types.String `tfsdk:"description"`
	Min              types.Int64  `tfsdk:"min"`
	Max              types.Int64  `tfsdk:"max"`
	IsSample         types.Bool   `tfsdk:"is_sample"`
	FileNamePatterns types.List   `tfsdk:"file_name_patterns"`
}

type FileNamePatternModel struct {
	ExampleName           types.String `tfsdk:"example_name"`
	Description           types.String `tfsdk:"description"`
	SampleMatchingPattern types.String `tfsdk:"sample_matching_pattern"`
}

var fileNamePatternAttrTypes = map[string]attr.Type{
	"example_name":            types.StringType,
	"description":             types.StringType,
	"sample_matching_pattern": types.StringType,
}

var fileMappingRuleAttrTypes = map[string]attr.Type{
	"description":        types.StringType,
	"min":                types.Int64Type,
	"max":                types.Int64Type,
	"is_sample":          types.BoolType,
	"file_name_patterns": types.ListType{ElemType: types.ObjectType{AttrTypes: fileNamePatternAttrTypes}},
}

var pipelineCodeAttrTypes = map[string]attr.Type{
	"repository_path":  types.StringType,
	"version":          types.StringType,
	"repository_type":  types.StringType,
	"entry_point":      types.StringType,
	"executor_version": types.StringType,
}

var customSettingsAttrTypes = map[string]attr.Type{
	"repository":      types.StringType,
	"branch":          types.StringType,
	"folder":          types.StringType,
	"repository_type": types.StringType,
	"last_sync":       types.StringType,
	"sync_status":     types.StringType,
	"commit_hash":     types.StringType,
}

func (r *ProcessResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_process"
}

func (r *ProcessResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a custom Cirro process (pipeline or ingest data type). Destroying archives the process.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "Unique process ID (lowercase letters, numbers, underscores, dashes; 4-80 chars). Cannot be changed after creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(4, 80),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Friendly name for the process (4-80 characters).",
				Validators: []validator.String{
					stringvalidator.LengthBetween(4, 80),
				},
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Description of the process (4-500 characters).",
				Validators: []validator.String{
					stringvalidator.LengthBetween(4, 500),
				},
			},
			"executor": schema.StringAttribute{
				Required:    true,
				Description: "How the workflow is executed: INGEST, NEXTFLOW, CROMWELL, or OMICS_READY2RUN.",
				Validators: []validator.String{
					stringvalidator.OneOf("INGEST", "NEXTFLOW", "CROMWELL", "OMICS_READY2RUN"),
				},
			},
			"child_process_ids": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "IDs of processes that can be run downstream of this one.",
			},
			"parent_process_ids": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "IDs of processes that produce input for this one.",
			},
			"linked_project_ids": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "IDs of projects that can run this process.",
			},
			"data_type": schema.StringAttribute{
				Optional:    true,
				Description: "Name of the data type this process produces.",
			},
			"category": schema.StringAttribute{
				Optional:    true,
				Description: "Category label shown in the UI (e.g. Microbial Analysis).",
			},
			"documentation_url": schema.StringAttribute{
				Optional:    true,
				Description: "Link to process documentation.",
			},
			"file_requirements_message": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the files to be uploaded (INGEST processes).",
			},
			"is_tenant_wide": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the process is shared across the entire tenant.",
			},
			"allow_multiple_sources": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the process accepts multiple dataset sources.",
			},
			"uses_sample_sheet": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Whether the process uses the Cirro-provided sample sheet.",
			},
			"pipeline_code": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Location of the workflow analysis code (not required for INGEST).",
				Attributes: map[string]schema.Attribute{
					"repository_path": schema.StringAttribute{
						Required:    true,
						Description: "GitHub repository containing the workflow code (org/repo).",
					},
					"version": schema.StringAttribute{
						Required:    true,
						Description: "Branch, tag, or commit hash of the workflow code.",
					},
					"repository_type": schema.StringAttribute{
						Required:    true,
						Description: "Repository type: NONE, AWS, GITHUB_PUBLIC, or GITHUB_PRIVATE.",
						Validators: []validator.String{
							stringvalidator.OneOf("NONE", "AWS", "GITHUB_PUBLIC", "GITHUB_PRIVATE"),
						},
					},
					"entry_point": schema.StringAttribute{
						Required:    true,
						Description: "Main script for running the workflow (e.g. main.nf).",
					},
					"executor_version": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString(""),
						Description: "Version of the executor.",
					},
				},
			},
			"custom_settings": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Location of the Cirro process definition in a GitHub repository.",
				Attributes: map[string]schema.Attribute{
					"repository": schema.StringAttribute{
						Required:    true,
						Description: "GitHub repository containing the process definition (org/repo).",
					},
					"branch": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString("main"),
						Description: "Branch, tag, or commit hash of the process definition.",
					},
					"folder": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Default:     stringdefault.StaticString(".cirro"),
						Description: "Folder within the repo containing the process definition.",
					},
					"repository_type": schema.StringAttribute{
						Optional:    true,
						Description: "Repository type: NONE, AWS, GITHUB_PUBLIC, or GITHUB_PRIVATE.",
						Validators: []validator.String{
							stringvalidator.OneOf("NONE", "AWS", "GITHUB_PUBLIC", "GITHUB_PRIVATE"),
						},
					},
					// Computed
					"last_sync": schema.StringAttribute{
						Computed:    true,
						Description: "Timestamp of the last successful sync from the repository.",
					},
					"sync_status": schema.StringAttribute{
						Computed:    true,
						Description: "Status of the last repository sync.",
					},
					"commit_hash": schema.StringAttribute{
						Computed:    true,
						Description: "Commit hash of the last successful sync.",
					},
				},
			},
			"maintainers": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Usernames of other users who can manage the process.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"execution_modes": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Execution modes the pipeline supports: STANDARD (AWS Batch) and/or HEALTHOMICS (AWS HealthOmics).",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(stringvalidator.OneOf("STANDARD", "HEALTHOMICS")),
				},
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Process tags.",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
			},
			"config_source": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Storage backend for the process configuration files: WITH_PROCESS, EXTERNAL, or STORAGE.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("WITH_PROCESS", "EXTERNAL", "STORAGE"),
				},
			},
			"file_mapping_rules": schema.ListNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Rules describing the groups of files expected for the process (INGEST processes).",
				PlanModifiers: []planmodifier.List{
					listplanmodifier.UseStateForUnknown(),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"description": schema.StringAttribute{
							Required:    true,
							Description: "Describes the group of files that meet a single file type criteria (1-500 characters).",
							Validators: []validator.String{
								stringvalidator.LengthBetween(1, 500),
							},
						},
						"min": schema.Int64Attribute{
							Optional:    true,
							Description: "Minimum number of files to expect for the file type group.",
							Validators: []validator.Int64{
								int64validator.AtLeast(0),
							},
						},
						"max": schema.Int64Attribute{
							Optional:    true,
							Description: "Maximum number of files to expect for the file type group.",
							Validators: []validator.Int64{
								int64validator.AtMost(50000),
							},
						},
						"is_sample": schema.BoolAttribute{
							Optional:    true,
							Computed:    true,
							Default:     booldefault.StaticBool(false),
							Description: "Whether the file type is associated with a sample.",
						},
						"file_name_patterns": schema.ListNestedAttribute{
							Required:    true,
							Description: "File name patterns to expect for the file type group.",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"example_name": schema.StringAttribute{
										Required:    true,
										Description: "User-readable name for the file type used for display.",
									},
									"description": schema.StringAttribute{
										Optional:    true,
										Description: "File description.",
									},
									"sample_matching_pattern": schema.StringAttribute{
										Required:    true,
										Description: "Regex used to extract the sample name and other metadata from file names.",
									},
								},
							},
						},
					},
				},
			},
			// Computed
			"owner": schema.StringAttribute{
				Computed:    true,
				Description: "Username of the process creator.",
			},
			"is_archived": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the process has been archived.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the process was created (ISO 8601).",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp when the process was last updated (ISO 8601).",
			},
		},
	}
}

func (r *ProcessResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*cirroclient.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("got %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *ProcessResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProcessResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := planToProcessInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.CreateProcess(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Error creating process", err.Error())
		return
	}

	process, err := r.client.GetProcess(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading process after create", err.Error())
		return
	}

	resp.Diagnostics.Append(processDetailToState(ctx, process, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProcessResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProcessResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	process, err := r.client.GetProcess(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading process", err.Error())
		return
	}

	resp.Diagnostics.Append(processDetailToState(ctx, process, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ProcessResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProcessResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := planToProcessInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.UpdateProcess(ctx, plan.ID.ValueString(), input); err != nil {
		resp.Diagnostics.AddError("Error updating process", err.Error())
		return
	}

	process, err := r.client.GetProcess(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading process after update", err.Error())
		return
	}

	resp.Diagnostics.Append(processDetailToState(ctx, process, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProcessResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProcessResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.ArchiveProcess(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error archiving process", err.Error())
	}
}

func (r *ProcessResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	process, err := r.client.GetProcess(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing process", err.Error())
		return
	}

	var state ProcessResourceModel
	resp.Diagnostics.Append(processDetailToState(ctx, process, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ---- helpers ----

func planToProcessInput(ctx context.Context, m ProcessResourceModel) (cirroclient.ProcessInput, diag.Diagnostics) {
	var diags diag.Diagnostics

	var childIDs, parentIDs, linkedIDs []string
	diags.Append(m.ChildProcessIDs.ElementsAs(ctx, &childIDs, false)...)
	diags.Append(m.ParentProcessIDs.ElementsAs(ctx, &parentIDs, false)...)
	diags.Append(m.LinkedProjectIDs.ElementsAs(ctx, &linkedIDs, false)...)

	input := cirroclient.ProcessInput{
		ID:                   m.ID.ValueString(),
		Name:                 m.Name.ValueString(),
		Description:          m.Description.ValueString(),
		Executor:             m.Executor.ValueString(),
		DataType:             m.DataType.ValueString(),
		Category:             m.Category.ValueString(),
		DocumentationURL:     m.DocumentationURL.ValueString(),
		FileRequirementsMsg:  m.FileRequirementsMsg.ValueString(),
		ChildProcessIDs:      childIDs,
		ParentProcessIDs:     parentIDs,
		LinkedProjectIDs:     linkedIDs,
		IsTenantWide:         m.IsTenantWide.ValueBool(),
		AllowMultipleSources: m.AllowMultipleSources.ValueBool(),
		UsesSampleSheet:      m.UsesSampleSheet.ValueBool(),
	}

	if !m.PipelineCode.IsNull() && !m.PipelineCode.IsUnknown() {
		var pc PipelineCodeModel
		diags.Append(m.PipelineCode.As(ctx, &pc, basetypes.ObjectAsOptions{})...)
		input.PipelineCode = &cirroclient.PipelineCode{
			RepositoryPath:  pc.RepositoryPath.ValueString(),
			Version:         pc.Version.ValueString(),
			RepositoryType:  pc.RepositoryType.ValueString(),
			EntryPoint:      pc.EntryPoint.ValueString(),
			ExecutorVersion: pc.ExecutorVersion.ValueString(),
		}
	}

	if isKnown(m.Maintainers) {
		diags.Append(m.Maintainers.ElementsAs(ctx, &input.Maintainers, false)...)
	}
	if isKnown(m.ExecutionModes) {
		diags.Append(m.ExecutionModes.ElementsAs(ctx, &input.ExecutionModes, false)...)
	}
	if isKnown(m.Tags) {
		var tagStrings []string
		diags.Append(m.Tags.ElementsAs(ctx, &tagStrings, false)...)
		for _, t := range tagStrings {
			input.Tags = append(input.Tags, cirroclient.Tag{Value: t})
		}
	}
	if !m.ConfigSource.IsUnknown() {
		input.ConfigSource = m.ConfigSource.ValueString()
	}
	if isKnown(m.FileMappingRules) {
		rules, d := fileMappingRulesFromTF(ctx, m.FileMappingRules)
		diags.Append(d...)
		input.FileMappingRules = rules
	}

	if !m.CustomSettings.IsNull() && !m.CustomSettings.IsUnknown() {
		var cs CustomSettingsModel
		diags.Append(m.CustomSettings.As(ctx, &cs, basetypes.ObjectAsOptions{})...)
		input.CustomSettings = &cirroclient.CustomPipelineSettings{
			Repository:     cs.Repository.ValueString(),
			Branch:         cs.Branch.ValueString(),
			Folder:         cs.Folder.ValueString(),
			RepositoryType: cs.RepositoryType.ValueString(),
		}
	}

	return input, diags
}

func processDetailToState(ctx context.Context, p *cirroclient.ProcessDetail, m *ProcessResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ID = types.StringValue(p.ID)
	m.Name = types.StringValue(p.Name)
	m.Description = types.StringValue(p.Description)
	m.Executor = types.StringValue(p.Executor)
	m.DataType = types.StringValue(p.DataType)
	m.Category = types.StringValue(p.Category)
	m.DocumentationURL = types.StringValue(p.DocumentationURL)
	m.FileRequirementsMsg = types.StringValue(p.FileRequirementsMsg)
	m.IsTenantWide = types.BoolValue(p.IsTenantWide)
	m.AllowMultipleSources = types.BoolValue(p.AllowMultipleSources)
	m.UsesSampleSheet = types.BoolValue(p.UsesSampleSheet)
	m.Owner = types.StringValue(p.Owner)
	m.IsArchived = types.BoolValue(p.IsArchived)
	m.CreatedAt = types.StringValue(p.CreatedAt)
	m.UpdatedAt = types.StringValue(p.UpdatedAt)

	childIDs, d := types.ListValueFrom(ctx, types.StringType, p.ChildProcessIDs)
	diags.Append(d...)
	m.ChildProcessIDs = childIDs

	parentIDs, d := types.ListValueFrom(ctx, types.StringType, p.ParentProcessIDs)
	diags.Append(d...)
	m.ParentProcessIDs = parentIDs

	linkedIDs, d := types.ListValueFrom(ctx, types.StringType, p.LinkedProjectIDs)
	diags.Append(d...)
	m.LinkedProjectIDs = linkedIDs

	maintainers, d := types.ListValueFrom(ctx, types.StringType, nonNil(p.Maintainers))
	diags.Append(d...)
	m.Maintainers = maintainers

	execModes, d := types.ListValueFrom(ctx, types.StringType, nonNil(p.ExecutionModes))
	diags.Append(d...)
	m.ExecutionModes = execModes

	tagStrings := make([]string, len(p.Tags))
	for i, t := range p.Tags {
		tagStrings[i] = t.Value
	}
	tags, d := types.ListValueFrom(ctx, types.StringType, tagStrings)
	diags.Append(d...)
	m.Tags = tags

	m.ConfigSource = types.StringValue(p.ConfigSource)

	rules, d := fileMappingRulesToTF(p.FileMappingRules)
	diags.Append(d...)
	m.FileMappingRules = rules

	if p.PipelineCode != nil {
		pc, d := types.ObjectValue(pipelineCodeAttrTypes, map[string]attr.Value{
			"repository_path":  types.StringValue(p.PipelineCode.RepositoryPath),
			"version":          types.StringValue(p.PipelineCode.Version),
			"repository_type":  types.StringValue(p.PipelineCode.RepositoryType),
			"entry_point":      types.StringValue(p.PipelineCode.EntryPoint),
			"executor_version": types.StringValue(p.PipelineCode.ExecutorVersion),
		})
		diags.Append(d...)
		m.PipelineCode = pc
	} else {
		m.PipelineCode = types.ObjectNull(pipelineCodeAttrTypes)
	}

	if p.CustomSettings != nil {
		cs, d := types.ObjectValue(customSettingsAttrTypes, map[string]attr.Value{
			"repository":      types.StringValue(p.CustomSettings.Repository),
			"branch":          types.StringValue(p.CustomSettings.Branch),
			"folder":          types.StringValue(p.CustomSettings.Folder),
			"repository_type": types.StringValue(p.CustomSettings.RepositoryType),
			"last_sync":       types.StringValue(p.CustomSettings.LastSync),
			"sync_status":     types.StringValue(p.CustomSettings.SyncStatus),
			"commit_hash":     types.StringValue(p.CustomSettings.CommitHash),
		})
		diags.Append(d...)
		m.CustomSettings = cs
	} else {
		m.CustomSettings = types.ObjectNull(customSettingsAttrTypes)
	}

	return diags
}

func isKnown(l types.List) bool {
	return !l.IsNull() && !l.IsUnknown()
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func fileMappingRulesFromTF(ctx context.Context, l types.List) ([]cirroclient.FileMappingRule, diag.Diagnostics) {
	var diags diag.Diagnostics

	var ruleModels []FileMappingRuleModel
	diags.Append(l.ElementsAs(ctx, &ruleModels, false)...)

	rules := make([]cirroclient.FileMappingRule, 0, len(ruleModels))
	for _, rm := range ruleModels {
		rule := cirroclient.FileMappingRule{
			Description: rm.Description.ValueString(),
			IsSample:    rm.IsSample.ValueBool(),
		}
		if !rm.Min.IsNull() && !rm.Min.IsUnknown() {
			v := int(rm.Min.ValueInt64())
			rule.Min = &v
		}
		if !rm.Max.IsNull() && !rm.Max.IsUnknown() {
			v := int(rm.Max.ValueInt64())
			rule.Max = &v
		}

		var patternModels []FileNamePatternModel
		diags.Append(rm.FileNamePatterns.ElementsAs(ctx, &patternModels, false)...)
		for _, pm := range patternModels {
			rule.FileNamePatterns = append(rule.FileNamePatterns, cirroclient.FileNamePattern{
				ExampleName:           pm.ExampleName.ValueString(),
				Description:           pm.Description.ValueString(),
				SampleMatchingPattern: pm.SampleMatchingPattern.ValueString(),
			})
		}
		rules = append(rules, rule)
	}

	return rules, diags
}

func fileMappingRulesToTF(rules []cirroclient.FileMappingRule) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	ruleType := types.ObjectType{AttrTypes: fileMappingRuleAttrTypes}
	patternType := types.ObjectType{AttrTypes: fileNamePatternAttrTypes}

	ruleObjs := make([]attr.Value, len(rules))
	for i, r := range rules {
		patternObjs := make([]attr.Value, len(r.FileNamePatterns))
		for j, p := range r.FileNamePatterns {
			obj, d := types.ObjectValue(fileNamePatternAttrTypes, map[string]attr.Value{
				"example_name":            types.StringValue(p.ExampleName),
				"description":             nullableString(p.Description),
				"sample_matching_pattern": types.StringValue(p.SampleMatchingPattern),
			})
			diags.Append(d...)
			patternObjs[j] = obj
		}
		patterns, d := types.ListValue(patternType, patternObjs)
		diags.Append(d...)

		obj, d := types.ObjectValue(fileMappingRuleAttrTypes, map[string]attr.Value{
			"description":        types.StringValue(r.Description),
			"min":                optionalInt64(r.Min),
			"max":                optionalInt64(r.Max),
			"is_sample":          types.BoolValue(r.IsSample),
			"file_name_patterns": patterns,
		})
		diags.Append(d...)
		ruleObjs[i] = obj
	}

	list, d := types.ListValue(ruleType, ruleObjs)
	diags.Append(d...)
	return list, diags
}

func optionalInt64(v *int) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*v))
}
