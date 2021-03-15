/*
Copyright 2021 Adevinta
*/

package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/adevinta/vulcan-api/cmd/vulcan-cli/cli"
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

var DownloadReport = &cobra.Command{
	Use:   `report`,
	Short: `Downloads scan's report email to a temporary folder`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDownloadReport(args, apiClient)
	},
}

var SendReport = &cobra.Command{
	Use:   `send`,
	Short: `Sends scan's report emails`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSendReport(args, apiClient)
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

	DownloadReport.PersistentFlags().StringVarP(&scanID, "scan", "", "", "Scan ID of the report to download")
	DownloadReport.PersistentFlags().StringVarP(&team, "team", "", "", "Team's name of the report to download")
	DownloadReport.PersistentFlags().StringVarP(&scansFile, "scan-file", "i", "", "Scans file with a list of reports to download")
	ScanTeam.AddCommand(DownloadReport)

	DownloadReport.AddCommand(SendReport)
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

	tmpfile, err := ioutil.TempFile("", "vulcan-scan-*.txt")
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

	return ioutil.WriteFile(path, []byte(scans.String()), perm)
}

func runDownloadReport(args []string, apiClient *cli.CLI) error {
	m := make(map[string]string)
	switch {
	case scanID != "" && scansFile != "":
		fallthrough
	case scanID == "" && scansFile == "":
		return errors.New("only one of the flags allowed (scan or scan-file)")
	case scanID != "" && team == "":
		return errors.New("team name can not be empty when using scan flag")
	case scansFile != "" && team != "":
		return errors.New("team name can not be set when using scan-file flag")
	case scanID != "":
		email, err := apiClient.ReportEmail(team, scanID)
		if err != nil {
			return err
		}
		m[team] = email
	case scansFile != "":
		scans, err := cli.ParseScans(scansFile)
		if err != nil {
			return err
		}

		for _, scan := range scans {
			scan := scan

			if strings.Contains(scan.Status, StatusError) {
				continue
			}

			if !isFinalState(scan.Status) {
				scan, err = apiClient.RefreshScan(scan)
				if err != nil {
					return err
				}
			}
			if scan.Status == StatusFinished {
				email, err := apiClient.ReportEmail(scan.Team, scan.ID)
				if err != nil {
					return err
				}
				m[scan.Team] = email
			}
		}
	}

	dir, err := ioutil.TempDir("", "vulcan-reports-")
	if err != nil {
		return err
	}

	fmt.Printf("[*] Writing report emails to file '%s'\n", dir)

	for k, v := range m {
		tmpfn := filepath.Join(dir, fmt.Sprintf("%s.html", k))

		if err := ioutil.WriteFile(tmpfn, []byte(v), 0666); err != nil {
			return err
		}
	}

	return nil
}

func runSendReport(args []string, apiClient *cli.CLI) error {
	switch {
	case scanID != "" && scansFile != "":
		fallthrough
	case scanID == "" && scansFile == "":
		return errors.New("only one of the flags allowed (scan or scan-file)")
	case scanID != "" && team == "":
		return errors.New("team name can not be empty when using scan flag")
	case scansFile != "" && team != "":
		return errors.New("team name can not be set when using scan-file flag")
	case scanID != "":
		return apiClient.SendReport(team, scanID)
	case scansFile != "":
		scans, err := cli.ParseScans(scansFile)
		if err != nil {
			return err
		}

		for _, scan := range scans {
			scan := scan

			if strings.Contains(scan.Status, StatusError) {
				continue
			}

			if !isFinalState(scan.Status) {
				scan, err = apiClient.RefreshScan(scan)
				if err != nil {
					return err
				}
			}
			if scan.Status == StatusFinished {
				if err := apiClient.SendReport(scan.Team, scan.ID); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func isFinalState(state string) bool {
	if strings.Contains(state, StatusError) || state == StatusFinished || state == StatusAborted {
		return true
	}
	return false
}
