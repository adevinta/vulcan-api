/*
Copyright 2021 Adevinta
*/

package design

import (
	// Use . imports to enable the DSL
	. "github.com/goadesign/goa/design"        // nolint
	. "github.com/goadesign/goa/design/apidsl" // nolint
)

var TeamMedia = MediaType("team", func() {
	Description("Team")
	Attributes(func() { // Attributes define the media type shape.
		Attribute("id", String, "Team ID", func() { Example("ff74a33c-e683-4612-a62f-a9cf385f6522") })
		Attribute("name", String, "Name", func() { Example("Security") })
		Attribute("description", String, "Description", func() { Example("Security Team") })
		Attribute("tag", String, "tag", func() { Example("team:security") })
	})
	View("default", func() {
		Attribute("id")
		Attribute("name")
		Attribute("description")
		Attribute("tag")
	})
})

var TeamPayload = Type("TeamPayload", func() {
	Attribute("name", String, "name", func() { Example("Security") })
	Attribute("description", String, "description", func() { Example("Security Team") })
	Attribute("tag", String, "tag", func() { Example("team:security") })
	Required("name")
	Required("tag")
})

var TeamUpdatePayload = Type("TeamUpdatePayload", func() {
	Attribute("name", String, "name", func() { Example("Security") })
	Attribute("description", String, "description", func() { Example("Security Team") })
	Attribute("tag", String, "tag", func() { Example("team:security") })
})

var _ = Resource("teams", func() {
	BasePath("/teams")
	DefaultMedia(TeamMedia)

	Action("list", func() {
		Description("List all teams in Vulcan.")
		Routing(GET(""))
		Params(func() {
			Param("tag", String, "Team tag")
		})
		Security("Bearer")
		Response(OK, CollectionOf(TeamMedia))
	})

	Action("create", func() {
		Description("Create a new team.")
		Routing(POST(""))
		Payload(TeamPayload)
		Security("Bearer")
		Response(Created, TeamMedia)
	})

	Action("show", func() {
		Description("Show information about a team.")
		Routing(GET("/:team_id"))
		Params(func() {
			Param("team_id", String, "Team ID")
		})
		Security("Bearer")
		Response(OK, TeamMedia)
	})

	Action("update", func() {
		Description("Update information about a team.")
		Routing(PATCH("/:team_id"))
		Params(func() {
			Param("team_id", String, "team ID")
		})
		Payload(TeamUpdatePayload)
		Security("Bearer")
		Response(OK, TeamMedia)
	})

	Action("delete", func() {
		Description("Delete a team.")
		Routing(DELETE("/:team_id"))
		Params(func() {
			Param("team_id", String, "Team ID")
		})
		Security("Bearer")
		Response(NoContent, func() {})
	})
})
