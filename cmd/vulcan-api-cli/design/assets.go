/*
Copyright 2021 Adevinta
*/

package design

import (
	// Use . imports to enable the DSL
	. "github.com/goadesign/goa/design"        // nolint
	. "github.com/goadesign/goa/design/apidsl" // nolint
)

var AssetMedia = MediaType("asset", func() {
	Description("Asset")
	Attributes(func() {
		Attribute("id", String, "Asset ID", func() { Example("a8720503-0284-45fd-9cf4-5bb6c500966f") })
		Attribute("type", AssetTypeMedia, "Type")
		Attribute("identifier", String, "Identifier", func() { Example("vulcan.example.com") })
		Attribute("options", String, "Options", func() { Example("{\"timeout\":60}") })
		Attribute("environmental_cvss", String, "Environmental CVSS", func() { Example("AV:N/AC:H/PR:N/UI:R/S:U/C:H/I:H/A:H") })
		Attribute("rolfp", String, "Rolfp plus scope vector", func() { Example("R:1/O:1/L:0/F:0/P:0+S:1") })
		Attribute("scannable", Boolean, "Scannable", func() { Example(true) })
		Attribute("classified_at", String, "Classified At", func() { Example("2020-09-03T15:00:42.112975Z") })
		Attribute("alias", String, "Alias", func() { Example("AnAlias") })
		Attribute("annotations", HashOf(String, String), func() {
			Example(map[string]string{
				"annotation/1": "value/1",
				"annotation/2": "value/2",
			})
		})
	})
	View("default", func() {
		Attribute("id")
		Attribute("type")
		Attribute("identifier")
		Attribute("options")
		Attribute("environmental_cvss")
		Attribute("rolfp")
		Attribute("alias")
		Attribute("scannable")
		Attribute("classified_at")
		Attribute("annotations")
	})
})

var ListAssetMedia = MediaType("ListAssetEntry", func() {
	Description("List Asset Entry")
	Attributes(func() {
		Attribute("id", String, "Asset ID", func() { Example("a8720503-0284-45fd-9cf4-5bb6c500966f") })
		Attribute("type", AssetTypeMedia, "Type")
		Attribute("identifier", String, "Identifier", func() { Example("vulcan.example.com") })
		Attribute("options", String, "Options", func() { Example("{\"timeout\":60}") })
		Attribute("environmental_cvss", String, "Environmental CVSS", func() { Example("AV:N/AC:H/PR:N/UI:R/S:U/C:H/I:H/A:H") })
		Attribute("rolfp", String, "Rolfp plus scope vector", func() { Example("R:1/O:1/L:0/F:0/P:0+S:1") })
		Attribute("scannable", Boolean, "Scannable", func() { Example(true) })
		Attribute("classified_at", String, "Classified At", func() { Example("2020-09-03T15:00:42.112975Z") })
		Attribute("alias", String, "Alias", func() { Example("AnAlias") })
		Attribute("groups", CollectionOf(GroupMedia), "Groups")
		Attribute("annotations", HashOf(String, String), func() {
			Example(map[string]string{
				"annotation/1": "value/1",
				"annotation/2": "value/2",
			})
		})
	})
	View("default", func() {
		Attribute("id")
		Attribute("type")
		Attribute("identifier")
		Attribute("options")
		Attribute("environmental_cvss")
		Attribute("rolfp")
		Attribute("alias")
		Attribute("scannable")
		Attribute("classified_at")
		Attribute("groups", func() {
			View("WithoutAssetsCount")
		})
		Attribute("annotations")
	})
})

var AssetResponseMedia = MediaType("assetResponse", func() {
	Description("Asset")
	Attributes(func() {
		Attribute("id", String, "Asset ID", func() { Example("a8720503-0284-45fd-9cf4-5bb6c500966f") })
		Attribute("type", AssetTypeMedia, "Type")
		Attribute("identifier", String, "Identifier", func() { Example("vulcan.example.com") })
		Attribute("options", String, "Options", func() { Example("{\"timeout\":60}") })
		Attribute("environmental_cvss", String, "Environmental CVSS", func() { Example("AV:N/AC:H/PR:N/UI:R/S:U/C:H/I:H/A:H") })
		Attribute("rolfp", String, "Rolfp plus scope vector", func() { Example("R:1/O:1/L:0/F:0/P:0+S:1") })
		Attribute("scannable", Boolean, "Scannable", func() { Example(true) })
		Attribute("classified_at", String, "Classified At", func() { Example("2020-09-03T15:00:42.112975Z") })
		Attribute("alias", String, "Alias", func() { Example("AnAlias") })
		Attribute("status", APIErrorMedia, "Status")
	})
	View("default", func() {
		Attribute("id")
		Attribute("type")
		Attribute("identifier")
		Attribute("options")
		Attribute("environmental_cvss")
		Attribute("rolfp")
		Attribute("scannable")
		Attribute("classified_at")
		Attribute("alias")
		Attribute("status")
	})
})

