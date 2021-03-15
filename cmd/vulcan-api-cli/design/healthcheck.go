/*
Copyright 2021 Adevinta
*/

package design

import (
	// Use . imports to enable the DSL
	. "github.com/goadesign/goa/design"        // nolint
	. "github.com/goadesign/goa/design/apidsl" // nolint
)

var _ = Resource("healthcheck", func() {
	BasePath("/healthcheck")

	Action("show", func() {
		Description("A simple HTTP healthcheck.")
		Routing(GET(""))
		NoSecurity()
		Response(OK, HealthcheckMedia)
	})
})

var HealthcheckMedia = MediaType("healthcheck", func() {
	Description("Healthcheck")
	Attributes(func() {
		Attribute("status", String, "Status", func() { Example("OK") })
	})
	View("default", func() {
		Attribute("status")
	})
})
