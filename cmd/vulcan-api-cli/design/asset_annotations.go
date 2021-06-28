/*
Copyright 2021 Adevinta
*/

package design

import (
	// Use . imports to enable the DSL
	. "github.com/goadesign/goa/design"        // nolint
	. "github.com/goadesign/goa/design/apidsl" // nolint
)

var AssetAnnotationRequest = Type("AssetAnnotationRequest", func() {
	Attribute("annotations", HashOf(String, String), func() {
		Example(map[string]string{
			"annotation/1": "value/1",
			"annotation/2": "value/2",
		})
	})
})

var AssetAnnotationDeleteRequest = Type("AssetAnnotationDeleteRequest", func() {
	Attribute("annotations", ArrayOf(String), func() {
		Example([]string{
			"annotation/1",
			"annotation/2",
		})
	})
})

var AssetAnnotationsResponseExample = MediaType("assetannotations_response_example", func() {
	Description("Asset Annotations")
	Attributes(func() {
		Attribute("annotation/1", String, func() { Example("value/1") })
		Attribute("annotation/2", String, func() { Example("value/2") })
	})
	View("default", func() {
		Attribute("annotation/1")
		Attribute("annotation/2")
	})
})

var _ = Resource("asset-annotations", func() {
	Parent("assets")
	BasePath("annotations")
	Params(func() {
		Param("team_id", String, "Team ID")
		Param("asset_id", String, "Asset ID")
	})
	Action("list", func() {
		Description("List annotations of a given asset.")
		Routing(GET(""))
		Security("Bearer")
		Response(OK, AssetAnnotationsResponseExample)
	})
	Action("create", func() {
		Description("Create one or more annotation for a given asset.")
		Routing(POST(""))
		Payload(AssetAnnotationRequest)
		Security("Bearer")
		Response(Created, AssetAnnotationsResponseExample)
	})
	Action("update", func() {
		Description("Update one or more annotation for a given asset.")
		Routing(PATCH(""))
		Payload(AssetAnnotationRequest)
		Security("Bearer")
		Response(OK, AssetAnnotationsResponseExample)
	})
	Action("delete", func() {
		Description("Delete one or more annotation for a given asset.")
		Routing(DELETE(""))
		Payload(AssetAnnotationDeleteRequest)
		Security("Bearer")
		Response(OK, AssetAnnotationsResponseExample)
	})
})
