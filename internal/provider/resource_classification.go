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

var _ resource.Resource = &ClassificationResource{}
var _ resource.ResourceWithImportState = &ClassificationResource{}

func NewClassificationResource() resource.Resource {
	return &ClassificationResource{}
}

type ClassificationResource struct {
	client *cirroclient.Client
}

type ClassificationResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	RequirementIDs types.List   `tfsdk:"requirement_ids"`
	CreatedBy      types.String `tfsdk:"created_by"`
	CreatedAt      types.String `tfsdk:"created_at"`
	UpdatedAt      types.String `tfsdk:"updated_at"`
}

func (r *ClassificationResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_classification"
}

func (r *ClassificationResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Cirro data governance classification. Classifications are applied to projects to enforce compliance requirements.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Classification name (max 100 characters).",
			},
			"description": schema.StringAttribute{
				Required:    true,
				Description: "Description of what this classification means.",
			},
			"requirement_ids": schema.ListAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Description: "IDs of governance requirements attached to this classification.",
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

func (r *ClassificationResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ClassificationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ClassificationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := planToClassificationInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	classification, err := r.client.CreateClassification(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Error creating classification", err.Error())
		return
	}

	resp.Diagnostics.Append(classificationToState(ctx, classification, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ClassificationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ClassificationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	classification, err := r.client.GetClassification(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading classification", err.Error())
		return
	}

	resp.Diagnostics.Append(classificationToState(ctx, classification, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ClassificationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ClassificationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := planToClassificationInput(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	classification, err := r.client.UpdateClassification(ctx, plan.ID.ValueString(), input)
	if err != nil {
		resp.Diagnostics.AddError("Error updating classification", err.Error())
		return
	}

	resp.Diagnostics.Append(classificationToState(ctx, classification, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ClassificationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ClassificationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteClassification(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting classification", err.Error())
	}
}

func (r *ClassificationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	classification, err := r.client.GetClassification(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing classification", err.Error())
		return
	}

	var state ClassificationResourceModel
	resp.Diagnostics.Append(classificationToState(ctx, classification, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ---- helpers ----

func planToClassificationInput(ctx context.Context, m ClassificationResourceModel) (cirroclient.ClassificationInput, diag.Diagnostics) {
	var diags diag.Diagnostics

	var reqIDs []string
	diags.Append(m.RequirementIDs.ElementsAs(ctx, &reqIDs, false)...)

	return cirroclient.ClassificationInput{
		Name:           m.Name.ValueString(),
		Description:    m.Description.ValueString(),
		RequirementIDs: reqIDs,
	}, diags
}

func classificationToState(ctx context.Context, c *cirroclient.GovernanceClassification, m *ClassificationResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ID = types.StringValue(c.ID)
	m.Name = types.StringValue(c.Name)
	m.Description = types.StringValue(c.Description)
	m.CreatedBy = types.StringValue(c.CreatedBy)
	m.CreatedAt = types.StringValue(c.CreatedAt)
	m.UpdatedAt = types.StringValue(c.UpdatedAt)

	reqIDs, d := types.ListValueFrom(ctx, types.StringType, c.RequirementIDs)
	diags.Append(d...)
	m.RequirementIDs = reqIDs

	return diags
}
