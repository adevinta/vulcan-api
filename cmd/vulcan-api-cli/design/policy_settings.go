/*
Copyright 2021 Adevinta
*/

package design

import (
	// Use . imports to enable the DSL
	. "github.com/goadesign/goa/design"        // nolint
	. "github.com/goadesign/goa/design/apidsl" // nolint
)

var PolicySettingsMedia = MediaType("PolicySetting", func() {
	Description("Policy Setting")
	Attributes(func() {
		Attribute("id", String, "Policy ID", func() { Example("5e51174d-3755-48b2-98bf-37e37cb07f6e") })
		Attribute("checktype_name", String, "Check Type Name", func() { Example("vulcan-tls") })
		Attribute("options", String, "options", func() { Example("{\"timeout\":60}") })
	})
	View("default", func() {
		Attribute("id")
		Attribute("checktype_name")
		Attribute("options")
	})
})

var PolicySettingPayload = Type("PolicySettingPayload", func() {
	Attribute("checktype_name", String, "Check Type Name", func() { Example("vulcan-tls") })
	Attribute("options", String, "Options", func() { Example("{\"timeout\":60}") })
	Required("checktype_name")
})

var PolicySettingUploadPayload = Type("PolicySettingUploadPayload", func() {
	Attribute("checktype_name", String, "Check Type Name", func() { Example("vulcan-tls") })
	Attribute("options", String, "Options", func() { Example("{\"timeout\":60}") })
})

var _ = Resource("policy-settings", func() {
	Parent("policies")
	BasePath("/settings")
	DefaultMedia(PolicySettingsMedia)

	Action("list", func() {
		Description("List settings for a policy.")
		Routing(GET(""))
		Security("Bearer")
		Response(OK, CollectionOf(PolicySettingsMedia))
	})

	Action("create", func() {
		Description("Create a new policy setting.")
		Routing(POST(""))
		Payload(PolicySettingPayload)
		Security("Bearer")
		Response(Created, PolicySettingsMedia)
	})

	Action("show", func() {
		Description("Describe a policy setting.")
		Routing(GET("/:settings_id"))
		Params(func() {
			Param("settings_id", String, "CheckType Settings ID")
		})
		Security("Bearer")
		Response(OK, PolicySettingsMedia)
	})

	Action("update", func() {
		Description("Update a policy setting.")
		Routing(PATCH("/:settings_id"))
		Params(func() {
			Param("settings_id", String, "Policy Settings ID")
		})
		Payload(PolicySettingUploadPayload)
		Security("Bearer")
		Response(OK, PolicySettingsMedia)
	})

	Action("delete", func() {
		Description("Delete a policy setting.")
		Routing(DELETE("/:settings_id"))
		Params(func() {
			Param("settings_id", String, "Policy Settings ID")
		})
		Security("Bearer")
		Response(NoContent, func() {})
	})
})
