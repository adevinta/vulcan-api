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
		Attribute("identifier", String, "Identifier", func() { Example("vulcan.com") })
		Attribute("options", String, "Options", func() { Example("{\"timeout\":60}") })
		Attribute("environmental_cvss", String, "Environmental CVSS", func() { Example("AV:N/AC:H/PR:N/UI:R/S:U/C:H/I:H/A:H") })
		Attribute("rolfp", String, "Rolfp plus scope vector", func() { Example("R:1/O:1/L:0/F:0/P:0+S:1") })
		Attribute("scannable", Boolean, "Scannable", func() { Example(true) })
		Attribute("classified_at", String, "Classified At", func() { Example("2020-09-03T15:00:42.112975Z") })
		Attribute("alias", String, "Alias", func() { Example("AnAlias") })
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
	})
})

var ListAssetMedia = MediaType("ListAssetEntry", func() {
	Description("List Asset Entry")
	Attributes(func() {
		Attribute("id", String, "Asset ID", func() { Example("a8720503-0284-45fd-9cf4-5bb6c500966f") })
		Attribute("type", AssetTypeMedia, "Type")
		Attribute("identifier", String, "Identifier", func() { Example("vulcan.com") })
		Attribute("options", String, "Options", func() { Example("{\"timeout\":60}") })
		Attribute("environmental_cvss", String, "Environmental CVSS", func() { Example("AV:N/AC:H/PR:N/UI:R/S:U/C:H/I:H/A:H") })
		Attribute("rolfp", String, "Rolfp plus scope vector", func() { Example("R:1/O:1/L:0/F:0/P:0+S:1") })
		Attribute("scannable", Boolean, "Scannable", func() { Example(true) })
		Attribute("classified_at", String, "Classified At", func() { Example("2020-09-03T15:00:42.112975Z") })
		Attribute("alias", String, "Alias", func() { Example("AnAlias") })
		Attribute("groups", CollectionOf(GroupMedia), "Groups")
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
	})
})

var AssetResponseMedia = MediaType("assetResponse", func() {
	Description("Asset")
	Attributes(func() {
		Attribute("id", String, "Asset ID", func() { Example("a8720503-0284-45fd-9cf4-5bb6c500966f") })
		Attribute("type", AssetTypeMedia, "Type")
		Attribute("identifier", String, "Identifier", func() { Example("vulcan.com") })
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
	Attribute("identifier", String, "Identifier", func() { Example("vulcan.com") })
	Attribute("options", String, "Options", func() { Example("{\"timeout\":60}") })
	Attribute("environmental_cvss", String, "Environmental CVSS", func() { Example("AV:N/AC:H/PR:N/UI:R/S:U/C:H/I:H/A:H") })
	Attribute("rolfp", String, "Rolfp plus scope vector", func() { Example("R:1/O:1/L:0/F:0/P:0+S:1") })
	Attribute("scannable", Boolean, "Scannable", func() {
		Example(true)
	})
	Attribute("alias", String, "The alias of the asset in Vulcan", func() { Example("AnAlias") })
	Required("identifier")
})

var AssetUpdatePayload = Type("AssetUpdatePayload", func() {
	Attribute("type", String, "Type", func() { Example("Hostname") })
	Attribute("identifier", String, "Identifier", func() { Example("vulcan.com") })
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
	Required("assets")
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
		Security("Bearer")
		Response(OK, CollectionOf(ListAssetMedia))
	})

	Action("create", func() {
		Description(`Creates assets in bulk mode.
			This operation accepts an array of assets and an optional array of group identifiers, and returns an array of successfully created assets.
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
			For instance, an user trying to create an asset for "vulcan.com", without specifying the asset type, will end up with two assets created:
			- vulcan.com (DomainName) and
			- vulcan.com (Hostname).`)
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
				· For each asset with no type specified which its creation produced an error, returns one AssetResponse indicating the failure for the creation of its detected assets without specifying which exact type failed.
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
})
