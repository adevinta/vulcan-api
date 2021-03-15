/*
Copyright 2021 Adevinta
*/

package design

import (
	// Use . imports to enable the DSL
	. "github.com/goadesign/goa/design"        // nolint
	. "github.com/goadesign/goa/design/apidsl" // nolint
)

var ScheduleMedia = MediaType("schedule", func() { //nolint
	Description("Schedule")
	Attributes(func() {
		Attribute("cron", String, "Cron Expression", func() { Example("0 7 1 * *") })
	})
	Required("cron")
	View("default", func() {
		Attribute("cron")
	})
})

var SchedulePayload = Type("SchedulePayload", func() {
	Attribute("cron", String, "Cron Expression", func() { Example("0 7 1 * *") })
})

var ScheduleUpdatePayload = Type("ScheduleUpdatePayload", func() {
	Attribute("cron", String, "Cron Expression", func() { Example("0 7 1 * *") })
})

var _ = Resource("schedule", func() {
	Parent("programs")
	BasePath("schedule")
	DefaultMedia(ScheduleMedia)

	Action("create", func() {
		Description("Create a new schedule.")
		Routing(POST(""))
		Params(func() {
			Param("team_id", String, "Team ID")
			Param("program_id", String, "Program ID")
		})
		Payload(SchedulePayload)
		Security("Bearer")
		Response(Created, ScheduleMedia)
	})

	Action("update", func() {
		Description("Update information about a schedule.")
		Routing(POST(""))
		Params(func() {
			Param("team_id", String, "Team ID")
			Param("program_id", String, "Program ID")
		})
		Payload(ScheduleUpdatePayload)
		Security("Bearer")
		Response(OK, ScheduleMedia)
	})

	Action("delete", func() {
		Description("Delete a schedule.")
		Routing(DELETE(""))
		Params(func() {
			Param("team_id", String, "Team ID")
			Param("program_id", String, "Program ID")
		})
		Security("Bearer")
		Response(NoContent, func() {})
	})
})
