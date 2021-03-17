/*
Copyright 2021 Adevinta
*/

package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/adevinta/vulcan-api/cmd/vulcan-api-cli/client"
	"github.com/adevinta/vulcan-api/cmd/vulcan-api-cli/tool/cli"
	goaclient "github.com/goadesign/goa/client"
	"github.com/spf13/cobra"
)

func main() {
	// Create command line parser
	app := &cobra.Command{
		Use:   "Vulcan-API-cli",
		Short: `CLI client for the Vulcan-API service`,
	}

	// Create client struct
	httpClient := newHTTPClient()
	c := client.New(goaclient.HTTPClientDoer(httpClient))

	// Register global flags
	app.PersistentFlags().StringVarP(&c.Scheme, "scheme", "s", "", "Set the requests scheme")
	app.PersistentFlags().StringVarP(&c.Host, "host", "H", "www.vulcan.example.com", "API hostname")
	app.PersistentFlags().DurationVarP(&httpClient.Timeout, "timeout", "t", time.Duration(20)*time.Second, "Set the request timeout")
	app.PersistentFlags().BoolVar(&c.Dump, "dump", false, "Dump HTTP request and response.")

	// Register signer flags
	var key, format string
	app.PersistentFlags().StringVar(&key, "key", "", "API key used for authentication")
	app.PersistentFlags().StringVar(&format, "format", "Bearer %s", "Format used to create auth header or query from key")

	// Parse flags and setup signers
	app.ParseFlags(os.Args)
	bearerSigner := newBearerSigner(key, format)

	// Initialize API client
	c.SetBearerSigner(bearerSigner)
	c.UserAgent = "Vulcan-API-cli/0"

	// Register API commands
	cli.RegisterCommands(app, c)

	// Execute!
	if err := app.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(-1)
	}
}

// newHTTPClient returns the HTTP client used by the API client to make requests to the service.
func newHTTPClient() *http.Client {
	// TBD: Change as needed (e.g. to use a different transport to control redirection policy or
	// disable cert validation or...)
	return http.DefaultClient
}

// newBearerSigner returns the request signer used for authenticating
// against the Bearer security scheme.
func newBearerSigner(key, format string) goaclient.Signer {
	return &goaclient.APIKeySigner{
		SignQuery: false,
		KeyName:   "authorization",
		KeyValue:  key,
		Format:    format,
	}

}
