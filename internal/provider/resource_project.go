package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	cirroclient "github.com/cirro-bio/terraform-provider-cirro/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ resource.Resource = &ProjectResource{}
var _ resource.ResourceWithImportState = &ProjectResource{}
var _ resource.ResourceWithValidateConfig = &ProjectResource{}

func NewProjectResource() resource.Resource {
	return &ProjectResource{}
}

// The backend only accepts settings changes once a deployment has settled, so
// Update waits for one of these statuses when the API reports the project busy.
const (
	projectSettleTimeout  = 30 * time.Minute
	projectSettleInterval = 15 * time.Second
)

// emptyStringList is the default for optional list attributes. The API returns
// an empty list (never null) for these, so state must hold the same.
var emptyStringList = types.ListValueMust(types.StringType, []attr.Value{})

type ProjectResource struct {
	client *cirroclient.Client
}

type ProjectResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	BillingAccountID  types.String `tfsdk:"billing_account_id"`
	Status            types.String `tfsdk:"status"`
	StatusMessage     types.String `tfsdk:"status_message"`
	Organization      types.String `tfsdk:"organization"`
	ClassificationIDs types.List   `tfsdk:"classification_ids"`
	Tags              types.List   `tfsdk:"tags"`
	Contacts          types.List   `tfsdk:"contacts"`
	Settings          types.Object `tfsdk:"settings"`
	Account           types.Object `tfsdk:"account"`
	CreatedBy         types.String `tfsdk:"created_by"`
	CreatedAt         types.String `tfsdk:"created_at"`
	UpdatedAt         types.String `tfsdk:"updated_at"`
	DeployedAt        types.String `tfsdk:"deployed_at"`
	DeployedVersion   types.String `tfsdk:"deployed_version"`
	RoleVersion       types.String `tfsdk:"role_version"`
}

type ProjectSettingsModel struct {
	BudgetAmount                 types.Int64  `tfsdk:"budget_amount"`
	BudgetPeriod                 types.String `tfsdk:"budget_period"`
	EnableBackup                 types.Bool   `tfsdk:"enable_backup"`
	EnableSftp                   types.Bool   `tfsdk:"enable_sftp"`
	RetentionPolicyDays          types.Int64  `tfsdk:"retention_policy_days"`
	TemporaryStorageLifetimeDays types.Int64  `tfsdk:"temporary_storage_lifetime_days"`
	ServiceConnections           types.List   `tfsdk:"service_connections"`
	KmsArn                       types.String `tfsdk:"kms_arn"`
	VpcID                        types.String `tfsdk:"vpc_id"`
	BatchSubnets                 types.List   `tfsdk:"batch_subnets"`
	WorkspaceSubnets             types.List   `tfsdk:"workspace_subnets"`
	MaxSpotVCPU                  types.Int64  `tfsdk:"max_spot_vcpu"`
	MaxFPGAVCPU                  types.Int64  `tfsdk:"max_fpga_vcpu"`
	MaxGPUVCPU                   types.Int64  `tfsdk:"max_gpu_vcpu"`
	EnableDragen                 types.Bool   `tfsdk:"enable_dragen"`
	DragenAmi                    types.String `tfsdk:"dragen_ami"`
	MaxWorkspacesVCPU            types.Int64  `tfsdk:"max_workspaces_vcpu"`
	MaxWorkspacesGPUVCPU         types.Int64  `tfsdk:"max_workspaces_gpu_vcpu"`
	MaxWorkspacesPerUser         types.Int64  `tfsdk:"max_workspaces_per_user"`
	EnableAdvancedGpuConfig      types.Bool   `tfsdk:"enable_advanced_gpu_config"`
	EnableCustomWorkspaceRoles   types.Bool   `tfsdk:"enable_custom_workspace_roles"`
	MaxSharedFilesystems         types.Int64  `tfsdk:"max_shared_filesystems"`
	IsDiscoverable               types.Bool   `tfsdk:"is_discoverable"`
	IsShareable                  types.Bool   `tfsdk:"is_shareable"`
	IsAiEnabled                  types.Bool   `tfsdk:"is_ai_enabled"`
	// Computed — populated from API response.
	HasPipelinesEnabled         types.Bool `tfsdk:"has_pipelines_enabled"`
	HasWorkspacesEnabled        types.Bool `tfsdk:"has_workspaces_enabled"`
	HasSharedFilesystemsEnabled types.Bool `tfsdk:"has_shared_filesystems_enabled"`
	IsOmicsAvailable            types.Bool `tfsdk:"is_omics_available"`
}

