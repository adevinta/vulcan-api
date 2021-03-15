/*
Copyright 2021 Adevinta
*/

package main

import (
	_ "github.com/adevinta/vulcan-api/cmd/vulcan-api-cli/design"

	"github.com/goadesign/goa/design"
	"github.com/goadesign/goa/goagen/codegen"
	genclient "github.com/goadesign/goa/goagen/gen_client"
	genswagger "github.com/goadesign/goa/goagen/gen_swagger"
)

func main() {
	codegen.ParseDSL()
	codegen.Run(
		genswagger.NewGenerator(
			genswagger.API(design.Design),
			genswagger.OutDir("../../."),
		),
		genclient.NewGenerator(
			genclient.API(design.Design),
		),
	)
}