var AssetTypeMedia = MediaType("assetType", func() {
	Description("Asset Type")
	Attributes(func() {
		Attribute("id", String, "Asset Type ID", func() { Example("29f83d16-09c4-4be5-8cae-e232d41f388a") })
		Attribute("name", String, "Name", func() { Example("Hostname") })
	})
	View("default", func() {
		Attribute("id")
		Attribute("name")
	})
})

var AssetPayload = Type("AssetPayload", func() {
	Attribute("type", String, "Type", func() { Example("Hostname") })
	Attribute("identifier", String, "Identifier", func() { Example("vulcan.example.com") })
	Attribute("options", String, "Options", func() { Example("{\"timeout\":60}") })
	Attribute("environmental_cvss", String, "Environmental CVSS", func() { Example("AV:N/AC:H/PR:N/UI:R/S:U/C:H/I:H/A:H") })
	Attribute("rolfp", String, "Rolfp plus scope vector", func() { Example("R:1/O:1/L:0/F:0/P:0+S:1") })
	Attribute("scannable", Boolean, "Scannable", func() {
		Example(true)
	})
	Attribute("alias", String, "The alias of the asset in Vulcan", func() { Example("AnAlias") })
	Required("identifier")
})

var AssetWithAnnotationsPayload = Type("AssetWithAnnotationsPayload", func() {
	Attribute("type", String, "Type", func() { Example("Hostname") })
	Attribute("identifier", String, "Identifier", func() { Example("vulcan.example.com") })
	Attribute("options", String, "Options", func() { Example("{\"timeout\":60}") })
	Attribute("environmental_cvss", String, "Environmental CVSS", func() { Example("AV:N/AC:H/PR:N/UI:R/S:U/C:H/I:H/A:H") })
	Attribute("rolfp", String, "Rolfp plus scope vector", func() { Example("R:1/O:1/L:0/F:0/P:0+S:1") })
	Attribute("scannable", Boolean, "Scannable", func() {
		Example(true)
	})
	Attribute("alias", String, "The alias of the asset in Vulcan", func() { Example("AnAlias") })
	Attribute("annotations", HashOf(String, String), func() {
		Description(`The provided annotations may differ from the ones that
will be stored, because they will include a prefix to not mess with any other
annotations already present in the asset.`)
		Example(map[string]string{
			"annotation/1": "value/1",
			"annotation/2": "value/2",
		})
	})
	Required("identifier")
	Required("type")
})

var AssetUpdatePayload = Type("AssetUpdatePayload", func() {
	Attribute("type", String, "Type", func() { Example("Hostname") })
	Attribute("identifier", String, "Identifier", func() { Example("vulcan.example.com") })
	Attribute("options", String, "Options", func() { Example("{\"timeout\":60}") })
	Attribute("environmental_cvss", String, "Environmental CVSS", func() { Example("AV:N/AC:H/PR:N/UI:R/S:U/C:H/I:H/A:H") })
	Attribute("rolfp", String, "Rolfp plus scope vector", func() { Example("R:1/O:1/L:0/F:0/P:0+S:1") })
	Attribute("scannable", Boolean, "Scannable", func() {
		Example(true)
	})
	Attribute("alias", String, "Alias", func() { Example("AnAlias") })
})

var CreateAssetsMedia = MediaType("create_assets", func() {
	Description("Create Assets")
	Attributes(func() {
		Attribute("assets", CollectionOf(AssetMedia), "Assets")
		Attribute("errors", CollectionOf(AssetErrorMedia), "Errors")
	})
	View("default", func() {
		Attribute("assets")
		Attribute("errors")
	})
})

var AssetErrorMedia = MediaType("asseterror", func() {
	Description("Create Assets Errors")
	Attributes(func() {
		Attribute("id", Integer, "ID", func() { Example(0) })
		Attribute("error", String, "Error", func() { Example("Invalid asset type: FooBar") })
	})
	View("default", func() {
		Attribute("id")
		Attribute("error")
	})
})

var CreateAssetPayload = Type("CreateAssetPayload", func() {
	Attribute("assets", ArrayOf(AssetPayload))
	Attribute("groups", ArrayOf(String), func() {
		Example([]string{
			"a14c7c65-66ab-4676-bcf6-0dea9719f5c6",
			"9f7a0c78-b752-4126-aa6d-0f286ada7b8f",
		})
	})
	Attribute("annotations", HashOf(String, String), func() {
		Example(map[string]string{
			"annotation/1": "value/1",
			"annotation/2": "value/2",
		})
	})
	Required("assets")
})

var DiscoveredAssetsPayload = Type("DiscoveredAssetsPayload", func() {
	Attribute("assets", ArrayOf(AssetWithAnnotationsPayload))
	Attribute("group_name", String, func() {
		Description(`The discovery group name where assets will be added. It
		must end with '-discovered-assets'. The first part of the name should
		identify the discovery service using the endpoint`)
		Example("discoveryserviceX-discovered-assets")
	})
	Required("group_name")
})

