/*
Copyright 2021 Adevinta
*/

package design

import (
	// Use . imports to enable the DSL
	. "github.com/goadesign/goa/design"        // nolint
	. "github.com/goadesign/goa/design/apidsl" // nolint
)

// Pagination
var PaginationMedia = MediaType("pagination", func() {
	Description("Pagination info")
	Attributes(func() {
		Attribute("limit", Number, "Limit of results for the list")
		Attribute("offset", Number, "Results list offset")
		Attribute("total", Number, "Total number of results for the list")
		Attribute("more", Boolean, "Indicates if there are more results to query for the list")
	})
	View("default", func() {
		Attribute("limit")
		Attribute("offset")
		Attribute("total")
		Attribute("more")
	})
})

// Target
var TargetMedia = MediaType("target", func() {
	Description("target")
	Attributes(func() {
		Attribute("id", String, "Target ID", func() { Example("f8129c7f-7abf-41ba-ab1f-c97090bb3db4") })
		Attribute("identifier", String, "Target identifier", func() { Example("www.test.com") })
		Attribute("tags", ArrayOf(String), "List of tags associated with target", func() { Example([]string{"team:vulcan"}) })
	})
	View("default", func() {
		Attribute("id")
		Attribute("identifier")
		Attribute("tags")
	})
})

// Issue
var IssueMedia = MediaType("issue", func() {
	Description("Issue")
	Attributes(func() {
		Attribute("id", String, "Issue ID", func() { Example("b0720503-0a84-43fd-9cf4-5bb6c500226f") })
		Attribute("summary", String, "Issue summary", func() { Example("MX presence") })
		Attribute("cwe_id", Number, "Common Weakness Enumeration ID", func() { Example(358) })
		Attribute("description", String, "Issue description", func() { Example("This domain has at least one MX record") })
		Attribute("recommendations", ArrayOf(String), "Recommendations to fix the issue", func() { Example([]string{"It is recommended to run DMARC"}) })
		Attribute("reference_links", ArrayOf(String), "Documentation reference for the issue", func() { Example([]string{}) })
		Attribute("labels", ArrayOf(String), "Associated labels", func() { Example([]string{"Web", "HTTP"}) })
	})
	View("default", func() {
		Attribute("id")
		Attribute("summary")
		Attribute("cwe_id")
		Attribute("description")
		Attribute("recommendations")
		Attribute("reference_links")
		Attribute("labels")
	})
})

// Source
var SourceMedia = MediaType("source", func() {
	Description("source")
	Attributes(func() {
		Attribute("id", String, "Source ID", func() { Example("f8129c7f-7abf-41ba-ab1f-c97090bb3db4") })
		Attribute("name", String, "Source name", func() { Example("vulcan") })
		Attribute("component", String, "Source component", func() { Example("vulcan-tls") })
		Attribute("instance", String, "Source instance ID", func() { Example("b3320f03-1284-42fd-9cf4-5bb6c500966f") })
		Attribute("options", String, "Source options", func() { Example("{\"timeout\":60}") })
		Attribute("time", String, "Source execution end time", func() { Example("2019-06-08 11:16:40") })
	})
	View("default", func() {
		Attribute("id")
		Attribute("name")
		Attribute("component")
		Attribute("instance")
		Attribute("options")
		Attribute("time")
	})
})

// Resource
var ResourceMedia = MediaType("resource", func() {
	Description("resource")
	Attributes(func() {
		Attribute("name", String, "Resource name", func() { Example("Network Resources") })
		Attribute("attributes", ArrayOf(String), "attributes of a resource", func() { Example([]string{"Hostname", "Port"}) })
		Attribute("resources", ArrayOf(HashOf(String, String)), "values for attributes of a resource", func() { Example([]map[string]string{{"Hostname": "test.example.com", "Port": "443"}}) })
	})
	View("default", func() {
		Attribute("name")
		Attribute("attributes")
		Attribute("resources")
	})
})

// Attachment
var AttachmentMedia = MediaType("attachment", func() {
	Description("attachment")
	Attributes(func() {
		Attribute("name", String, "Attachment name", func() { Example("JSON string") })
		Attribute("content_type", String, "Content Type of the attachment", func() { Example("application/json") })
		Attribute("data", Any, "attachment data", func() { Example([]byte{123, 125}) })
	})
	View("default", func() {
		Attribute("name")
		Attribute("content_type")
		Attribute("data")
	})
})

