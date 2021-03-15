/*
Copyright 2021 Adevinta
*/

package design

import (
	// Use . imports to enable the DSL
	. "github.com/goadesign/goa/design"        // nolint
	. "github.com/goadesign/goa/design/apidsl" // nolint
)

var GroupMedia = MediaType("group", func() {
	Description("Group")
	Attributes(func() {
		Attribute("id", String, "Group ID", func() { Example("f6360346-77a5-4f97-b919-331363b3af26") })
		Attribute("name", String, "Name", func() { Example("Default group") })
		Attribute("description", String, "Description", func() { Example("All newly created assets are added to the Default group") })
		Attribute("options", String, "Options", func() { Example("{\"timeout\":60}") })
		Attribute("assets_count", Integer, "Assets Count", func() { Example(1) })
	})
	View("default", func() {
		Attribute("id")
		Attribute("name")
		Attribute("description")
		Attribute("options")
		Attribute("assets_count")
	})
	View("WithoutAssetsCount", func() {
		Attribute("id")
		Attribute("name")
		Attribute("description")
		Attribute("options")
	})
})

var GroupPayload = Type("GroupPayload", func() {
	Attribute("name", String, "name", func() { Example("Default group") })
	Attribute("options", String, "Options", func() { Example("{\"timeout\":60}") })
	Required("name")
})

var AssetGroupMedia = MediaType("assetGroup", func() {
	Description("Asset group")
	Attributes(func() {
		Attribute("asset", AssetMedia, "Asset")
		Attribute("group", GroupMedia, "Group")
	})
	View("default", func() {
		Attribute("asset")
		Attribute("group")
	})
})

var ListAssetGroupMedia = MediaType("listAssetGroup", func() {
	Description("List asset group")
	Attributes(func() {
		Attribute("assets", CollectionOf(AssetMedia), "Asset")
		Attribute("group", GroupMedia, "Group")
	})
	View("default", func() {
		Attribute("assets")
		Attribute("group")
	})
})

var AssetGroupPayload = Type("AssetGroupPayload", func() {
	Attribute("asset_id", String, "Asset ID", func() { Example("0fc67150-5cd9-486a-aca5-9c9167478e4d") })
	Required("asset_id")
})

var _ = Resource("group", func() {
	Parent("teams")
	BasePath("/groups")
	DefaultMedia(GroupMedia)

	Action("list", func() {
		Description("List all groups of assets from a team.")
		Routing(GET(""))
		Security("Bearer")
		Response(OK, CollectionOf(GroupMedia))
	})

	Action("create", func() {
		Description("Create a new group of assets.")
		Routing(POST(""))
		Payload(GroupPayload)
		Security("Bearer")
		Response(Created, GroupMedia)
	})

	Action("show", func() {
		Description("Describe a group of assets.")
		Routing(GET("/:group_id"))
		Params(func() {
			Param("group_id", String, "Group ID")
		})
		Security("Bearer")
		Response(OK, GroupMedia)
	})

	Action("update", func() {
		Description("Update a group of assets.")
		Routing(PATCH("/:group_id"))
		Params(func() {
			Param("group_id", String, "Group ID")
		})
		Payload(GroupPayload)
		Security("Bearer")
		Response(OK, GroupMedia)
	})

	Action("delete", func() {
		Description("Delete a group of assets.")
		Routing(DELETE("/:group_id"))
		Params(func() {
			Param("group_id", String, "Group ID")
		})
		Security("Bearer")
		Response(NoContent, func() {})
	})
})
