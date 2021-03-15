/*
Copyright 2021 Adevinta
*/

package design

import (
	// Use . imports to enable the DSL
	. "github.com/goadesign/goa/design"        // nolint
	. "github.com/goadesign/goa/design/apidsl" // nolint
)

var ProgramMedia = MediaType("program", func() {
	Description("Program")
	Attributes(func() {
		Attribute("id", String, "Program ID", func() { Example("9a100645-e51a-4e71-8a4f-1e462ce9a40d") })
		Attribute("policy_groups", ArrayOf(ProgramPolicyGroupMedia), "PolicyGroup")
		Attribute("name", String, "Name", func() { Example("Every midnight") })
		Attribute("global", Boolean, "Global")
		Attribute("schedule", ScheduleMedia, "Schedule")
		Attribute("autosend", Boolean, "Autosend")
		Attribute("disabled", Boolean, "Disabled")
	})
	View("default", func() {
		Attribute("id")
		Attribute("policy_groups")
		Attribute("name")
		Attribute("global")
		Attribute("schedule")
		Attribute("autosend")
		Attribute("disabled")
	})
})

var ProgramPolicyGroupMedia = MediaType("program_policy_group", func() {
	Attributes(func() {
		Attribute("group", GroupMedia, "group")
		Attribute("policy", PolicyMedia, "policy")
	})
	View("default", func() {
		Attribute("group")
		Attribute("policy")
	})
})

var ProgramPayload = Type("ProgramPayload", func() {
	Attribute("policy_groups", ArrayOf(ProgramPolicyGroupPayload), "PolicyGroups")
	Attribute("name", String, "name", func() { Example("Every midnight") })
	Attribute("autosend", Boolean, "Autosend")
	Attribute("disabled", Boolean, "Disabled")
})

var ProgramUpdatePayload = Type("ProgramUpdatePayload", func() {
	Attribute("policy_groups", ArrayOf(ProgramPolicyGroupPayload), "PolicyGroups")
	Attribute("name", String, "name", func() { Example("Every midnight") })
	Attribute("autosend", Boolean, "Autosend")
	Attribute("disabled", Boolean, "Disabled")
})

var ProgramPolicyGroupPayload = Type("ProgramPolicyGroupPayload", func() {
	Attribute("group_id", String, "group")
	Attribute("policy_id", String, "policy")
})

var _ = Resource("programs", func() {
	Parent("teams")
	BasePath("programs")
	DefaultMedia(ProgramMedia)

	Action("list", func() {
		Description("List all programs from a team.")
		Routing(GET(""))
		Security("Bearer")
		Response(OK, CollectionOf(ProgramMedia))
	})

	Action("create", func() {
		Description("Create a new program.")
		Routing(POST(""))
		Payload(ProgramPayload)
		Security("Bearer")
		Response(Created, ProgramMedia)
	})

	Action("show", func() {
		Description("Show information about a program.")
		Routing(GET("/:program_id"))
		Params(func() {
			Param("program_id", String, "Program ID")
		})
		Security("Bearer")
		Response(OK, ProgramMedia)
	})

	Action("update", func() {
		Description("Update information about a program.")
		Routing(PATCH("/:program_id"))
		Params(func() {
			Param("program_id", String, "Program ID")
		})
		Payload(ProgramUpdatePayload)
		Security("Bearer")
		Response(OK, ProgramMedia)
	})

	Action("delete", func() {
		Description("Delete a program.")
		Routing(DELETE("/:program_id"))
		Params(func() {
			Param("program_id", String, "Program ID")
		})
		Security("Bearer")
		Response(NoContent, func() {})
	})
})

var _ = Resource("program-scans", func() {
	Parent("programs")
	BasePath("/scans")
	DefaultMedia(ScanMedia)
	Action("list", func() {
		Description("List the scans of a program.")
		Routing(GET(""))
		Security("Bearer")
		Response(OK, CollectionOf(ScanMedia))
	})
})