// Findings
var FindingMedia = MediaType("finding", func() {
	Description("Finding")
	Attributes(func() {
		Attribute("id", String, "Finding ID", func() { Example("a8720503-0284-45fd-9cf4-5bb6c500966f") })
		Attribute("issue", IssueMedia, "Issue")
		Attribute("target", TargetMedia, "Target")
		Attribute("affected_resource", String, "Affected Resource", func() { Example("") })
		Attribute("source", SourceMedia, "Source")
		Attribute("details", String, "Details", func() { Example("") })
		Attribute("impact_details", String, "Impact details", func() { Example("") })
		Attribute("status", String, "Status", func() { Example("OPEN") })
		Attribute("score", Number, "Score", func() { Example(6.9) })
		Attribute("current_exposure", Number, "Current exposure (hours). Only for OPEN findings", func() { Example(4631) })
		Attribute("total_exposure", Number, "Total exposure (hours)", func() { Example(4631) })
		Attribute("resources", ArrayOf(ResourceMedia), "Resources")
		Attribute("attachments", ArrayOf(AttachmentMedia), "Attachments")
	})
	View("default", func() {
		Attribute("id")
		Attribute("issue")
		Attribute("target")
		Attribute("affected_resource")
		Attribute("source")
		Attribute("details")
		Attribute("impact_details")
		Attribute("status")
		Attribute("score")
		Attribute("current_exposure")
		Attribute("total_exposure")
		Attribute("resources")
		Attribute("attachments")
	})
})

var FindingsListMedia = MediaType("findings_list", func() {
	Description("Findings list")
	Attributes(func() {
		Attribute("findings", CollectionOf(FindingMedia), "List of findings")
		Attribute("pagination", PaginationMedia, "Pagination info")
	})
	View("default", func() {
		Attribute("findings")
		Attribute("pagination")
	})
})

// FindingsIssue
var FindingsIssueMedia = MediaType("findings_issue", func() {
	Description("Findings by Issue")
	Attributes(func() {
		Attribute("issue_id", String, "Issue ID", func() { Example("a8720503-0284-45fd-9cf4-5bb6c500966f") })
		Attribute("summary", String, "Issue summary", func() { Example("Site Without HTTPS") })
		Attribute("targets_count", Number, "Number of targets affected by the issue", func() { Example(14) })
		Attribute("max_score", Number, "Max score for the issue among the affected assets", func() { Example(6.9) })
	})
	View("default", func() {
		Attribute("issue_id")
		Attribute("summary")
		Attribute("targets_count")
		Attribute("max_score")
	})
})

var FindingsIssuesListMedia = MediaType("findings_issues_list", func() {
	Description("Findings by Issue list")
	Attributes(func() {
		Attribute("issues", CollectionOf(FindingsIssueMedia), "List of affected assets by issue")
		Attribute("pagination", PaginationMedia, "Pagination info")
	})
	View("default", func() {
		Attribute("issues")
		Attribute("pagination")
	})
})

// FindingsTarget
var FindingsTargetMedia = MediaType("findings_target", func() {
	Description("Findings by Target")
	Attributes(func() {
		Attribute("target_id", String, "Target ID", func() { Example("a8720503-0284-45fd-9cf4-5bb6c500966f") })
		Attribute("identifier", String, "Target Identifier", func() { Example("vulcan.example.com") })
		Attribute("findings_count", Number, "Number of findings for the target", func() { Example(5) })
		Attribute("max_score", Number, "Max score for the issue among the affected assets", func() { Example(6.9) })
	})
	View("default", func() {
		Attribute("target_id")
		Attribute("identifier")
		Attribute("findings_count")
		Attribute("max_score")
	})
})

var FindingsTargetsListMedia = MediaType("findings_targets_list", func() {
	Description("Findings by Target list")
	Attributes(func() {
		Attribute("targets", CollectionOf(FindingsTargetMedia), "List of findings per asset")
		Attribute("pagination", PaginationMedia, "Pagination info")
	})
	View("default", func() {
		Attribute("targets")
		Attribute("pagination")
	})
})

