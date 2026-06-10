package provider

import (
	"context"
	"fmt"

	cirroclient "github.com/cirro-bio/terraform-provider-cirro/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &BillingAccountResource{}
var _ resource.ResourceWithImportState = &BillingAccountResource{}

func NewBillingAccountResource() resource.Resource {
	return &BillingAccountResource{}
}

type BillingAccountResource struct {
	client *cirroclient.Client
}

type BillingAccountResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Organization        types.String `tfsdk:"organization"`
	CustomerType        types.String `tfsdk:"customer_type"`
	BillingMethod       types.String `tfsdk:"billing_method"`
	PrimaryBudgetNumber types.String `tfsdk:"primary_budget_number"`
	Owner               types.String `tfsdk:"owner"`
	SharedWith          types.List   `tfsdk:"shared_with"`
	Contacts            types.List   `tfsdk:"contacts"`
	IsArchived          types.Bool   `tfsdk:"is_archived"`
}

func (r *BillingAccountResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_billing_account"
}

func (r *BillingAccountResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Cirro billing account.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Billing account name.",
			},
			"organization": schema.StringAttribute{
				Computed:    true,
				Description: "Organization the billing account belongs to.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"customer_type": schema.StringAttribute{
				Required:    true,
				Description: "Customer type: INTERNAL, CONSORTIUM, or EXTERNAL.",
			},
			"billing_method": schema.StringAttribute{
				Required:    true,
				Description: "Billing method: BUDGET_NUMBER, PURCHASE_ORDER, or CREDIT.",
			},
			"primary_budget_number": schema.StringAttribute{
				Required:    true,
				Description: "Primary budget number or reference code.",
			},
			"owner": schema.StringAttribute{
				Required:    true,
				Description: "Username of the billing account owner.",
			},
			"shared_with": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Usernames the billing account is shared with.",
			},
			"contacts": schema.ListNestedAttribute{
				Required:    true,
				Description: "Billing contacts (at least one required).",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":         schema.StringAttribute{Required: true},
						"organization": schema.StringAttribute{Required: true},
						"email":        schema.StringAttribute{Required: true},
						"phone":        schema.StringAttribute{Required: true},
					},
				},
			},
			"is_archived": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the billing account is archived.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *BillingAccountResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BillingAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BillingAccountResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := planToBillingAccountRequest(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	account, err := r.client.CreateBillingAccount(ctx, input)
	if err != nil {
		resp.Diagnostics.AddError("Error creating billing account", err.Error())
		return
	}

	resp.Diagnostics.Append(billingAccountToState(ctx, account, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BillingAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BillingAccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	account, err := r.client.GetBillingAccount(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading billing account", err.Error())
		return
	}

	resp.Diagnostics.Append(billingAccountToState(ctx, account, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *BillingAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan BillingAccountResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input, diags := planToBillingAccountRequest(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.UpdateBillingAccount(ctx, plan.ID.ValueString(), input); err != nil {
		resp.Diagnostics.AddError("Error updating billing account", err.Error())
		return
	}

	// Re-read to get computed fields after update.
	account, err := r.client.GetBillingAccount(ctx, plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading billing account after update", err.Error())
		return
	}

	resp.Diagnostics.Append(billingAccountToState(ctx, account, &plan)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *BillingAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state BillingAccountResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteBillingAccount(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting billing account", err.Error())
	}
}

func (r *BillingAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	account, err := r.client.GetBillingAccount(ctx, req.ID)
	if err != nil {
		resp.Diagnostics.AddError("Error importing billing account", err.Error())
		return
	}

	var state BillingAccountResourceModel
	resp.Diagnostics.Append(billingAccountToState(ctx, account, &state)...)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// ---- helpers ----

func planToBillingAccountRequest(ctx context.Context, m BillingAccountResourceModel) (cirroclient.BillingAccountRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	var sharedWith []string
	diags.Append(m.SharedWith.ElementsAs(ctx, &sharedWith, false)...)

	var contactModels []ContactModel
	diags.Append(m.Contacts.ElementsAs(ctx, &contactModels, false)...)

	contacts := make([]cirroclient.Contact, len(contactModels))
	for i, c := range contactModels {
		contacts[i] = cirroclient.Contact{
			Name:         c.Name.ValueString(),
			Organization: c.Organization.ValueString(),
			Email:        c.Email.ValueString(),
			Phone:        c.Phone.ValueString(),
		}
	}

	return cirroclient.BillingAccountRequest{
		Name:                m.Name.ValueString(),
		CustomerType:        m.CustomerType.ValueString(),
		BillingMethod:       m.BillingMethod.ValueString(),
		PrimaryBudgetNumber: m.PrimaryBudgetNumber.ValueString(),
		Owner:               m.Owner.ValueString(),
		SharedWith:          sharedWith,
		Contacts:            contacts,
	}, diags
}

func billingAccountToState(ctx context.Context, a *cirroclient.BillingAccount, m *BillingAccountResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	m.ID = types.StringValue(a.ID)
	m.Name = types.StringValue(a.Name)
	m.Organization = types.StringValue(a.Organization)
	m.CustomerType = types.StringValue(a.CustomerType)
	m.BillingMethod = types.StringValue(a.BillingMethod)
	m.PrimaryBudgetNumber = types.StringValue(a.PrimaryBudgetNumber)
	m.Owner = types.StringValue(a.Owner)
	m.IsArchived = types.BoolValue(a.IsArchived)

	sharedWith, d := types.ListValueFrom(ctx, types.StringType, a.SharedWith)
	diags.Append(d...)
	m.SharedWith = sharedWith

	contactObjs := make([]attr.Value, len(a.Contacts))
	for i, c := range a.Contacts {
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

	return diags
}
