/*
Copyright 2021 Adevinta
*/

package design

import (
	// Use . imports to enable the DSL
	. "github.com/goadesign/goa/design"        // nolint
	. "github.com/goadesign/goa/design/apidsl" // nolint
)

var ScanMedia = MediaType("scan", func() {
	Description("Scan")
	Attributes(func() {
		Attribute("id", String, "Scan ID", func() { Example("23e7034f-8180-4895-8e7f-73f2f7d90631") })
		Attribute("program", ProgramMedia, "Program")
		Attribute("scheduled_time", DateTime, "Scheduled Time", func() { Example("2018-09-07T10:40:52Z") })
		Attribute("start_time", DateTime, "Start Time", func() { Example("2018-09-07T10:40:52Z") })
		Attribute("end_time", DateTime, "End Time", func() { Example("2018-09-07T10:40:52Z") })
		Attribute("progress", Number, "Progress", func() { Example(1) })
		Attribute("checks_count", Integer, "Checks Count", func() { Example(20) })
		Attribute("status", String, "Status", func() { Example("FINISHED") })
		Attribute("requested_by", String, "Requested By", func() { Example("john.doe@vulcan.example.com") })
	})
	View("default", func() {
		Attribute("id")
		Attribute("program")
		Attribute("scheduled_time")
		Attribute("start_time")
		Attribute("end_time")
		Attribute("progress")
		Attribute("checks_count")
		Attribute("status")
		Attribute("requested_by")
	})
})

var ScanPayload = Type("ScanPayload", func() {
	Attribute("program_id", String, "Program ID")
	Attribute("program_id", String, "Program ID", func() { Example("1bb4c953-245e-477b-b005-400f319274f2") })
	Attribute("scheduled_time", DateTime, "Group ID", func() { Example("2018-09-07T10:40:52Z") })
	Required("program_id")
})

var _ = Resource("scan", func() {
	Parent("teams")
	BasePath("/scans")
	DefaultMedia(ScanMedia)

	Action("create", func() {
		Description("Create scan")
		Routing(POST(""))
		Payload(ScanPayload)
		Security("Bearer")
		Response(Created, ScanMedia)
	})

	Action("show", func() {
		Description("Describe scan")
		Routing(GET("/:scan_id"))
		Params(func() {
			Param("scan_id", String, "Scan ID")
		})
		Security("Bearer")
		Response(OK, ScanMedia)
	})
})
