package provider

import (
	"context"
	"fmt"

	cirroclient "github.com/cirro-bio/terraform-provider-cirro/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &UserDataSource{}

func NewUserDataSource() datasource.DataSource {
	return &UserDataSource{}
}

type UserDataSource struct {
	client *cirroclient.Client
}

type UserDataSourceModel struct {
	Username     types.String `tfsdk:"username"`
	Name         types.String `tfsdk:"name"`
	Email        types.String `tfsdk:"email"`
	Organization types.String `tfsdk:"organization"`
	Phone        types.String `tfsdk:"phone"`
	Department   types.String `tfsdk:"department"`
	JobTitle     types.String `tfsdk:"job_title"`
	GlobalRoles  types.List   `tfsdk:"global_roles"`
}

func (d *UserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *UserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches details of a Cirro user by username.",
		Attributes: map[string]schema.Attribute{
			"username": schema.StringAttribute{
				Required:    true,
				Description: "Cirro username.",
			},
			"name":         schema.StringAttribute{Computed: true},
			"email":        schema.StringAttribute{Computed: true},
			"organization": schema.StringAttribute{Computed: true},
			"phone":        schema.StringAttribute{Computed: true},
			"department":   schema.StringAttribute{Computed: true},
			"job_title":    schema.StringAttribute{Computed: true},
			"global_roles": schema.ListAttribute{Computed: true, ElementType: types.StringType},
		},
	}
}

func (d *UserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config UserDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := d.client.GetUser(ctx, config.Username.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading user", err.Error())
		return
	}

	var state UserResourceModel
	resp.Diagnostics.Append(userDetailToState(ctx, user, &state)...)

	config.Username = state.Username
	config.Name = state.Name
	config.Email = state.Email
	config.Organization = state.Organization
	config.Phone = state.Phone
	config.Department = state.Department
	config.JobTitle = state.JobTitle
	config.GlobalRoles = state.GlobalRoles

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
