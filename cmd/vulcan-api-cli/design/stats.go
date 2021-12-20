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

var StatsAveragesMedia = MediaType("statsAverages", func() {
	Description("Stats by different averages")
	Attributes(func() {
		Attribute("percentile_10", Number, "Percentile 10 of the stats")
		Attribute("percentile_25", Number, "Percentile 25 of the stats")
		Attribute("percentile_50", Number, "Percentile 50 or median of the stats")
		Attribute("percentile_75", Number, "Percentile 75 or third quartile of the stats")
		Attribute("percentile_90", Number, "Percentile 90 of the stats")
		Attribute("mean", Number, "Mean of the stats")

	})
	View("default", func() {
		Attribute("percentile_10")
		Attribute("percentile_25")
		Attribute("percentile_50")
		Attribute("percentile_75")
		Attribute("percentile_90")
		Attribute("mean")
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
		Attribute("exposure", StatsAveragesMedia, "Stats for exposure by different averages")
	})
	View("default", func() {
		Attribute("exposure")
	})
})

var StatsCurrentExposureMedia = MediaType("current_exposure", func() {
	Description("Current exposure stats")
	Attributes(func() {
		Attribute("current_exposure", StatsAveragesMedia, "Stats for current exposure by different averages")
	})
	View("default", func() {
		Attribute("current_exposure")
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
		Description("Get exposure statistics for a team. This metric takes into account the exposure across all lifecycle of vulnerabilities.")
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

	Action("current exposure", func() {
		Description("Get current exposure statistics for a team. This metric takes into account only the exposure for open vulnerabilities since the last time they were detected.")
		Routing(GET("/exposure/current"))
		Params(func() {
			Param("team_id", String, "Team ID")
			Param("minScore", Number, "Minimum issues score filter")
			Param("maxScore", Number, "Maximum issues score filter")
		})
		Security("Bearer")
		Response(OK, StatsCurrentExposureMedia)
	})

	Action("open", func() {
		Description("Get open issues statistics for a team.")
		Routing(GET("/open"))
		Params(func() {
			Param("team_id", String, "Team ID")
			Param("minDate", String, "Minimum date to filter statistics by")
			Param("maxDate", String, "Maximum date to filter statistics by")
			Param("atDate", String, "Specific date to get statistics at (incompatible and preferential to min and max date params)")
			Param("identifiers", String, "A comma separated list of asset identifiers")
			Param("labels", String, "A comma separated list of associated labels")
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
			Param("identifiers", String, "A comma separated list of asset identifiers")
			Param("labels", String, "A comma separated list of associated labels")
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
			Param("tags", String, "Comma separated list of team tags to filter by. Only admin and observer users are allowed to set this field.")
			Param("minDate", String, "Minimum date to filter statistics by")
			Param("maxDate", String, "Maximum date to filter statistics by")
		})
		Security("Bearer")
		Response(OK, StatsMTTRMedia)
	})

	Action("exposure", func() {
		Description("Get global exposure statistics. This metric takes into account the exposure across all lifecycle of vulnerabilities.")
		Routing(GET("/exposure"))
		Params(func() {
			Param("tags", String, "Comma separated list of team tags to filter by. Only admin and observer users are allowed to set this field.")
			Param("atDate", String, "Specific date to get statistics at")
			Param("minScore", Number, "Minimum issues score filter")
			Param("maxScore", Number, "Maximum issues score filter")
		})
		Security("Bearer")
		Response(OK, StatsExposureMedia)
	})

	Action("current exposure", func() {
		Description("Get global current exposure statistics. This metric takes into account only the exposure for open vulnerabilities since the last time they were detected.")
		Routing(GET("/exposure/current"))
		Params(func() {
			Param("tags", String, "Comma separated list of team tags to filter by. Only admin and observer users are allowed to set this field.")
			Param("minScore", Number, "Minimum issues score filter")
			Param("maxScore", Number, "Maximum issues score filter")
		})
		Security("Bearer")
		Response(OK, StatsCurrentExposureMedia)
	})

	Action("open", func() {
		Description("Get global open issues statistics.")
		Routing(GET("/open"))
		Params(func() {
			Param("tags", String, "Comma separated list of team tags to filter by. Only admin and observer users are allowed to set this field.")
			Param("minDate", String, "Minimum date to filter statistics by")
			Param("maxDate", String, "Maximum date to filter statistics by")
			Param("atDate", String, "Specific date to get statistics at (incompatible and preferential to min and max date params)")
			Param("identifiers", String, "A comma separated list of asset identifiers")
			Param("labels", String, "A comma separated list of associated labels")
		})
		Security("Bearer")
		Response(OK, StatsOpenMedia)
	})

	Action("fixed", func() {
		Description("Get global fixed issues statistics.")
		Routing(GET("/fixed"))
		Params(func() {
			Param("tags", String, "Comma separated list of team tags to filter by. Only admin and observer users are allowed to set this field.")
			Param("minDate", String, "Minimum date to filter statistics by")
			Param("maxDate", String, "Maximum date to filter statistics by")
			Param("atDate", String, "Specific date to get statistics at (incompatible and preferential to min and max date params)")
			Param("identifiers", String, "A comma separated list of asset identifiers")
			Param("labels", String, "A comma separated list of associated labels")
		})
		Security("Bearer")
		Response(OK, StatsFixedMedia)
	})
})
