/*
Copyright 2021 Adevinta
*/

package design

import (
	// Use . imports to enable the DSL
	. "github.com/goadesign/goa/design"        // nolint
	. "github.com/goadesign/goa/design/apidsl" // nolint
)

var ReportMedia = MediaType("report", func() {
	Description("Report")
	Attributes(func() {
		Attribute("id", String, "Report ID", func() { Example("a7f3a072-67bb-41ad-941e-25afcafc0ed5") })
		Attribute("scan_id", String, "Scan ID", func() { Example("360f1c3a-f0e9-4c10-b557-ebf3015a61a9") })
		Attribute("report", String, "Report URL", func() {
			Example("https://insights.vulcan.example.com/2b3123c0b0083ab6d87b7bb743652b9a58125079258b8b4650511bd47bc1a552/2018-07-10/360f1c3a-f0e9-4c10-b557-ebf3015a61a9-full-report.html")
		})
		Attribute("report_json", String, "Report JSON URL", func() {
			Example("https://insights.vulcan.example.com/2b3123c0b0083ab6d87b7bb743652b9a58125079258b8b4650511bd47bc1a552/2018-07-10/360f1c3a-f0e9-4c10-b557-ebf3015a61a9-full-report.json")
		})
		Attribute("status", String, "Status", func() { Example("FINISHED") })
		Attribute("delivered_to", String, "Delivered To", func() { Example("john.doe@vulcan.example.com, jane.doe@vulcan.example.com") })
	})
	View("default", func() {
		Attribute("id")
		Attribute("scan_id")
		Attribute("report")
		Attribute("report_json")
		Attribute("status")
		Attribute("delivered_to")
	})
})

var ReportEmailMedia = MediaType("reportEmail", func() {
	Description("Report Email Body")
	Attributes(func() {
		Attribute("email_body", String, "Email Body", func() { Example("<html><body><title>...</title></body></html>") })
	})
	View("default", func() {
		Attribute("email_body")
	})
})

var _ = Resource("scan report", func() {
	Parent("scan")
	BasePath("report")
	DefaultMedia(ReportMedia)

	Action("show", func() {
		Description("Show information about a scan report")
		Routing(GET(""))
		Security("Bearer")
		Response(OK, ReportMedia)
	})

	Action("generate", func() {
		Description("Triggers report generation. The report will be generated asynchronously on Vulcan Backend.")
		Routing(POST(""))
		Security("Bearer")
		Response(Accepted)
	})

	Action("send", func() {
		Description("Send the generated report by email to the team recipients.")
		Routing(POST("/send"))
		Security("Bearer")
		Response(OK, NoContent)
	})

	Action("email", func() {
		Description("Retrieve the report in html format to be sent by email.")
		Routing(GET("/email"))
		Security("Bearer")
		Response(OK, ReportEmailMedia)
	})
})

var DigestPayload = Type("DigestPayload", func() {
	Attribute("start_date", String, "Start Date", func() { Example("2020-08-11") })
	Attribute("end_date", String, "End Date", func() { Example("2020-09-10") })
})

var _ = Resource("report", func() {
	Parent("teams")
	BasePath("report")

	Action("send digest", func() {
		Description("Send digest report.\nIf no dates are specified, the time range will be set for the last 30 days.")
		Routing(POST("/digest"))
		Payload(DigestPayload)
		Security("Bearer")
		Response(OK)
	})
})
