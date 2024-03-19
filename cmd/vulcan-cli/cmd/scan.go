/*
Copyright 2021 Adevinta
*/

package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/adevinta/vulcan-api/cmd/vulcan-cli/cli"
	"github.com/spf13/cobra"
)

const StatusFinished = "FINISHED"
const StatusAborted = "ABORTED"
const StatusError = "ERROR"

var ScanTeam = &cobra.Command{
	Use:   `scan <teams_dir> <team_name>`,
	Short: `Launches a scan against a team storing its info in a temporary file`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScanTeam(args, apiClient)
	},
}

var RefreshScans = &cobra.Command{
	Use:   `refresh <scans_file>`,
	Short: `Refreshes scans stored in a temporary file`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runRefreshScans(args, apiClient)
	},
}

var (
	programID string
	scanID    string
	team      string
	scansFile string
)

func init() {
	ScanTeam.Flags().StringVarP(&programID, "program", "p", "periodic-full-scan", "Program ID to be launched")
	rootCmd.AddCommand(ScanTeam)

	ScanTeam.AddCommand(RefreshScans)

}

func runScanTeam(args []string, apiClient *cli.CLI) error {
	path := args[0]
	name := args[1]

	localTeams, err := readLocalTeams(path, name)
	if err != nil {
		return err
	}

	var scans cli.Scans
	for _, t := range localTeams {
		var scan *cli.Scan
		scan, err = apiClient.LaunchScan(t, programID)
		if err != nil {
			scan = &cli.Scan{
				Program: programID,
				Status:  fmt.Sprintf("ERROR: %s", err.Error()),
				Team:    t.Name,
			}
		}
		scans = append(scans, scan)
	}

	tmpfile, err := os.CreateTemp("", "vulcan-scan-*.txt")
	if err != nil {
		return err
	}
	defer tmpfile.Close() // nolint

	fmt.Printf("[*] Writing scans to file '%s'\n", tmpfile.Name())

	_, err = tmpfile.Write([]byte(scans.String()))

	return err
}

func runRefreshScans(args []string, apiClient *cli.CLI) error {
	path := args[0]

	scans, err := cli.ParseScans(path)
	if err != nil {
		return err
	}

	for i, scan := range scans {
		if isFinalState(scan.Status) {
			continue
		}

		s, err := apiClient.RefreshScan(scan) // nolint
		if err != nil {
			return err
		}

		scans[i] = s
	}

	fmt.Println(scans)

	f, err := os.Open(path) // nolint
	if err != nil {
		return err
	}
	stat, err := f.Stat()
	if err != nil {
		f.Close() // nolint
		return err
	}
	perm := stat.Mode()

	if err := f.Close(); err != nil {
		return errors.New("can not close file before writing")
	}

	return os.WriteFile(path, []byte(scans.String()), perm)
}

func isFinalState(state string) bool {
	if strings.Contains(state, StatusError) || state == StatusFinished || state == StatusAborted {
		return true
	}
	return false
}