type ContactModel struct {
	Name         types.String `tfsdk:"name"`
	Organization types.String `tfsdk:"organization"`
	Email        types.String `tfsdk:"email"`
	Phone        types.String `tfsdk:"phone"`
}

type CloudAccountModel struct {
	AccountID   types.String `tfsdk:"account_id"`
	AccountName types.String `tfsdk:"account_name"`
	RegionName  types.String `tfsdk:"region_name"`
	AccountType types.String `tfsdk:"account_type"`
}

var settingsAttrTypes = map[string]attr.Type{
	"budget_amount":                   types.Int64Type,
	"budget_period":                   types.StringType,
	"enable_backup":                   types.BoolType,
	"enable_sftp":                     types.BoolType,
	"retention_policy_days":           types.Int64Type,
	"temporary_storage_lifetime_days": types.Int64Type,
	"service_connections":             types.ListType{ElemType: types.StringType},
	"kms_arn":                         types.StringType,
	"vpc_id":                          types.StringType,
	"batch_subnets":                   types.ListType{ElemType: types.StringType},
	"workspace_subnets":               types.ListType{ElemType: types.StringType},
	"max_spot_vcpu":                   types.Int64Type,
	"max_fpga_vcpu":                   types.Int64Type,
	"max_gpu_vcpu":                    types.Int64Type,
	"enable_dragen":                   types.BoolType,
	"dragen_ami":                      types.StringType,
	"max_workspaces_vcpu":             types.Int64Type,
	"max_workspaces_gpu_vcpu":         types.Int64Type,
	"max_workspaces_per_user":         types.Int64Type,
	"enable_advanced_gpu_config":      types.BoolType,
	"enable_custom_workspace_roles":   types.BoolType,
	"max_shared_filesystems":          types.Int64Type,
	"is_discoverable":                 types.BoolType,
	"is_shareable":                    types.BoolType,
	"is_ai_enabled":                   types.BoolType,
	"has_pipelines_enabled":           types.BoolType,
	"has_workspaces_enabled":          types.BoolType,
	"has_shared_filesystems_enabled":  types.BoolType,
	"is_omics_available":              types.BoolType,
}

var contactAttrTypes = map[string]attr.Type{
	"name":         types.StringType,
	"organization": types.StringType,
	"email":        types.StringType,
	"phone":        types.StringType,
}

var accountAttrTypes = map[string]attr.Type{
	"account_id":   types.StringType,
	"account_name": types.StringType,
	"region_name":  types.StringType,
	"account_type": types.StringType,
}

