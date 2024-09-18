/*
Copyright 2024 Chainguard, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	registry "chainguard.dev/sdk/proto/platform/registry/v1"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &groupDataSource{}
	_ datasource.DataSourceWithConfigure = &groupDataSource{}
)

// NewGroupDataSource is a helper function to simplify the provider implementation.
func NewVersionsDataSource() datasource.DataSource {
	return &versionsDataSource{}
}

// groupDataSource is the data source implementation.
type versionsDataSource struct {
	dataSource
}

type versionsDataSourceModel struct {
	Package  types.String `tfsdk:"name"`
	Metadata types.Object `tfsdk:"metadata"`
}

func (d versionsDataSourceModel) InputParams() string {
	return fmt.Sprintf("[package=%s]", d.Package)
}

// Metadata returns the data source type name.
func (d *versionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_versions"
}

func (d *versionsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(ctx, req, resp)
}

// Schema defines the schema for the data source.
func (d *versionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lookup version metadata for the given package name.",
		Attributes: map[string]schema.Attribute{
			"package": schema.StringAttribute{
				Description: "The name of the package to lookup",
				Optional:    false,
			},
			"metadata": schema.SingleNestedAttribute{
				Optional: false,
				Attributes: map[string]schema.Attribute{
					"eolVersions":          versionSchema(),
					"lastUpdatedTimestamp": schema.StringAttribute{},
					"latestVersion":        schema.StringAttribute{},
					"versions":             versionSchema(),
				},
			},
		},
	}
}

func versionSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"eolDate":     schema.StringAttribute{},
				"exists":      schema.BoolAttribute{},
				"fips":        schema.BoolAttribute{},
				"lts":         schema.StringAttribute{},
				"releaseDate": schema.StringAttribute{},
				"version":     schema.StringAttribute{},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *versionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data versionsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Info(ctx, fmt.Sprintf("read versions data-source request: package=%s", data.Package))

	f := &registry.PackageVersionMetadataRequest{
		Package: data.Package.String(),
	}
	packageVersionMetadata, err := d.prov.client.Registry().Registry().GetPackageVersionMetadata(ctx, f)
	if err != nil {
		resp.Diagnostics.Append(errorToDiagnostic(err, "failed to get package version metadata"))
		return
	}

	// TODO: hmmm
	m, _ := types.ObjectValueFrom(ctx, nil, &packageVersionMetadata)
	data.Metadata = m

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
