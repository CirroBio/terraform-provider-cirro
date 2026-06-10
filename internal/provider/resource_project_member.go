package provider

import (
	"context"
	"fmt"
	"strings"

	cirroclient "github.com/cirro-bio/terraform-provider-cirro/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ProjectMemberResource{}
var _ resource.ResourceWithImportState = &ProjectMemberResource{}

func NewProjectMemberResource() resource.Resource {
	return &ProjectMemberResource{}
}

type ProjectMemberResource struct {
	client *cirroclient.Client
}

type ProjectMemberResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	ProjectID            types.String `tfsdk:"project_id"`
	Username             types.String `tfsdk:"username"`
	Role                 types.String `tfsdk:"role"`
	SuppressNotification types.Bool   `tfsdk:"suppress_notification"`
}

func (r *ProjectMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_member"
}

func (r *ProjectMemberResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a user's role within a Cirro project. Destroying this resource sets the user's role to NONE, removing their access.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Composite identifier: {project_id}/{username}.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "Project identifier.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"username": schema.StringAttribute{
				Required:    true,
				Description: "Cirro username of the member.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"role": schema.StringAttribute{
				Required:    true,
				Description: "Project role: OWNER, ADMIN, CONTRIBUTOR, or COLLABORATOR.",
			},
			"suppress_notification": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "When true, suppresses the email notification sent to the user.",
			},
		},
	}
}

func (r *ProjectMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProjectMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProjectMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.SetProjectUserRole(ctx, plan.ProjectID.ValueString(), cirroclient.SetUserProjectRoleRequest{
		Username:             plan.Username.ValueString(),
		Role:                 plan.Role.ValueString(),
		SuppressNotification: plan.SuppressNotification.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error setting project member role", err.Error())
		return
	}

	plan.ID = types.StringValue(plan.ProjectID.ValueString() + "/" + plan.Username.ValueString())
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProjectMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProjectMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	users, err := r.client.GetProjectUsers(ctx, state.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading project members", err.Error())
		return
	}

	for _, u := range users {
		if u.Username == state.Username.ValueString() {
			state.Role = types.StringValue(u.Role)
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}

	// User no longer has a role in this project — remove from state.
	resp.State.RemoveResource(ctx)
}

func (r *ProjectMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProjectMemberResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.SetProjectUserRole(ctx, plan.ProjectID.ValueString(), cirroclient.SetUserProjectRoleRequest{
		Username:             plan.Username.ValueString(),
		Role:                 plan.Role.ValueString(),
		SuppressNotification: plan.SuppressNotification.ValueBool(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating project member role", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ProjectMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProjectMemberResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.SetProjectUserRole(ctx, state.ProjectID.ValueString(), cirroclient.SetUserProjectRoleRequest{
		Username:             state.Username.ValueString(),
		Role:                 "NONE",
		SuppressNotification: true,
	})
	if err != nil {
		resp.Diagnostics.AddError("Error removing project member", err.Error())
	}
}

func (r *ProjectMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import format: {project_id}/{username}
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 {
		resp.Diagnostics.AddError("Invalid import ID", "Expected format: {project_id}/{username}")
		return
	}

	projectID, username := parts[0], parts[1]

	users, err := r.client.GetProjectUsers(ctx, projectID)
	if err != nil {
		resp.Diagnostics.AddError("Error reading project members", err.Error())
		return
	}

	for _, u := range users {
		if u.Username == username {
			state := ProjectMemberResourceModel{
				ID:                   types.StringValue(req.ID),
				ProjectID:            types.StringValue(projectID),
				Username:             types.StringValue(username),
				Role:                 types.StringValue(u.Role),
				SuppressNotification: types.BoolValue(false),
			}
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}

	resp.Diagnostics.AddError("User not found in project", fmt.Sprintf("user %q has no role in project %q", username, projectID))
}
