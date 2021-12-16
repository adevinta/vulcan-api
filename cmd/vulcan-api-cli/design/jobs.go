/*
Copyright 2021 Adevinta
*/

package design

import (
	// Use . imports to enable the DSL
	. "github.com/goadesign/goa/design"        // nolint
	. "github.com/goadesign/goa/design/apidsl" // nolint
)

var JobMedia = MediaType("job", func() {
	Description("Job")
	Attributes(func() {
		Attribute("id", String, "Job ID", func() { Example("967d9966-b561-4233-bd6f-cac603fd8320") })
		Attribute("team_id", String, "Team ID", func() { Example("9cb0bb2b-ca36-4877-acad-9dde23880595") })
		Attribute("operation", String, "Operation that triggered the job", func() { Example("OnboardDiscoveredAssets") })
		Attribute("status", String, func() {
			Description(`Indicates the status of the operation. The possible values are:
	- 'PENDING': The job has been noted and is pending to be processed
	- 'RUNNING': The job is on execution
	- 'DONE': The job has finished, either successfully or unsuccesfully. Result.error needs to be processed to determine it`)
			Example("PROCESSING")
		})
		Attribute("result", func() {
			Description("Result of the job operation")
			Attribute("data", Any, func() {
				Description("Optionally populated field when the job finishes correctly, that returns execution related data. The format of the data is defined per operation type")
				Example(`{"assets : ["abb77c5e-2442-4673-9d0a-957fb43e416c", "2e016860-f772-416d-b551-bf384351dd5f"}`)
			})
			Attribute("error", String, func() {
				Description("When not empty indicates that the job failed")
				Example("Invalid asset type")
			})
		})
	})
	View("default", func() {
		Attribute("id")
		Attribute("team_id")
		Attribute("operation")
		Attribute("status")
		Attribute("result")
	})
})

var _ = Resource("jobs", func() {
	BasePath("/jobs")
	DefaultMedia(JobMedia)

	Action("show", func() {
		Description(`Describes job status and results. The possible values for the status are:
	- 'PENDING': The job has been noted and is pending to be processed
	- 'RUNNING': The job is on execution
	- 'DONE': The job has finished, either successfully or unsuccesfully. Result.error needs to be processed to determine it

The results field indicates if there was an error during the execution of the job, and otherwise can return data from the job execution`)
		Routing(GET("/:job_id"))
		Params(func() {
			Param("job_id", String, "Job ID")
		})
		Security("Bearer")
		Response(OK, JobMedia)
	})
})
