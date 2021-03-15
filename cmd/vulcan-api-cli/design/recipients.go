/*
Copyright 2021 Adevinta
*/

package design

import (
	// Use . imports to enable the DSL
	. "github.com/goadesign/goa/design"        // nolint
	. "github.com/goadesign/goa/design/apidsl" // nolint
)

var RecipientMedia = MediaType("recipient", func() {
	Description("Recipient")
	Attributes(func() {
		Attribute("email", String, "email", func() { Example("john.doe@vulcan.com") })
	})
	View("default", func() {
		Attribute("email")
	})
})

var RecipientsPayload = Type("RecipientsPayload", func() {
	Attribute("emails", ArrayOf(String), "Emails", func() { Example([]string{"john.doe@vulcan.com", "jane.doe@vulcan.com"}) })
	Required("emails")
})

var _ = Resource("recipients", func() {
	BasePath("/teams/:team_id/recipients")
	Params(func() {
		Param("team_id", String, "Team ID")
	})

	Action("list", func() {
		Description("List all recipients from a team.")
		Routing(GET(""))
		Security("Bearer")
		Response(OK, CollectionOf(RecipientMedia))
	})

	Action("update", func() {
		Description("Update team recipients.")
		Routing(PUT(""))
		Payload(RecipientsPayload)
		Security("Bearer")
		Response(OK, CollectionOf(RecipientMedia))
	})
})
