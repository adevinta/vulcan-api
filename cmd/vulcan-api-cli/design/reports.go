/*
Copyright 2021 Adevinta
*/

package design

import (
	// Use . imports to enable the DSL
	. "github.com/goadesign/goa/design"        // nolint
	. "github.com/goadesign/goa/design/apidsl" // nolint
)

var DigestPayload = Type("DigestPayload", func() {
	Attribute("start_date", String, "Start Date", func() { Example("2020-08-11") })
	Attribute("end_date", String, "End Date", func() { Example("2020-09-10") })
})

var _ = Resource("report", func() {
	Parent("teams")
	BasePath("report")

	Action("send digest", func() {
		Description("Send digest report.\nIf no dates are specified, the time range will be set for the last 30 days.")
		Routing(POST("/digest"))
		Payload(DigestPayload)
		Security("Bearer")
		Response(OK)
	})
})
