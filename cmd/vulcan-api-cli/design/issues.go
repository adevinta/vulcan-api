/*
Copyright 2021 Adevinta
*/

package design

import (
	// Use . imports to enable the DSL
	. "github.com/goadesign/goa/design"        // nolint
	. "github.com/goadesign/goa/design/apidsl" // nolint
)

var _ = Resource("issues", func() {
	BasePath("/issues")
	DefaultMedia(IssueMedia)

	Action("list", func() {
		Description(`List all the issues.`)
		Routing(GET("/"))
		Params(func() {
			Param("page", Number, "Requested page")
			Param("size", Number, "Requested page size")
		})
		Security("Bearer")
		Response(OK, IssuesListMedia)
	})
})

var IssuesListMedia = MediaType("issues_list", func() {
	Description("Issue list")
	Attributes(func() {
		Attribute("issues", CollectionOf(IssueMedia), "List of issues")
		Attribute("pagination", PaginationMedia, "Pagination info")
	})
	View("default", func() {
		Attribute("issues")
		Attribute("pagination")
	})
})