// FindingsLabels
var FindingsLabelsMedia = MediaType("findings_labels", func() {
	Description("Findings Labels")
	Attributes(func() {
		Attribute("labels", ArrayOf(String), "associated labels", func() { Example([]string{"Web", "HTTP"}) })
	})
	View("default", func() {
		Attribute("labels")
	})
})

var _ = Resource("findings", func() {
	Parent("teams")
	BasePath("findings")

	DefaultMedia(FindingsListMedia)

	Action("list findings", func() {
		Description("List all findings from a team.")
		Routing(GET(""))
		Params(func() {
			Param("team_id", String, "Team ID")
			Param("status", String, "Findings Status")
			Param("minScore", Number, "Findings minimum score")
			Param("maxScore", Number, "Findings maximum score")
			Param("atDate", String, "Allows to get findings list at a specific date (incompatible and preferential to min and max date params)")
			Param("minDate", String, "Allows to get findings list from a specific date")
			Param("maxDate", String, "Allows to get findings list until a specific date")
			Param("sortBy", String, "Sorting criteria. Supported fields: score, -score (for descending order)")
			Param("page", Number, "Requested page")
			Param("size", Number, "Requested page size")
			Param("identifier", String, "Allows to get findings list for a specific asset identifier")
			Param("targetID", String, "Target ID (Vulnerability DB)")
			Param("issueID", String, "Issue ID (Vulnerability DB)")
			Param("identifiers", String, "A comma separated list of identifiers")
			Param("labels", String, "A comma separated list of associated labels")
		})
		Security("Bearer")
		Response(OK, FindingsListMedia)
	})

	Action("list findings issues", func() {
		Description("List number of findings and max score per issue.")
		Routing(GET("/issues"))
		Params(func() {
			Param("team_id", String, "Team ID")
			Param("status", String, "Findings Status")
			Param("minDate", String, "Allows to get findings list from a specific date")
			Param("maxDate", String, "Allows to get findings list until a specific date")
			Param("atDate", String, "Allows to get issues list at a specific date")
			Param("sortBy", String, "Sorting criteria. Supported fields: score, findings_count (use - for descending order. E.g.: -score)")
			Param("page", Number, "Requested page")
			Param("size", Number, "Requested page size")
			Param("targetID", String, "Target ID (Vulnerability DB)")
			Param("identifiers", String, "A comma separated list of identifiers")
			Param("labels", String, "A comma separated list of associated labels")
		})
		Security("Bearer")
		Response(OK, FindingsIssuesListMedia)
	})

	Action("Find findings from a Issue", func() {
		Description("Find all findings from a team and issue.")
		Routing(GET("/issues/:issue_id"))
		Params(func() {
			Param("team_id", String, "Team ID")
			Param("issue_id", String, "Issue ID")
			Param("status", String, "Findings Status")
			Param("minScore", Number, "Findings minimum score")
			Param("maxScore", Number, "Findings maximum score")
			Param("atDate", String, "Allows to get findings list at a specific date (incompatible and preferential to min and max date params)")
			Param("minDate", String, "Allows to get findings list from a specific date")
			Param("maxDate", String, "Allows to get findings list until a specific date")
			Param("sortBy", String, "Sorting criteria. Supported fields: score, -score (for descending order)")
			Param("page", Number, "Requested page")
			Param("size", Number, "Requested page size")
			Param("identifiers", String, "A comma separated list of identifiers")
			Param("labels", String, "A comma separated list of associated labels")
		})
		Security("Bearer")
		Response(OK, FindingsListMedia)
	})

	Action("list findings targets", func() {
		Description("List number of findings and max score per target.")
		Routing(GET("/targets"))
		Params(func() {
			Param("team_id", String, "Team ID")
			Param("status", String, "Findings Status")
			Param("minDate", String, "Allows to get findings list from a specific date")
			Param("maxDate", String, "Allows to get findings list until a specific date")
			Param("atDate", String, "Allows to get targets list at a specific date")
			Param("sortBy", String, "Sorting criteria. Supported fields: score, findings_count (use - for descending order. E.g.: -score)")
			Param("page", Number, "Requested page")
			Param("size", Number, "Requested page size")
			Param("issueID", String, "Issue ID (Vulnerability DB)")
			Param("identifiers", String, "A comma separated list of identifiers")
			Param("labels", String, "A comma separated list of associated labels")
		})
		Security("Bearer")
		Response(OK, FindingsTargetsListMedia)
	})

	Action("Find findings from a Target", func() {
		Description("Find all findings from a team and target.")
		Routing(GET("/targets/:target_id"))
		Params(func() {
			Param("team_id", String, "Team ID")
			Param("target_id", String, "Target ID")
			Param("status", String, "Findings Status")
			Param("minScore", Number, "Findings minimum score")
			Param("maxScore", Number, "Findings maximum score")
			Param("atDate", String, "Allows to get findings list at a specific date (incompatible and preferential to min and max date params)")
			Param("minDate", String, "Allows to get findings list from a specific date")
			Param("maxDate", String, "Allows to get findings list until a specific date")
			Param("sortBy", String, "Sorting criteria. Supported fields: score, -score (for descending order)")
			Param("page", Number, "Requested page")
			Param("size", Number, "Requested page size")
			Param("identifiers", String, "A comma separated list of identifiers")
			Param("labels", String, "A comma separated list of associated labels")
		})
		Security("Bearer")
		Response(OK, FindingsListMedia)
	})

	Action("List findings labels", func() {
		Description("List all findings labels.")
		Routing(GET("/labels"))
		Params(func() {
			Param("team_id", String, "Team ID")
			Param("status", String, "Findings status")
			Param("atDate", String, "Allows to get findings list at a specific date (incompatible and preferential to min and max date params)")
			Param("minDate", String, "Allows to get findings list from a specific date")
			Param("maxDate", String, "Allows to get findings list until a specific date")
			Param("identifiers", String, "A comma separated list of identifiers")
		})
		Security("Bearer")
		Response(OK, FindingsLabelsMedia)
	})

	Action("Find finding", func() {
		Description("Find a finding.")
		Routing(GET("/:finding_id"))
		Params(func() {
			Param("team_id", String, "Team ID")
			Param("finding_id", String, "Finding ID")
		})
		Security("Bearer")
		Response(OK, FindingMedia)
	})

	Action("Submit a Finding Overwrite", func() {
		Description("Overwrite data for a specific finding.")
		Routing(POST("/:finding_id/overwrites"))
		Payload(FindingOverwritePayload)
		Security("Bearer")
		Response(OK, func() {})
	})

	Action("List Finding Overwrites", func() {
		Description("List Finding Overwrites.")
		Routing(GET("/:finding_id/overwrites"))
		Security("Bearer")
		Response(OK, CollectionOf(FindingOverwriteMedia))
	})
})

