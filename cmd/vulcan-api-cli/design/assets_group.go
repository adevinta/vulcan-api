/*
Copyright 2021 Adevinta
*/

package design

import (
	// Use . imports to enable the DSL
	. "github.com/goadesign/goa/design"        // nolint
	. "github.com/goadesign/goa/design/apidsl" // nolint
)

var _ = Resource("asset-group", func() {
	Parent("teams")
	BasePath("/groups")
	DefaultMedia(GroupMedia)

	Action("list", func() {
		Description("List all assets from a group.")
		Routing(GET("/:group_id/assets"))
		Params(func() {
			Param("group_id", String, "Group ID")
		})
		Security("Bearer")
		Response(OK, CollectionOf(AssetMedia))
	})

	Action("create", func() {
		Description("Associate an asset to a group.")
		Routing(POST("/:group_id/assets"))
		Params(func() {
			Param("group_id", String, "Group ID")
		})
		Payload(AssetGroupPayload)
		Security("Bearer")
		Response(Created, AssetMedia)
	})

	Action("delete", func() {
		Description("Remove an asset from a group.")
		Routing(DELETE("/:group_id/assets/:asset_id"))
		Params(func() {
			Param("group_id", String, "Group ID")
			Param("asset_id", String, "Asset ID")
		})
		Security("Bearer")
		Response(NoContent, func() {})
	})
})
