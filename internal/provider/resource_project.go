package provider

import (
	"context"
	"fmt"

	cirroclient "github.com/cirro-bio/terraform-provider-cirro/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ resource.Resource = &ProjectResource{}
var _ resource.ResourceWithImportState = &ProjectResource{}

func NewProjectResource() resource.Resource {
	return &ProjectResource{}
}

type ProjectResource struct {
	client *cirroclient.Client
}

type ProjectResourceModel struct {
	ID                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Description       types.String `tfsdk:"description"`
	BillingAccountID  types.String `tfsdk:"billing_account_id"`
	Status            types.String `tfsdk:"status"`
	Organization      types.String `tfsdk:"organization"`
	ClassificationIDs types.List   `tfsdk:"classification_ids"`
	Tags              types.List   `tfsdk:"tags"`
	Contacts          types.List   `tfsdk:"contacts"`
	Settings          types.Object `tfsdk:"settings"`
	Account           types.Object `tfsdk:"account"`
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
			"organization": schema.StringAttribute{
				Computed:    true,
				Description: "Organization the project belongs to.",
			},
			"classification_ids": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Data classification identifiers.",
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
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
						ElementType: types.StringType,
						Description: "Service connection identifiers.",
					},
					"kms_arn": schema.StringAttribute{
						Optional:    true,
						Description: "AWS KMS key ARN for encryption.",
					},
					"vpc_id": schema.StringAttribute{
						Optional:    true,
						Description: "VPC identifier (format: vpc-*).",
					},
				},
			},
			"account": schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Cloud account configuration (required for BYOA projects). account_type and account_id cannot be changed after creation — modifying them requires replacing the project.",
				Attributes: map[string]schema.Attribute{
					"account_id": schema.StringAttribute{
						Optional:    true,
						Description: "AWS account ID (12-digit). Cannot be changed after project creation.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"account_name": schema.StringAttribute{Optional: true},
					"region_name":  schema.StringAttribute{Optional: true},
					"account_type": schema.StringAttribute{
						Required:    true,
						Description: "HOSTED or BYOA. Cannot be changed after project creation.",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
			},
		},
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

	project, err := r.client.UpdateProject(ctx, plan.ID.ValueString(), input)
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
	input.Settings = cirroclient.ProjectSettings{
		BudgetAmount:                 int(sm.BudgetAmount.ValueInt64()),
		BudgetPeriod:                 sm.BudgetPeriod.ValueString(),
		EnableBackup:                 sm.EnableBackup.ValueBool(),
		EnableSftp:                   sm.EnableSftp.ValueBool(),
		RetentionPolicyDays:          int(sm.RetentionPolicyDays.ValueInt64()),
		TemporaryStorageLifetimeDays: int(sm.TemporaryStorageLifetimeDays.ValueInt64()),
		KmsArn:                       sm.KmsArn.ValueString(),
		VpcID:                        sm.VpcID.ValueString(),
	}
	var svcConns []string
	diags.Append(sm.ServiceConnections.ElementsAs(ctx, &svcConns, false)...)
	input.Settings.ServiceConnections = svcConns

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
	m.Organization = types.StringValue(p.Organization)

	classIDs, d := types.ListValueFrom(ctx, types.StringType, p.ClassificationIDs)
	diags.Append(d...)
	m.ClassificationIDs = classIDs

	tagStrings := make([]string, len(p.Tags))
	for i, t := range p.Tags {
		tagStrings[i] = t.Value
	}
	tags, d := types.ListValueFrom(ctx, types.StringType, tagStrings)
	diags.Append(d...)
	m.Tags = tags

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
	svcConns, d := types.ListValueFrom(ctx, types.StringType, p.Settings.ServiceConnections)
	diags.Append(d...)
	settings, d := types.ObjectValue(settingsAttrTypes, map[string]attr.Value{
		"budget_amount":                   types.Int64Value(int64(p.Settings.BudgetAmount)),
		"budget_period":                   types.StringValue(p.Settings.BudgetPeriod),
		"enable_backup":                   types.BoolValue(p.Settings.EnableBackup),
		"enable_sftp":                     types.BoolValue(p.Settings.EnableSftp),
		"retention_policy_days":           types.Int64Value(int64(p.Settings.RetentionPolicyDays)),
		"temporary_storage_lifetime_days": types.Int64Value(int64(p.Settings.TemporaryStorageLifetimeDays)),
		"service_connections":             svcConns,
		"kms_arn":                         types.StringValue(p.Settings.KmsArn),
		"vpc_id":                          types.StringValue(p.Settings.VpcID),
	})
	diags.Append(d...)
	m.Settings = settings

	// account
	if p.Account != nil {
		acct, d := types.ObjectValue(accountAttrTypes, map[string]attr.Value{
			"account_id":   types.StringValue(p.Account.AccountID),
			"account_name": types.StringValue(p.Account.AccountName),
			"region_name":  types.StringValue(p.Account.RegionName),
			"account_type": types.StringValue(p.Account.AccountType),
		})
		diags.Append(d...)
		m.Account = acct
	} else {
		m.Account = types.ObjectNull(accountAttrTypes)
	}

	return diags
}
