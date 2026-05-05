// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/action"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &ProductiveProvider{}

type ProductiveProvider struct {
	version string
}

type ProductiveProviderModel struct {
	Token          types.String `tfsdk:"token"`
	OrganizationID types.String `tfsdk:"organization_id"`
}

func (p *ProductiveProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "productive"
	resp.Version = p.version
}

func (p *ProductiveProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provider for managing resources in [Productive.io](https://productive.io). " +
			"Credentials can be supplied via the provider block or through the " +
			"`PRODUCTIVE_TOKEN` and `PRODUCTIVE_ORGANIZATION_ID` environment variables.",
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Productive.io API token. May also be set via `PRODUCTIVE_TOKEN`.",
			},
			"organization_id": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Productive.io organization ID. May also be set via `PRODUCTIVE_ORGANIZATION_ID`.",
			},
		},
	}
}

func (p *ProductiveProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ProductiveProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	token := os.Getenv("PRODUCTIVE_TOKEN")
	if !data.Token.IsNull() && !data.Token.IsUnknown() {
		token = data.Token.ValueString()
	}

	organizationID := os.Getenv("PRODUCTIVE_ORGANIZATION_ID")
	if !data.OrganizationID.IsNull() && !data.OrganizationID.IsUnknown() {
		organizationID = data.OrganizationID.ValueString()
	}

	if token == "" {
		resp.Diagnostics.AddError(
			"Missing Productive.io API token",
			"Set the `token` provider attribute or the `PRODUCTIVE_TOKEN` environment variable.",
		)
	}
	if organizationID == "" {
		resp.Diagnostics.AddError(
			"Missing Productive.io organization ID",
			"Set the `organization_id` provider attribute or the `PRODUCTIVE_ORGANIZATION_ID` environment variable.",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	client := NewClient(token, organizationID)
	resp.ResourceData = client
	resp.DataSourceData = client
}

func (p *ProductiveProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewPersonResource,
	}
}

func (p *ProductiveProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *ProductiveProvider) Functions(_ context.Context) []func() function.Function {
	return []func() function.Function{}
}

func (p *ProductiveProvider) EphemeralResources(_ context.Context) []func() ephemeral.EphemeralResource {
	return []func() ephemeral.EphemeralResource{}
}

func (p *ProductiveProvider) Actions(_ context.Context) []func() action.Action {
	return []func() action.Action{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ProductiveProvider{
			version: version,
		}
	}
}
