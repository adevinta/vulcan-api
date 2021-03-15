/*
Copyright 2021 Adevinta
*/

package design

import (
	// Use . imports to enable the DSL
	. "github.com/goadesign/goa/design"        // nolint
	. "github.com/goadesign/goa/design/apidsl" // nolint
)

var PolicyMedia = MediaType("policy", func() {
	Description("Policy")
	Attributes(func() {
		Attribute("id", String, "Policy ID", func() {
			Example("af36818a-0f30-412c-9692-d37716075861")
		})
		Attribute("name", String, "Name", func() { Example("Sample Policy") })
		Attribute("settings_count", Integer, "Policy settings count", func() { Example(12) })
	})
	View("default", func() {
		Attribute("id")
		Attribute("name")
		Attribute("settings_count")
	})
})

var PolicyPayload = Type("PolicyPayload", func() {
	Attribute("name", String, "name", func() { Example("Sample Policy") })
	Required("name")
})

var PolicyUpdatePayload = Type("PolicyUpdatePayload", func() {
	Attribute("name", String, "name", func() { Example("Sample Policy") })
})

var _ = Resource("policies", func() {
	Parent("teams")
	BasePath("/policies")
	DefaultMedia(PolicyMedia)

	Action("list", func() {
		Description("List all policies from a team.")
		Routing(GET(""))
		Security("Bearer")
		Response(OK, CollectionOf(PolicyMedia))
	})

	Action("create", func() {
		Description("Create a new policy.")
		Routing(POST(""))
		Payload(PolicyPayload)
		Security("Bearer")
		Response(Created, PolicyMedia)
	})

	Action("show", func() {
		Description("Show information about a policy.")
		Routing(GET("/:policy_id"))
		Params(func() {
			Param("policy_id", String, "Policy ID")
		})
		Security("Bearer")
		Response(OK, PolicyMedia)
	})

	Action("update", func() {
		Description("Update information about a policy.")
		Routing(PATCH("/:policy_id"))
		Params(func() {
			Param("policy_id", String, "Policy ID")
		})
		Payload(PolicyUpdatePayload)
		Security("Bearer")
		Response(OK, PolicyMedia)
	})

	Action("delete", func() {
		Description("Delete a policy.")
		Routing(DELETE("/:policy_id"))
		Params(func() {
			Param("policy_id", String, "Policy ID")
		})
		Security("Bearer")
		Response(NoContent, func() {})
	})
})
