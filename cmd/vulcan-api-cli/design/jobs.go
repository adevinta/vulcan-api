/*
Copyright 2021 Adevinta
*/

package design

import (
	// Use . imports to enable the DSL
	. "github.com/goadesign/goa/design"        // nolint
	. "github.com/goadesign/goa/design/apidsl" // nolint
)

var JobMedia = MediaType("job", func() {
	Description("Job")
	Attributes(func() {
		Attribute("id", String, "Job ID", func() { Example("967d9966-b561-4233-bd6f-cac603fd8320") })
		Attribute("team_id", String, "Team ID", func() { Example("9cb0bb2b-ca36-4877-acad-9dde23880595") })
		Attribute("operation", String, "Operation", func() { Example("OnboardDiscoveredAssets") })
		Attribute("status", String, "Status", func() { Example("PROCESSING") })
	})
	View("default", func() {
		Attribute("id")
		Attribute("team_id")
		Attribute("operation")
		Attribute("status")
	})
})

var _ = Resource("job", func() {
	BasePath("/jobs")
	DefaultMedia(UserMedia)

	Action("show", func() {
		Description("Describe job")
		Routing(GET("/:job_id"))
		Params(func() {
			Param("job_id", String, "Job ID")
		})
		Security("Bearer")
		Response(OK, JobMedia)
	})
})
