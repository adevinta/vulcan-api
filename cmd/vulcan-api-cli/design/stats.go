/*
Copyright 2021 Adevinta
*/

package design

import (
	// Use . imports to enable the DSL
	. "github.com/goadesign/goa/design"        // nolint
	. "github.com/goadesign/goa/design/apidsl" // nolint
)

// Stats
var StatsTotalMedia = MediaType("statsTotal", func() {
	Description("Stats by severity")
	Attributes(func() {
		Attribute("critical", Number, "Stats for critical vulnerabilities")
		Attribute("high", Number, "Stats for high vulnerabilities")
		Attribute("medium", Number, "Stats for medium vulnerabilities")
		Attribute("low", Number, "Stats for low vulnerabilities")
		Attribute("informational", Number, "Stats for informational vulnerabilities")
		Attribute("total", Number, "Stats for all vulnerabilities")
	})
	View("default", func() {
		Attribute("critical")
		Attribute("high")
		Attribute("medium")
		Attribute("low")
		Attribute("informational")
		Attribute("total")
	})
})

var StatsMedia = MediaType("stats", func() {
	Description("Stats by severity")
	Attributes(func() {
		Attribute("critical", Number, "Stats for critical vulnerabilities")
		Attribute("high", Number, "Stats for high vulnerabilities")
		Attribute("medium", Number, "Stats for medium vulnerabilities")
		Attribute("low", Number, "Stats for low vulnerabilities")
		Attribute("informational", Number, "Stats for informational vulnerabilities")
	})
	View("default", func() {
		Attribute("critical")
		Attribute("high")
		Attribute("medium")
		Attribute("low")
		Attribute("informational")
	})
})

var StatsMTTRMedia = MediaType("mttr", func() {
	Description("MTTR stats")
	Attributes(func() {
		Attribute("mttr", StatsTotalMedia, "Stats for MTTR by severity")
	})
	View("default", func() {
		Attribute("mttr")
	})
})

var StatsExposureMedia = MediaType("exposure", func() {
	Description("Exposure stats")
	Attributes(func() {
		Attribute("exposure", StatsTotalMedia, "Stats for exposure by different averages")
	})
	View("default", func() {
		Attribute("exposure")
	})
})

var StatsOpenMedia = MediaType("statsOpen", func() {
	Description("Open issues stats")
	Attributes(func() {
		Attribute("open_issues", StatsMedia, "Stats for open issues by severity")
	})
	View("default", func() {
		Attribute("open_issues")
	})
})

var StatsFixedMedia = MediaType("statsFixed", func() {
	Description("Fixed issues stats")
	Attributes(func() {
		Attribute("fixed_issues", StatsMedia, "Stats for fixed issues by severity")
	})
	View("default", func() {
		Attribute("fixed_issues")
	})
})

var StatsCoverageMedia = MediaType("statsCoverage", func() {
	Description("Asset Coverage: discovered vs. confirmed")
	Attributes(func() {
		Attribute("coverage", Number, "Percentage of assets confirmed respect to discovered")
	})
	View("default", func() {
		Attribute("coverage")
	})
})

var _ = Resource("stats", func() {
	Parent("teams")
	BasePath("stats")

	DefaultMedia(StatsMedia)

	Action("mttr", func() {
		Description("Get MTR statistics for a team.")
		Routing(GET("/mttr"))
		Params(func() {
			Param("team_id", String, "Team ID")
			Param("minDate", String, "Minimum date to filter statistics by")
			Param("maxDate", String, "Maximum date to filter statistics by")
		})
		Security("Bearer")
		Response(OK, StatsMTTRMedia)
	})

	Action("exposure", func() {
		Description("Get exposure statistics for a team.")
		Routing(GET("/exposure"))
		Params(func() {
			Param("team_id", String, "Team ID")
			Param("atDate", String, "Specific date to get statistics at")
			Param("minScore", Number, "Minimum issues score filter")
			Param("maxScore", Number, "Maximum issues score filter")
		})
		Security("Bearer")
		Response(OK, StatsExposureMedia)
	})

	Action("open", func() {
		Description("Get open issues statistics for a team.")
		Routing(GET("/open"))
		Params(func() {
			Param("team_id", String, "Team ID")
			Param("minDate", String, "Minimum date to filter statistics by")
			Param("maxDate", String, "Maximum date to filter statistics by")
			Param("atDate", String, "Specific date to get statistics at (incompatible and preferential to min and max date params)")
		})
		Security("Bearer")
		Response(OK, StatsOpenMedia)
	})

	Action("fixed", func() {
		Description("Get fixed issues statistics for a team.")
		Routing(GET("/fixed"))
		Params(func() {
			Param("team_id", String, "Team ID")
			Param("minDate", String, "Minimum date to filter statistics by")
			Param("maxDate", String, "Maximum date to filter statistics by")
			Param("atDate", String, "Specific date to get statistics at (incompatible and preferential to min and max date params)")
		})
		Security("Bearer")
		Response(OK, StatsFixedMedia)
	})

	Action("coverage", func() {
		Description("Get asset coverage for a team.")
		Routing(GET("/coverage"))
		Params(func() {
			Param("team_id", String, "Team ID")
		})
		Security("Bearer")
		Response(OK, StatsCoverageMedia)
	})
})

var _ = Resource("global-stats", func() {
	BasePath("stats")

	DefaultMedia(StatsMedia)

	Action("mttr", func() {
		Description("Get global MTTR statistics.")
		Routing(GET("/mttr"))
		Params(func() {
			Param("minDate", String, "Minimum date to filter statistics by")
			Param("maxDate", String, "Maximum date to filter statistics by")
		})
		Security("Bearer")
		Response(OK, StatsMTTRMedia)
	})

	Action("exposure", func() {
		Description("Get global exposure statistics.")
		Routing(GET("/exposure"))
		Params(func() {
			Param("atDate", String, "Specific date to get statistics at")
			Param("minScore", Number, "Minimum issues score filter")
			Param("maxScore", Number, "Maximum issues score filter")
		})
		Security("Bearer")
		Response(OK, StatsExposureMedia)
	})
})
