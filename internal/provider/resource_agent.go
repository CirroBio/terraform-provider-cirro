package provider

import (
	"context"
	"fmt"

	cirroclient "github.com/cirro-bio/terraform-provider-cirro/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &AgentResource{}
var _ resource.ResourceWithImportState = &AgentResource{}

func NewAgentResource() resource.Resource {
	return &AgentResource{}
}

type AgentResource struct {
	client *cirroclient.Client
}

type AgentResourceModel struct {
	ID                       types.String `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	AgentRoleArn             types.String `tfsdk:"agent_role_arn"`
	Status                   types.String `tfsdk:"status"`
	Tags                     types.Map    `tfsdk:"tags"`
	EnvironmentConfiguration types.Map    `tfsdk:"environment_configuration"`
	// Registration fields (computed, populated once the agent checks in)
	RegistrationHostname     types.String `tfsdk:"registration_hostname"`
	RegistrationOS           types.String `tfsdk:"registration_os"`
	RegistrationAgentVersion types.String `tfsdk:"registration_agent_version"`
	CreatedBy                types.String `tfsdk:"created_by"`
	CreatedAt                types.String `tfsdk:"created_at"`
	UpdatedAt                types.String `tfsdk:"updated_at"`
}

func (r *AgentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_agent"
}

func (r *AgentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Cirro compute agent. The agent software must be installed separately on your compute infrastructure and will register itself once running.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Display name for the agent.",
			},
			"agent_role_arn": schema.StringAttribute{
				Required:    true,
				Description: "ARN of the AWS IAM role or user the agent will assume (format: arn:aws:iam::...).",
			},
			"tags": schema.MapAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Key-value tags displayed to users when selecting this agent.",
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
			"environment_configuration": schema.MapAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "Environment configuration key-value pairs passed to the agent.",
				PlanModifiers: []planmodifier.Map{
					mapplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Agent status: ONLINE or OFFLINE.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"registration_hostname": schema.StringAttribute{
				Computed:    true,
				Description: "Hostname reported by the agent after registration.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"registration_os": schema.StringAttribute{
				Computed:    true,
				Description: "Operating system reported by the agent after registration.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"registration_agent_version": schema.StringAttribute{
				Computed:    true,
				Description: "Agent software version reported after registration.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_by": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{Computed: true},
		},
	}
}

func (r *AgentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *AgentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan AgentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := planToAgentInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	agent, err := r.client.CreateAgent(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Error creating agent", err.Error())
		return
	}

	resp.Diagnostics.Append(agentDetailToState(ctx, agent, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AgentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state AgentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	agent, err := r.client.GetAgent(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading agent", err.Error())
		return
	}

	resp.Diagnostics.Append(agentDetailToState(ctx, agent, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *AgentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan AgentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := planToAgentInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.UpdateAgent(ctx, plan.ID.ValueString(), input); err != nil {
		resp.Diagnostics.AddError("Error updating agent", err.Error())
		return
	}

	agent, err := r.client.GetAgent(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading agent after update", err.Error())
		return
	}

	resp.Diagnostics.Append(agentDetailToState(ctx, agent, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *AgentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state AgentResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteAgent(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting agent", err.Error())
	}
}

func (r *AgentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	agent, err := r.client.GetAgent(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing agent", err.Error())
		return
	}

	var state AgentResourceModel
	resp.Diagnostics.Append(agentDetailToState(ctx, agent, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ---- helpers ----

func planToAgentInput(ctx context.Context, m AgentResourceModel) (cirroclient.AgentInput, diag.Diagnostics) {
	var diags diag.Diagnostics

	var tags map[string]string
	diags.Append(m.Tags.ElementsAs(ctx, &tags, false)...)

	var envConfig map[string]string
	diags.Append(m.EnvironmentConfiguration.ElementsAs(ctx, &envConfig, false)...)

	return cirroclient.AgentInput{
		Name:                     m.Name.ValueString(),
		AgentRoleArn:             m.AgentRoleArn.ValueString(),
		Tags:                     tags,
		EnvironmentConfiguration: envConfig,
	}, diags
}

func agentDetailToState(ctx context.Context, a *cirroclient.AgentDetail, m *AgentResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ID = types.StringValue(a.ID)
	m.Name = types.StringValue(a.Name)
	m.AgentRoleArn = types.StringValue(a.AgentRoleArn)
	m.Status = types.StringValue(a.Status)
	m.CreatedBy = types.StringValue(a.CreatedBy)
	m.CreatedAt = types.StringValue(a.CreatedAt)
	m.UpdatedAt = types.StringValue(a.UpdatedAt)

	tags, d := mapStringToTF(ctx, a.Tags)
	diags.Append(d...)
	m.Tags = tags

	envConfig, d := mapStringToTF(ctx, a.EnvironmentConfiguration)
	diags.Append(d...)
	m.EnvironmentConfiguration = envConfig

	if a.Registration != nil {
		m.RegistrationHostname = types.StringValue(a.Registration.Hostname)
		m.RegistrationOS = types.StringValue(a.Registration.OS)
		m.RegistrationAgentVersion = types.StringValue(a.Registration.AgentVersion)
	} else {
		m.RegistrationHostname = types.StringValue("")
		m.RegistrationOS = types.StringValue("")
		m.RegistrationAgentVersion = types.StringValue("")
	}

	return diags
}

func mapStringToTF(ctx context.Context, m map[string]string) (types.Map, diag.Diagnostics) {
	if len(m) == 0 {
		return types.MapValueMust(types.StringType, map[string]attr.Value{}), nil
	}
	return types.MapValueFrom(ctx, types.StringType, m)
}
