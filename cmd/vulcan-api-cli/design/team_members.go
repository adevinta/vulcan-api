/*
Copyright 2021 Adevinta
*/

package design

import (
	// Use . imports to enable the DSL
	. "github.com/goadesign/goa/design"        // nolint
	. "github.com/goadesign/goa/design/apidsl" // nolint
)

var TeamMemberMedia = MediaType("teamMember", func() {
	Description("Team Member")
	Attributes(func() {
		Attribute("user", UserMedia, "User")
		Attribute("role", String, "Role", func() { Example("owner") })
	})
	View("default", func() {
		Attribute("user")
		Attribute("role")
	})
})

var TeamMemberPayload = Type("TeamMemberPayload", func() {
	Attribute("user_id", String, "User ID", func() { Example("967d9966-b561-4233-bd6f-cac603fd8320") })
	Attribute("email", String, "Email", func() { Example("john.doe@vulcan.com") })
	Attribute("role", String, "Member role. Valid values are: owner, member", func() { Example("owner") })
})

var TeamMemberUpdatePayload = Type("TeamMemberUpdatePayload", func() {
	Attribute("role", String, "Member role. Valid values are: owner, member", func() { Example("owner") })
})

var _ = Resource("team-members", func() {
	Parent("teams")
	BasePath("/members")
	DefaultMedia(TeamMemberMedia)

	Action("list", func() {
		Description("List all members from a team.")
		Routing(GET(""))
		Security("Bearer")
		Response(OK, CollectionOf(TeamMemberMedia))
	})

	Action("create", func() {
		Description(`Create a team-member association.
			---
			At least one of the following fields must be specified: "email", "user_id".
			Otherwise the operation will fail.
			If an email is specified, but the user does not exists on the database yet, a new user will be created for that email.
			---
			Valid values for 'role' attribute:
			- member
			- owner`)
		Routing(POST(""))
		Payload(TeamMemberPayload)
		Security("Bearer")
		Response(Created, TeamMemberMedia)
	})

	Action("show", func() {
		Description("Describe a team-member association.")
		Routing(GET("/:user_id"))
		Params(func() {
			Param("user_id", String, "User ID")
		})
		Security("Bearer")
		Response(OK, TeamMemberMedia)
	})

	Action("update", func() {
		Description("Update a team-member association. \nValid values for 'role' attribute: 'member', 'owner'.")
		Routing(PATCH("/:user_id"))
		Params(func() {
			Param("user_id", String, "User ID")
		})
		Payload(TeamMemberUpdatePayload)
		Security("Bearer")
		Response(OK, TeamMemberMedia)
	})

	Action("delete", func() {
		Description("Delete a member from a team.")
		Routing(DELETE("/:user_id"))
		Params(func() {
			Param("user_id", String, "User ID")
		})
		Security("Bearer")
		Response(NoContent, func() {})
	})
})
