package provider

import (
	"context"
	"fmt"

	cirroclient "github.com/cirro-bio/terraform-provider-cirro/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &UserResource{}
var _ resource.ResourceWithImportState = &UserResource{}

func NewUserResource() resource.Resource {
	return &UserResource{}
}

type UserResource struct {
	client *cirroclient.Client
}

type UserResourceModel struct {
	Username     types.String `tfsdk:"username"`
	Name         types.String `tfsdk:"name"`
	Email        types.String `tfsdk:"email"`
	Organization types.String `tfsdk:"organization"`
	Phone        types.String `tfsdk:"phone"`
	Department   types.String `tfsdk:"department"`
	JobTitle     types.String `tfsdk:"job_title"`
	GlobalRoles  types.List   `tfsdk:"global_roles"`
}

func (r *UserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *UserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Invites a user to Cirro and manages their profile. " +
			"Destroying this resource removes it from state only — Cirro does not expose a user deletion API.",
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Computed:    true,
				Description: "Cirro-assigned username (populated after invite).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Full name (3-70 characters).",
			},
			"email": schema.StringAttribute{
				Required:    true,
				Description: "Email address. Used to look up the user after invitation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"organization": schema.StringAttribute{
				Required:    true,
				Description: "Organization name (2-40 characters).",
			},
			"phone": schema.StringAttribute{
				Optional:    true,
				Description: "Phone number.",
			},
			"department": schema.StringAttribute{
				Optional:    true,
				Description: "Department.",
			},
			"job_title": schema.StringAttribute{
				Optional:    true,
				Description: "Job title.",
			},
			"global_roles": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Global roles assigned to the user (admin-only).",
			},
		},
	}
}

func (r *UserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan UserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.InviteUser(ctx, cirroclient.InviteUserRequest{
		Name:         plan.Name.ValueString(),
		Organization: plan.Organization.ValueString(),
		Email:        plan.Email.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error inviting user", err.Error())
		return
	}

	found, err := r.client.FindUserByEmail(ctx, plan.Email.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error locating invited user", err.Error())
		return
	}

	if !plan.Phone.IsNull() || !plan.Department.IsNull() || !plan.JobTitle.IsNull() {
		_, err = r.client.UpdateUser(ctx, found.Username, buildUpdateRequest(plan))
		if err != nil {
			resp.Diagnostics.AddError("Error updating user profile after invite", err.Error())
			return
		}
	}

	user, err := r.client.GetUser(ctx, found.Username)
	if err != nil {
		resp.Diagnostics.AddError("Error reading user after invite", err.Error())
		return
	}

	resp.Diagnostics.Append(userDetailToState(ctx, user, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state UserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.GetUser(ctx, state.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading user", err.Error())
		return
	}

	resp.Diagnostics.Append(userDetailToState(ctx, user, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan UserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var globalRoles []string
	resp.Diagnostics.Append(plan.GlobalRoles.ElementsAs(ctx, &globalRoles, false)...)

	updateReq := buildUpdateRequest(plan)
	updateReq.GlobalRoles = globalRoles

	user, err := r.client.UpdateUser(ctx, plan.Username.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating user", err.Error())
		return
	}

	resp.Diagnostics.Append(userDetailToState(ctx, user, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes the user from Terraform state only.
func (r *UserResource) Delete(_ context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning(
		"User not deleted in Cirro",
		"Cirro does not expose a user deletion API. The user has been removed from Terraform state but still exists in Cirro.",
	)
}

func (r *UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	user, err := r.client.GetUser(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing user", err.Error())
		return
	}

	var state UserResourceModel
	resp.Diagnostics.Append(userDetailToState(ctx, user, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ---- helpers ----

func buildUpdateRequest(m UserResourceModel) cirroclient.UpdateUserRequest {
	return cirroclient.UpdateUserRequest{
		Name:         m.Name.ValueString(),
		Email:        m.Email.ValueString(),
		Phone:        m.Phone.ValueString(),
		Department:   m.Department.ValueString(),
		JobTitle:     m.JobTitle.ValueString(),
		Organization: m.Organization.ValueString(),
	}
}

func userDetailToState(ctx context.Context, u *cirroclient.UserDetail, m *UserResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	m.Username = types.StringValue(u.Username)
	m.Name = types.StringValue(u.Name)
	m.Email = types.StringValue(u.Email)
	m.Organization = types.StringValue(u.Organization)
	m.Phone = types.StringValue(u.Phone)
	m.Department = types.StringValue(u.Department)
	m.JobTitle = types.StringValue(u.JobTitle)

	roles, d := types.ListValueFrom(ctx, types.StringType, u.GlobalRoles)
	diags.Append(d...)
	m.GlobalRoles = roles

	return diags
}