func (r *ProjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *ProjectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Cirro project.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Project identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Project name (3-100 characters).",
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Project description.",
			},
			"billing_account_id": schema.StringAttribute{
				Required:    true,
				Description: "Billing account identifier.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Project status.",
			},
			"status_message": schema.StringAttribute{
				Computed:    true,
				Description: "Additional status detail.",
			},
			"organization": schema.StringAttribute{
				Computed:    true,
				Description: "Organization the project belongs to.",
			},
			"created_by": schema.StringAttribute{
				Computed:    true,
				Description: "User who created the project.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp the project was created (ISO 8601).",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp the project was last updated (ISO 8601).",
			},
			"deployed_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp the project was last deployed (ISO 8601).",
			},
			"deployed_version": schema.StringAttribute{
				Computed:    true,
				Description: "Version of the project infrastructure that was last deployed.",
			},
			"role_version": schema.StringAttribute{
				Computed:    true,
				Description: "Version of the Cirro deployment role installed in the project's cloud account.",
			},
			"classification_ids": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(emptyStringList),
				ElementType: types.StringType,
				Description: "Data classification identifiers.",
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(emptyStringList),
				ElementType: types.StringType,
				Description: "Project tags.",
			},
			"contacts": schema.ListNestedAttribute{
				Required:    true,
				Description: "Project contacts (1-10). Each contact must have name, organization, email, and phone.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":         schema.StringAttribute{Required: true},
						"organization": schema.StringAttribute{Required: true},
						"email":        schema.StringAttribute{Required: true},
						"phone":        schema.StringAttribute{Required: true},
					},
				},
			},
			"settings": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Project settings.",
				Attributes: map[string]schema.Attribute{
					"budget_amount": schema.Int64Attribute{
						Required:    true,
						Description: "Budget amount (must be > 0).",
					},
					"budget_period": schema.StringAttribute{
						Required:    true,
						Description: "Budget period: MONTHLY, QUARTERLY, or ANNUALLY.",
					},
					"enable_backup": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Enable backup for the project.",
					},
					"enable_sftp": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Enable SFTP access for the project.",
					},
					"retention_policy_days": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(7),
						Description: "Data retention period in days.",
					},
					"temporary_storage_lifetime_days": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(14),
						Description: "Temporary storage lifetime in days.",
					},
					"service_connections": schema.ListAttribute{
						Optional:    true,
						Computed:    true,
						Default:     listdefault.StaticValue(emptyStringList),
						ElementType: types.StringType,
						Description: "Service connection identifiers.",
					},
					"kms_arn": schema.StringAttribute{
						Optional:    true,
						Description: "AWS KMS key ARN for encryption. Cannot be changed after project creation.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"vpc_id": schema.StringAttribute{
						Optional:    true,
						Description: "VPC identifier (format: vpc-*).",
					},
					"batch_subnets": schema.ListAttribute{
						Optional:    true,
						Computed:    true,
						Default:     listdefault.StaticValue(emptyStringList),
						ElementType: types.StringType,
						Description: "Subnet IDs used for batch compute (BYOA projects).",
					},
					"workspace_subnets": schema.ListAttribute{
						Optional:    true,
						Computed:    true,
						Default:     listdefault.StaticValue(emptyStringList),
						ElementType: types.StringType,
						Description: "Subnet IDs used for workspaces (BYOA projects).",
					},
					"max_spot_vcpu": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(0),
						Description: "Maximum spot vCPU quota.",
					},
					"max_fpga_vcpu": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(0),
						Description: "Maximum FPGA vCPU quota.",
					},
					"max_gpu_vcpu": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(0),
						Description: "Maximum GPU vCPU quota.",
					},
					"enable_dragen": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Whether DRAGEN acceleration is enabled.",
					},
					"dragen_ami": schema.StringAttribute{
						Optional:    true,
						Description: "AMI used for DRAGEN instances.",
					},
					"max_workspaces_vcpu": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(0),
						Description: "Maximum vCPU quota for workspaces.",
					},
					"max_workspaces_gpu_vcpu": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(0),
						Description: "Maximum GPU vCPU quota for workspaces.",
					},
					"max_workspaces_per_user": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(0),
						Description: "Maximum concurrent workspaces per user.",
					},
					"enable_advanced_gpu_config": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Whether advanced GPU configuration is enabled.",
					},
					"enable_custom_workspace_roles": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Whether custom workspace roles are enabled (BYOA projects only).",
					},
					"max_shared_filesystems": schema.Int64Attribute{
						Optional:    true,
						Computed:    true,
						Default:     int64default.StaticInt64(0),
						Description: "Maximum number of shared filesystems.",
					},
					"is_discoverable": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Whether the project is discoverable by other users.",
					},
					"is_shareable": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Whether datasets in the project can be shared.",
					},
					"is_ai_enabled": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
						Description: "Whether AI features are enabled.",
					},
					"has_pipelines_enabled": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether pipeline execution is enabled.",
					},
					"has_workspaces_enabled": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether interactive workspaces are enabled.",
					},
					"has_shared_filesystems_enabled": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether shared filesystems are enabled.",
					},
					"is_omics_available": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether the project is deployed with the resources needed for AWS HealthOmics runs.",
					},
				},
			},
			"account": schema.SingleNestedAttribute{
				Required:    true,
				Description: "Cloud account configuration. account_type must be set; BYOA also requires account_id, while HOSTED has Cirro provision the account. Attributes left unset are populated by Cirro. account_type and account_id cannot be changed after creation — modifying them requires replacing the project.",
				Attributes: map[string]schema.Attribute{
					"account_id": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "AWS account ID (12-digit). Required for BYOA projects; assigned by Cirro for hosted projects when unset. Cannot be changed after project creation.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
							stringplanmodifier.RequiresReplaceIfConfigured(),
						},
					},
					"account_name": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Name used to describe the account (BYOA projects).",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"region_name": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "AWS region code. Defaults to the region of the Cirro deployment.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"account_type": schema.StringAttribute{
						Required:    true,
						Description: "BYOA or HOSTED. Cannot be changed after project creation.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
			},
		},
	}
}

