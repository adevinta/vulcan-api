/*
Copyright 2021 Adevinta
*/

package design

import (
	// Use . imports to enable the DSL
	. "github.com/goadesign/goa/design"        // nolint
	. "github.com/goadesign/goa/design/apidsl" // nolint
)

var APIKey = APIKeySecurity("Bearer", func() {
	Header("authorization")
})

var _ = API("Vulcan-API", func() {
	Title("Vulcan API")
	Description("Public API for Vulcan Vulnerability Scan Engine")
	Version("1.1.3")
	Scheme("https")
	Host("www.vulcan.example.com")
	BasePath("/api/v1")
	Consumes("application/json")
	//Consumes("application/x-www-form-urlencoded", func() {
	//	Package("github.com/goadesign/goa/encoding/form")
	//})
})

var APIErrorMedia = MediaType("error", func() {
	Description("Error")
	Attributes(func() {
		Attribute("code", Integer, "Code")
		Attribute("error", String, "Error")
		Attribute("type", String, "Type")
		Required("code", "error", "type")
	})
	View("default", func() {
		Attribute("code")
		Attribute("error")
		Attribute("type")
	})
})