var FindingOverwritePayload = Type("FindingOverwritePayload", func() {
	Attribute("status", String, "Status", func() { Example("FALSE_POSITIVE") })
	Attribute("notes", String, "Notes", func() { Example("This is a false positive because...") })
	Required("status")
	Required("notes")
})

// FindingOverwrite
var FindingOverwriteMedia = MediaType("finding_overwrite", func() {
	Description("Finding Overwrite")
	Attributes(func() {
		Attribute("id", String, "Finding Overwrite ID", func() { Example("b0720503-0a84-43fd-9cf4-5bb6c500226f") })
		Attribute("user", String, "User who requested the finding overwrite", func() { Example("user@adevinta.com") })
		Attribute("finding_id", String, "Finding ID", func() { Example("3c7d7003-c53d-4ccc-80e7-f21da241b2d4") })
		Attribute("status", String, "The status requested for the finding referenced by the finding_id field", func() { Example("FALSE_POSITIVE") })
		Attribute("status_previous", String, "The previous status for the finding referenced by the finding_id field", func() { Example("OPEN") })
		Attribute("notes", String, "Complementary information", func() { Example("This finding is a false positive because...") })
		Attribute("tag", String, "The tag associated to the user/team who requested this overwrite", func() { Example("team:security") })
		Attribute("created_at", DateTime, "Creation time", func() { Example("2021-03-27T00:26:43.211506Z") })
	})
	View("default", func() {
		Attribute("id")
		Attribute("user")
		Attribute("finding_id")
		Attribute("status")
		Attribute("status_previous")
		Attribute("notes")
		Attribute("tag")
		Attribute("created_at")
	})
})
