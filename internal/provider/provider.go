package provider

import (
	"context"
	"fmt"
	"os"
	"strings"

	cirroauth "github.com/cirro-bio/terraform-provider-cirro/internal/auth"
	cirroclient "github.com/cirro-bio/terraform-provider-cirro/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &CirroProvider{}
var _ provider.ProviderWithFunctions = &CirroProvider{}

type CirroProvider struct {
	version string
}

type CirroProviderModel struct {
	BaseURL      types.String `tfsdk:"base_url"`
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &CirroProvider{version: version}
	}
}

func (p *CirroProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "cirro"
	resp.Version = p.version
}

func (p *CirroProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage Cirro resources via the Cirro API.",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Optional: true,
				Description: "Cirro tenant base URL, e.g. https://app.cirro.bio. " +
					"May also be set via CIRRO_BASE_URL.",
			},
			"client_id": schema.StringAttribute{
				Optional:    true,
				Description: "OAuth client ID. May also be set via CIRRO_CLIENT_ID.",
			},
			"client_secret": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "OAuth client secret. May also be set via CIRRO_CLIENT_SECRET.",
			},
		},
	}
}

func (p *CirroProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config CirroProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	baseURL := firstNonEmpty(config.BaseURL.ValueString(), os.Getenv("CIRRO_BASE_URL"))
	clientID := firstNonEmpty(config.ClientID.ValueString(), os.Getenv("CIRRO_CLIENT_ID"))
	clientSecret := firstNonEmpty(config.ClientSecret.ValueString(), os.Getenv("CIRRO_CLIENT_SECRET"))

	if baseURL == "" {
		resp.Diagnostics.AddError("Missing base_url",
			"Set the base_url attribute or the CIRRO_BASE_URL environment variable.")
	}
	if clientID == "" {
		resp.Diagnostics.AddError("Missing client_id",
			"Set the client_id attribute or the CIRRO_CLIENT_ID environment variable.")
	}
	if clientSecret == "" {
		resp.Diagnostics.AddError("Missing client_secret",
			"Set the client_secret attribute or the CIRRO_CLIENT_SECRET environment variable.")
	}
	if resp.Diagnostics.HasError() {
		return
	}

	baseURL = strings.TrimRight(baseURL, "/")
	apiURL := fmt.Sprintf("%s/api", baseURL)
	tokenURL := fmt.Sprintf("%s/auth/token", apiURL)

	creds := cirroauth.NewClientCredentials(clientID, clientSecret, tokenURL)
	client := cirroclient.New(apiURL, creds.GetToken)

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *CirroProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewProjectResource,
		NewProjectMemberResource,
		NewUserResource,
		NewBillingAccountResource,
		NewAgentResource,
		NewClassificationResource,
		NewPipelineResource,
	}
}

func (p *CirroProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewProjectDataSource,
		NewUserDataSource,
	}
}

func (p *CirroProvider) Functions(_ context.Context) []func() function.Function {
	return []func() function.Function{}
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
