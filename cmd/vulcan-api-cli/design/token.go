/*
Copyright 2021 Adevinta
*/

package design

import (
	// Use . imports to enable the DSL
	. "github.com/goadesign/goa/design"        // nolint
	. "github.com/goadesign/goa/design/apidsl" // nolint
)

var TokenMedia = MediaType("token", func() {
	Description("Token")
	Attributes(func() {
		Attribute("email", String, "Email", func() { Example("john.doe@vulcan.com") })
		Attribute("hash", String, "Hash", func() { Example("903af1a77ea4eda46b60d85fac0f312ff9d3b6ea092434c8a6ef1868bf95da77") })
		Attribute("creation_time", String, "Creation time", func() { Example("2018-09-07T10:40:52Z") })
		Attribute("token", String, "Token", func() {
			Example("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpYXQiOjE1MzcxODQ0NjksInN1YiI6Im1haWNvbkBjb3N0YS5jb20iLCJ0eXBlIjoiQVBJIn0.zu8mf5nSpvmFyZNZM7Im-omHa_J9Ck1KH49zuy1wvjY")
		})
	})
	View("default", func() {
		Attribute("email")
		Attribute("hash")
		Attribute("creation_time")
		Attribute("token")
	})
	View("metadata", func() {
		Attribute("email")
		Attribute("hash")
		Attribute("creation_time")
	})
})

var _ = Resource("api-token", func() {
	BasePath("/token")
	Parent("user")
	DefaultMedia(TokenMedia)

	Action("create", func() {
		Description("Generate an API token for an user.")
		Routing(POST(""))
		Security("Bearer")
		Response(Created, TokenMedia)
	})
})
