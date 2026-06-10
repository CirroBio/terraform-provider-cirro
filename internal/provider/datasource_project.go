package provider

import (
	"context"
	"fmt"

	cirroclient "github.com/cirro-bio/terraform-provider-cirro/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ProjectDataSource{}

func NewProjectDataSource() datasource.DataSource {
	return &ProjectDataSource{}
}

type ProjectDataSource struct {
	client *cirroclient.Client
}

type ProjectDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	BillingAccountID types.String `tfsdk:"billing_account_id"`
	Status           types.String `tfsdk:"status"`
	Organization     types.String `tfsdk:"organization"`
	ClassificationIDs types.List  `tfsdk:"classification_ids"`
	Tags             types.List   `tfsdk:"tags"`
	Contacts         types.List   `tfsdk:"contacts"`
	Settings         types.Object `tfsdk:"settings"`
	Account          types.Object `tfsdk:"account"`
}

func (d *ProjectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *ProjectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches details of a Cirro project by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "Project identifier.",
			},
			"name":               schema.StringAttribute{Computed: true},
			"description":        schema.StringAttribute{Computed: true},
			"billing_account_id": schema.StringAttribute{Computed: true},
			"status":             schema.StringAttribute{Computed: true},
			"organization":       schema.StringAttribute{Computed: true},
			"classification_ids": schema.ListAttribute{Computed: true, ElementType: types.StringType},
			"tags":               schema.ListAttribute{Computed: true, ElementType: types.StringType},
			"contacts": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":         schema.StringAttribute{Computed: true},
						"organization": schema.StringAttribute{Computed: true},
						"email":        schema.StringAttribute{Computed: true},
						"phone":        schema.StringAttribute{Computed: true},
					},
				},
			},
			"settings": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"budget_amount":                   schema.Int64Attribute{Computed: true},
					"budget_period":                   schema.StringAttribute{Computed: true},
					"enable_backup":                   schema.BoolAttribute{Computed: true},
					"enable_sftp":                     schema.BoolAttribute{Computed: true},
					"retention_policy_days":           schema.Int64Attribute{Computed: true},
					"temporary_storage_lifetime_days": schema.Int64Attribute{Computed: true},
					"service_connections":             schema.ListAttribute{Computed: true, ElementType: types.StringType},
					"kms_arn":                         schema.StringAttribute{Computed: true},
					"vpc_id":                          schema.StringAttribute{Computed: true},
				},
			},
			"account": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"account_id":   schema.StringAttribute{Computed: true},
					"account_name": schema.StringAttribute{Computed: true},
					"region_name":  schema.StringAttribute{Computed: true},
					"account_type": schema.StringAttribute{Computed: true},
				},
			},
		},
	}
}

func (d *ProjectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*cirroclient.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("got %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *ProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ProjectDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := d.client.GetProject(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading project", err.Error())
		return
	}

	// Reuse the resource state mapper (same field set).
	var state ProjectResourceModel
	resp.Diagnostics.Append(projectDetailToState(ctx, project, &state)...)

	// Map to data source model.
	config.ID = state.ID
	config.Name = state.Name
	config.Description = state.Description
	config.BillingAccountID = state.BillingAccountID
	config.Status = state.Status
	config.Organization = state.Organization
	config.ClassificationIDs = state.ClassificationIDs
	config.Tags = state.Tags
	config.Contacts = state.Contacts
	config.Settings = state.Settings
	config.Account = state.Account

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
