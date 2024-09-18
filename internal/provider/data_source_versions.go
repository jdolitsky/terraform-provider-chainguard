/*
Copyright 2024 Chainguard, Inc.
SPDX-License-Identifier: Apache-2.0
*/

package provider

import (
	"context"
	"encoding/json"
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
	Package     types.String `tfsdk:"package"`
	RawMetadata types.String `tfsdk:"raw_metadata"`
	//Metadata types.Object `tfsdk:"metadata"`
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
				Required:    true,
			},
			"raw_metadata": schema.StringAttribute{
				Computed: true,
			},
			/*
				"metadata": schema.SingleNestedAttribute{
					Computed: true,
					Attributes: map[string]schema.Attribute{
						"eol_versions": versionSchema(),
						"last_updated_timestamp": schema.StringAttribute{
							Required: true,
						},
						"latest_version": schema.StringAttribute{
							Required: true,
						},
						"versions": versionSchema(),
					},
				},
			*/
		},
	}
}

func versionSchema() schema.ListNestedAttribute {
	return schema.ListNestedAttribute{
		Required: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"eol_date": schema.StringAttribute{
					Required: true,
				},
				"exists": schema.BoolAttribute{
					Required: true,
				},
				"fips": schema.BoolAttribute{
					Required: true,
				},
				"lts": schema.StringAttribute{
					Required: true,
				},
				"release_date": schema.StringAttribute{
					Required: true,
				},
				"version": schema.StringAttribute{
					Required: true,
				},
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
		Package: data.Package.ValueString(),
	}
	packageVersionMetadata, err := d.prov.client.Registry().Registry().GetPackageVersionMetadata(ctx, f)
	if err != nil {
		// TODO: address var.variant == "fips"
		data.RawMetadata = types.StringValue(`{"lastUpdatedTimestamp": "", "latestVersion": "", "eolVersions": [], "versions": [{"version": "", "exists": true, "fips": false}]}`)
		resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
		//resp.Diagnostics.Append(errorToDiagnostic(err, "failed to get package version metadata"))
		return
	}

	b, err := json.Marshal(packageVersionMetadata)
	if err != nil {
		resp.Diagnostics.Append(errorToDiagnostic(err, "unable to convert package version metadata to JSON"))
		return
	}
	data.RawMetadata = types.StringValue(string(b))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	/*
		tmp := map[interface{}]interface{}{
			"metadata": map[string]string{
				"last_updated_timestamp": "abc",
			},
		}
		tmp := struct {
			last_updated_timestamp string
		}{
			last_updated_timestamp: "abc",
		}*/

	/*
		m, diag := types.ObjectValueFrom(ctx, data.Metadata.AttributeTypes(ctx), &tmp)
		if diag.HasError() {
			resp.Diagnostics.Append(diag.Errors()...)
			return
		}

		panic("GOT HERE")
		data.Metadata = m
	*/
}

/* terraform {
  required_providers {
    chainguard = { source = "chainguard-dev/chainguard" }
  }

  backend "inmem" {}
}

provider "chainguard" {
  console_api = "https://console-api.enforce.dev"
}

data "chainguard_versions" "versions" {
  package = "bazelx"
}

output "versions" {
  value = data.chainguard_versions.versions.raw_metadata
}
*/
