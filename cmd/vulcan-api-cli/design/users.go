/*
Copyright 2021 Adevinta
*/

package design

import (
	// Use . imports to enable the DSL
	. "github.com/goadesign/goa/design"        // nolint
	. "github.com/goadesign/goa/design/apidsl" // nolint
)

var UserMedia = MediaType("user", func() {
	Description("User")
	Attributes(func() {
		Attribute("id", String, "User ID", func() { Example("967d9966-b561-4233-bd6f-cac603fd8320") })
		Attribute("firstname", String, "First name", func() { Example("John") })
		Attribute("lastname", String, "Last name", func() { Example("Doe") })
		Attribute("email", String, "Email", func() { Example("john.doe@vulcan.com") })
		Attribute("admin", Boolean, "Admin", func() { Example(true) })
		Attribute("observer", Boolean, "Observer", func() { Example(true) })
		Attribute("active", Boolean, "Active", func() { Example(true) })
		Attribute("active", Boolean, "Active", func() { Example(true) })
		Attribute("last_login", DateTime, "last_login", func() { Example("2018-09-07T10:40:52Z") })
	})
	View("default", func() {
		Attribute("id")
		Attribute("firstname")
		Attribute("lastname")
		Attribute("email")
		Attribute("admin")
		Attribute("observer")
		Attribute("active")
		Attribute("last_login")
	})
})

var UserPayload = Type("UserPayload", func() {
	Attribute("firstname", String, "First Name", func() { Example("John") })
	Attribute("lastname", String, "Last Name", func() { Example("Doe") })
	Attribute("email", String, "Email", func() { Example("john.doe@vulcan.com") })
	Attribute("admin", Boolean, "Admin", func() {
		Example(false)
		Default(false)
	})
	Attribute("observer", Boolean, "Observer", func() {
		Example(false)
		Default(false)
	})
	Attribute("active", Boolean, "Active (Default: true)", func() { Example(true) })
	Required("email")
})

var UserUpdatePayload = Type("UserUpdatePayload", func() {
	Attribute("firstname", String, "First Name", func() { Example("John") })
	Attribute("lastname", String, "Last Name", func() { Example("Doe") })
	Attribute("admin", Boolean, "Admin", func() {
		Example(false)
		Default(false)
	})
	Attribute("active", Boolean, "Active (Default: true)", func() {
		Example(true)
	})
})

var _ = Resource("user", func() {
	BasePath("/users")
	DefaultMedia(UserMedia)

	Action("list", func() {
		Description("List all users")
		Routing(GET(""))
		Security("Bearer")
		Response(OK, CollectionOf(UserMedia))
	})

	Action("create", func() {
		Description("Create user")
		Routing(POST(""))
		Payload(UserPayload)
		Security("Bearer")
		Response(Created, UserMedia)
	})

	Action("show", func() {
		Description("Describe user")
		Routing(GET("/:user_id"))
		Params(func() {
			Param("user_id", String, "User ID")
		})
		Security("Bearer")
		Response(OK, UserMedia)
	})

	Action("update", func() {
		Description("Update user")
		Routing(PATCH("/:user_id"))
		Params(func() {
			Param("user_id", String, "User ID")
		})
		Payload(UserUpdatePayload)
		Security("Bearer")
		Response(OK, UserMedia)
	})

	Action("delete", func() {
		Description("Remove user")
		Routing(DELETE("/:user_id"))
		Params(func() {
			Param("user_id", String, "User ID")
		})
		Security("Bearer")
		Response(NoContent, func() {})
	})

	Action("profile", func() {
		Description("Show profile information for the current authenticated user based on the key used to make the request.")
		Routing(GET("../profile", func() {}))

		Security("Bearer")
		Response(OK, UserMedia)
	})

	Action("list-teams", func() {
		Description("List all teams for an user.")
		Routing(GET("/:user_id/teams"))
		Params(func() {
			Param("user_id", String, "User ID")
		})
		Security("Bearer")
		Response(OK, CollectionOf(TeamMedia))
	})
})