// ValidateConfig rejects a BYOA account without an account ID at plan time.
// Values that are not yet known (e.g. references to other resources) are
// skipped; the API enforces the same rule on create.
func (r *ProjectResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var account types.Object
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("account"), &account)...)
	if resp.Diagnostics.HasError() || account.IsNull() || account.IsUnknown() {
		return
	}

	var am CloudAccountModel
	resp.Diagnostics.Append(account.As(ctx, &am, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	if am.AccountType.ValueString() == "BYOA" && !am.AccountID.IsUnknown() && am.AccountID.ValueString() == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("account").AtName("account_id"),
			"Missing account ID",
			"account.account_id is required when account.account_type is \"BYOA\". Set it, or use account_type = \"HOSTED\" to have Cirro provision the account.",
		)
	}
}

func (r *ProjectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProjectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := planToProjectInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := r.client.CreateProject(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Error creating project", err.Error())
		return
	}

	resp.Diagnostics.Append(projectDetailToState(ctx, project, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProjectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := r.client.GetProject(ctx, state.ID.ValueString())
	if err != nil {
		if cirroclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading project", err.Error())
		return
	}

	resp.Diagnostics.Append(projectDetailToState(ctx, project, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProjectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := planToProjectInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := plan.ID.ValueString()

	// The API replaces the stored account with whatever is sent and rejects a
	// missing account or a changed account ID, so fill anything the plan does
	// not carry (e.g. an account ID assigned by Cirro) from the live project.
	current, err := r.client.GetProject(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading project before update", err.Error())
		return
	}
	input.Account = mergeAccount(input.Account, current.Account)

	project, err := r.client.UpdateProject(ctx, id, input)
	if err != nil && isProjectBusy(err) {
		// A deployment is still in flight; wait for it to settle and retry once.
		if err = waitForProjectSettled(ctx, r.client, id); err == nil {
			project, err = r.client.UpdateProject(ctx, id, input)
		}
	}
	if err != nil {
		resp.Diagnostics.AddError("Error updating project", err.Error())
		return
	}

	resp.Diagnostics.Append(projectDetailToState(ctx, project, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes the project from Terraform state. The Cirro API does not expose
// a project deletion endpoint, so the project is only removed from state.
func (r *ProjectResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning(
		"Project not deleted in Cirro",
		"Cirro does not expose a project deletion API. The project has been removed from Terraform state but still exists in Cirro.",
	)
}

func (r *ProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	project, err := r.client.GetProject(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing project", err.Error())
		return
	}

	var state ProjectResourceModel
	resp.Diagnostics.Append(projectDetailToState(ctx, project, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ---- helpers ----

// mergeAccount fills any account field missing from in with the value currently
// stored in Cirro.
func mergeAccount(in, current *cirroclient.CloudAccount) *cirroclient.CloudAccount {
	if current == nil {
		return in
	}
	if in == nil {
		out := *current
		return &out
	}
	out := *in
	if out.AccountID == "" {
		out.AccountID = current.AccountID
	}
	if out.AccountName == "" {
		out.AccountName = current.AccountName
	}
	if out.RegionName == "" {
		out.RegionName = current.RegionName
	}
	if out.AccountType == "" {
		out.AccountType = current.AccountType
	}
	return &out
}

// isProjectBusy reports whether an update was rejected because the project is
// still being deployed.
func isProjectBusy(err error) bool {
	var apiErr *cirroclient.APIError
	return errors.As(err, &apiErr) &&
		apiErr.StatusCode == http.StatusBadRequest &&
		strings.Contains(apiErr.Body, "is being updated")
}

// waitForProjectSettled polls until the project reaches a status in which the
// API accepts settings changes.
func waitForProjectSettled(ctx context.Context, c *cirroclient.Client, id string) error {
	ctx, cancel := context.WithTimeout(ctx, projectSettleTimeout)
	defer cancel()
	ticker := time.NewTicker(projectSettleInterval)
	defer ticker.Stop()

	for {
		p, err := c.GetProject(ctx, id)
		if err != nil {
			return err
		}
		switch p.Status {
		case "COMPLETED", "FAILED", "SUSPENDED":
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("timed out waiting for project %s to finish deploying (status %s)", id, p.Status)
		case <-ticker.C:
		}
	}
}

func planToProjectInput(ctx context.Context, plan ProjectResourceModel) (cirroclient.ProjectInput, diag.Diagnostics) {
	var diags diag.Diagnostics
	input := cirroclient.ProjectInput{
		Name:             plan.Name.ValueString(),
		Description:      plan.Description.ValueString(),
		BillingAccountID: plan.BillingAccountID.ValueString(),
	}

	// contacts
	var contactModels []ContactModel
	diags.Append(plan.Contacts.ElementsAs(ctx, &contactModels, false)...)
	for _, c := range contactModels {
		input.Contacts = append(input.Contacts, cirroclient.Contact{
			Name:         c.Name.ValueString(),
			Organization: c.Organization.ValueString(),
			Email:        c.Email.ValueString(),
			Phone:        c.Phone.ValueString(),
		})
	}

	// settings
	var sm ProjectSettingsModel
	diags.Append(plan.Settings.As(ctx, &sm, basetypes.ObjectAsOptions{})...)
	var svcConns, batchSubnets, workspaceSubnets []string
	diags.Append(sm.ServiceConnections.ElementsAs(ctx, &svcConns, false)...)
	diags.Append(sm.BatchSubnets.ElementsAs(ctx, &batchSubnets, false)...)
	diags.Append(sm.WorkspaceSubnets.ElementsAs(ctx, &workspaceSubnets, false)...)
	input.Settings = cirroclient.ProjectSettings{
		BudgetAmount:                 int(sm.BudgetAmount.ValueInt64()),
		BudgetPeriod:                 sm.BudgetPeriod.ValueString(),
		EnableBackup:                 sm.EnableBackup.ValueBool(),
		EnableSftp:                   sm.EnableSftp.ValueBool(),
		RetentionPolicyDays:          int(sm.RetentionPolicyDays.ValueInt64()),
		TemporaryStorageLifetimeDays: int(sm.TemporaryStorageLifetimeDays.ValueInt64()),
		KmsArn:                       sm.KmsArn.ValueString(),
		VpcID:                        sm.VpcID.ValueString(),
		ServiceConnections:           svcConns,
		BatchSubnets:                 batchSubnets,
		WorkspaceSubnets:             workspaceSubnets,
		MaxSpotVCPU:                  int(sm.MaxSpotVCPU.ValueInt64()),
		MaxFPGAVCPU:                  int(sm.MaxFPGAVCPU.ValueInt64()),
		MaxGPUVCPU:                   int(sm.MaxGPUVCPU.ValueInt64()),
		EnableDragen:                 sm.EnableDragen.ValueBool(),
		DragenAmi:                    sm.DragenAmi.ValueString(),
		MaxWorkspacesVCPU:            int(sm.MaxWorkspacesVCPU.ValueInt64()),
		MaxWorkspacesGPUVCPU:         int(sm.MaxWorkspacesGPUVCPU.ValueInt64()),
		MaxWorkspacesPerUser:         int(sm.MaxWorkspacesPerUser.ValueInt64()),
		EnableAdvancedGpuConfig:      sm.EnableAdvancedGpuConfig.ValueBool(),
		EnableCustomWorkspaceRoles:   sm.EnableCustomWorkspaceRoles.ValueBool(),
		MaxSharedFilesystems:         int(sm.MaxSharedFilesystems.ValueInt64()),
		IsDiscoverable:               sm.IsDiscoverable.ValueBool(),
		IsShareable:                  sm.IsShareable.ValueBool(),
		IsAiEnabled:                  sm.IsAiEnabled.ValueBool(),
	}

	// classification IDs
	var classIDs []string
	diags.Append(plan.ClassificationIDs.ElementsAs(ctx, &classIDs, false)...)
	input.ClassificationIDs = classIDs

	// tags
	var tagStrings []string
	diags.Append(plan.Tags.ElementsAs(ctx, &tagStrings, false)...)
	for _, t := range tagStrings {
		input.Tags = append(input.Tags, cirroclient.Tag{Value: t})
	}

	// account
	if !plan.Account.IsNull() && !plan.Account.IsUnknown() {
		var am CloudAccountModel
		diags.Append(plan.Account.As(ctx, &am, basetypes.ObjectAsOptions{})...)
		input.Account = &cirroclient.CloudAccount{
			AccountID:   am.AccountID.ValueString(),
			AccountName: am.AccountName.ValueString(),
			RegionName:  am.RegionName.ValueString(),
			AccountType: am.AccountType.ValueString(),
		}
	}

	return input, diags
}

func projectDetailToState(ctx context.Context, p *cirroclient.ProjectDetail, m *ProjectResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ID = types.StringValue(p.ID)
	m.Name = types.StringValue(p.Name)
	m.Description = types.StringValue(p.Description)
	m.BillingAccountID = types.StringValue(p.BillingAccountID)
	m.Status = types.StringValue(p.Status)
	m.StatusMessage = types.StringValue(p.StatusMessage)
	m.Organization = types.StringValue(p.Organization)
	m.CreatedBy = types.StringValue(p.CreatedBy)
	m.CreatedAt = types.StringValue(p.CreatedAt)
	m.UpdatedAt = types.StringValue(p.UpdatedAt)
	m.DeployedAt = types.StringValue(p.DeployedAt)
	m.DeployedVersion = types.StringValue(p.DeployedVersion)
	m.RoleVersion = types.StringValue(p.RoleVersion)

	m.ClassificationIDs = stringListValue(p.ClassificationIDs)

	tagStrings := make([]string, len(p.Tags))
	for i, t := range p.Tags {
		tagStrings[i] = t.Value
	}
	m.Tags = stringListValue(tagStrings)

	// contacts
	contactObjs := make([]attr.Value, len(p.Contacts))
	for i, c := range p.Contacts {
		obj, d := types.ObjectValue(contactAttrTypes, map[string]attr.Value{
			"name":         types.StringValue(c.Name),
			"organization": types.StringValue(c.Organization),
			"email":        types.StringValue(c.Email),
			"phone":        types.StringValue(c.Phone),
		})
		diags.Append(d...)
		contactObjs[i] = obj
	}
	contacts, d := types.ListValue(types.ObjectType{AttrTypes: contactAttrTypes}, contactObjs)
	diags.Append(d...)
	m.Contacts = contacts

	// settings
	settings, d := types.ObjectValue(settingsAttrTypes, map[string]attr.Value{
		"budget_amount":                   types.Int64Value(int64(p.Settings.BudgetAmount)),
		"budget_period":                   types.StringValue(p.Settings.BudgetPeriod),
		"enable_backup":                   types.BoolValue(p.Settings.EnableBackup),
		"enable_sftp":                     types.BoolValue(p.Settings.EnableSftp),
		"retention_policy_days":           types.Int64Value(int64(p.Settings.RetentionPolicyDays)),
		"temporary_storage_lifetime_days": types.Int64Value(int64(p.Settings.TemporaryStorageLifetimeDays)),
		"service_connections":             stringListValue(p.Settings.ServiceConnections),
		"kms_arn":                         nullableString(p.Settings.KmsArn),
		"vpc_id":                          nullableString(p.Settings.VpcID),
		"batch_subnets":                   stringListValue(p.Settings.BatchSubnets),
		"workspace_subnets":               stringListValue(p.Settings.WorkspaceSubnets),
		"max_spot_vcpu":                   types.Int64Value(int64(p.Settings.MaxSpotVCPU)),
		"max_fpga_vcpu":                   types.Int64Value(int64(p.Settings.MaxFPGAVCPU)),
		"max_gpu_vcpu":                    types.Int64Value(int64(p.Settings.MaxGPUVCPU)),
		"enable_dragen":                   types.BoolValue(p.Settings.EnableDragen),
		"dragen_ami":                      nullableString(p.Settings.DragenAmi),
		"max_workspaces_vcpu":             types.Int64Value(int64(p.Settings.MaxWorkspacesVCPU)),
		"max_workspaces_gpu_vcpu":         types.Int64Value(int64(p.Settings.MaxWorkspacesGPUVCPU)),
		"max_workspaces_per_user":         types.Int64Value(int64(p.Settings.MaxWorkspacesPerUser)),
		"enable_advanced_gpu_config":      types.BoolValue(p.Settings.EnableAdvancedGpuConfig),
		"enable_custom_workspace_roles":   types.BoolValue(p.Settings.EnableCustomWorkspaceRoles),
		"max_shared_filesystems":          types.Int64Value(int64(p.Settings.MaxSharedFilesystems)),
		"is_discoverable":                 types.BoolValue(p.Settings.IsDiscoverable),
		"is_shareable":                    types.BoolValue(p.Settings.IsShareable),
		"is_ai_enabled":                   types.BoolValue(p.Settings.IsAiEnabled),
		"has_pipelines_enabled":           types.BoolValue(p.Settings.HasPipelinesEnabled),
		"has_workspaces_enabled":          types.BoolValue(p.Settings.HasWorkspacesEnabled),
		"has_shared_filesystems_enabled":  types.BoolValue(p.Settings.HasSharedFilesystemsEnabled),
		"is_omics_available":              types.BoolValue(p.Settings.IsOmicsAvailable),
	})
	diags.Append(d...)
	m.Settings = settings

	// account
	if p.Account != nil {
		acct, d := types.ObjectValue(accountAttrTypes, map[string]attr.Value{
			"account_id":   nullableString(p.Account.AccountID),
			"account_name": nullableString(p.Account.AccountName),
			"region_name":  nullableString(p.Account.RegionName),
			"account_type": types.StringValue(p.Account.AccountType),
		})
		diags.Append(d...)
		m.Account = acct
	} else {
		m.Account = types.ObjectNull(accountAttrTypes)
	}

	return diags
}

// stringListValue converts a slice to a known list; nil becomes an empty list
// rather than null.
func stringListValue(raw []string) types.List {
	elems := make([]attr.Value, len(raw))
	for i, s := range raw {
		elems[i] = types.StringValue(s)
	}
	return types.ListValueMust(types.StringType, elems)
}

// nullableString returns null when v is empty so Optional attributes stay null
// in state when the API does not set them.
func nullableString(v string) attr.Value {
	if v == "" {
		return types.StringNull()
	}
	return types.StringValue(v)
}