var _ = Resource("assets", func() {
	Parent("teams")
	BasePath("assets")
	Params(func() {
		Param("team_id", String, "Team ID")
	})

	DefaultMedia(AssetMedia)

	Action("list", func() {
		Description("List all assets from a team.")
		Routing(GET(""))
		Params(func() {
			Param("identifier", String)
		})
		Security("Bearer")
		Response(OK, CollectionOf(ListAssetMedia))
	})

	Action("create", func() {
		Description(`Creates assets in bulk mode.
			This operation accepts an array of assets, an optional array of group identifiers, an optional map of annotations, and returns an array of successfully created assets.
			If no groups are specified, assets will be added to the team's Default group.
			If one of the specified assets already exists for the team but is currently not associated with the requested groups, the association is created.
			If for any reason, the creation of an asset fails, an error message will be returned referencing the failed asset and the entire operation will be rolled back.
			---
			Valid asset types:
			- AWSAccount
			- DomainName
			- Hostname
			- IP
			- IPRange
			- DockerImage
			- WebAddress
			- GitRepository
			---
			If the asset type is informed, then Vulcan will use that value to create the new asset.
			Otherwise, Vulcan will try to automatically discover the asset type.
			Notice that this may result in Vulcan creating more than one asset.
			For instance, an user trying to create an asset for "vulcan.example.com", without specifying the asset type, will end up with two assets created:
			- vulcan.example.com (DomainName) and
			- vulcan.example.com (Hostname).`)
		Routing(POST(""))
		Payload(CreateAssetPayload)
		Security("Bearer")
		Response(Created, CollectionOf(AssetMedia))
	})

	Action("createMultiStatus", func() {
		Description(`Creates assets in bulk mode (MultiStatus).
			This operation is similar to the "Create Assets in Bulk Mode", with 2 main differences:
			- This endpoint is not atomic. Each asset creation request will succeed or fail indenpendently of the other requests.
			- This endpoint will return an array of AssetResponse in the following way:
				· For each asset with specified type, returns an AssetResponse indicating the success or failure for its creation.
				· For each asset with no type specified and successfully created, returns one AssetResponse for each auto detected asset.
				· For each asset detected from the ones with no type indicated which their creation produced an error, returns one AssetResponse indicating the failure for its creation specifying its detected type.
			In the case of all assets being successfully created, this endpoint will return status code 201-Created. 
			Otherwise, it will return a 207-MultiStatus code, indicating that at least one of the requested operations failed.	
		`)
		Routing(POST("multistatus"))
		Payload(CreateAssetPayload)
		Security("Bearer")
		Response(Created, CollectionOf(AssetResponseMedia), func() {
			Description("Created: All assets were created with success.")
		})
		Response("MultiStatus", CollectionOf(AssetResponseMedia), func() {
			Description("Multiple Status: At least one of the assets failed to be created.")
			Status(207)
		})
	})

	Action("show", func() {
		Description("Describe an asset.")
		Routing(GET("/:asset_id"))
		Params(func() {
			Param("asset_id", String, "Asset ID")
		})
		Security("Bearer")
		Response(OK, AssetMedia)
	})

	Action("update", func() {
		Description("Update an asset.")
		Routing(PATCH("/:asset_id"))
		Params(func() {
			Param("asset_id", String, "Asset ID")
		})
		Payload(AssetUpdatePayload)
		Security("Bearer")
		Response(OK, AssetMedia)
	})

	Action("delete", func() {
		Description("Delete an asset.")
		Routing(DELETE("/:asset_id"))
		Params(func() {
			Param("asset_id", String, "Asset ID")
		})
		Security("Bearer")
		Response(NoContent, func() {})
	})

	Action("discover", func() {
		Description(`This endpoint receives a list of assets with embedded
asset annotations, and the group name where to be added. It should be used by
third-party asset discovery services to onboard the discovered assets into
Vulcan. The provided list of assets will overwrite the assets previously
present in the group, in a way that:
  - Assets that do not exist in the team will be created and associated to the
  group
  - Assets that were already existing in the team but not associated to the
  group will be associated
  - Existing assets where the scannable field or the annotations are different
  will be updated accordingly
  - Assets that were associated to the group and now are not present in the
  provided list will be de-associated from the group if they belong to any
  other group, or deleted otherwise
Because of the latency of this operation the endpoint is asynchronous. It
returns a 202-Accepted HTTP response with the Job information in the response
body.

The discovery group name must end with '-discovered-assets' to not mess with
manually managed asset groups. Also the first part of the name should identify
the discovery service using the endpoint, for example:
serviceX-discovered-assets.
Also be aware that the provided annotations may differ from the ones that will
be stored, because they will include a prefix to not mess with any other
annotations already present in the asset.

Duplicated assets (same identifier and type) in the payload do not produce an
error but only the first one will be taken into account.`)
		Routing(PUT("discovery"))
		Payload(DiscoveredAssetsPayload)
		Security("Bearer")
		Response(Accepted, JobMedia, func() {
			Description("Created: All assets were created with success.")
		})
	})
})
