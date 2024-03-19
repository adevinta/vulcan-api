/*
Copyright 2021 Adevinta
*/

package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/adevinta/vulcan-api/cmd/vulcan-cli/cli"
	"github.com/adevinta/vulcan-groupie/pkg/groupie"
	vulcanreport "github.com/adevinta/vulcan-report"
	"github.com/spf13/cobra"
)

var (
	minScore  float64
	csvOutput bool

	status = "OPEN"

	CPanel = &cobra.Command{
		Use:   `cpanel <teams_dir>`,
		Short: `Shows information about open issues in teams`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCPanel(args, apiClient)
		},
	}
)

func init() {
	CPanel.Flags().Float64VarP(&minScore, "minScore", "m", 7.0, "Minimum score of the retrieved findings.")
	CPanel.Flags().BoolVarP(&csvOutput, "csv", "c", false, "Use CSV output instead of JSON.")

	rootCmd.AddCommand(CPanel)
}

type issue struct {
	summary         string
	score           float32
	unclassified    bool
	checktype       string
	affectedTargets map[string]bool
	affectedTeams   map[string]*cli.Team
}

type issues []*issue

func (i issues) Len() int      { return len(i) }
func (i issues) Swap(j, k int) { i[j], i[k] = i[k], i[j] }
func (i issues) Less(j, k int) bool {
	switch {
	case i[j].score != i[k].score:
		return i[j].score > i[k].score
	case len(i[j].affectedTeams) != len(i[k].affectedTeams):
		return len(i[j].affectedTeams) > len(i[k].affectedTeams)
	case len(i[j].affectedTargets) != len(i[k].affectedTargets):
		return len(i[j].affectedTargets) > len(i[k].affectedTargets)
	default:
		return i[j].summary < i[k].summary
	}
}

func runCPanel(args []string, apiClient *cli.CLI) error {
	path := args[0]

	teams, err := cli.ReadLocalTeams(path)
	if err != nil {
		return err
	}

	issuesMap := make(map[string]*issue)
	for _, team := range teams {
		findings, err := apiClient.Findings(team.ID, minScore, &status)
		if err != nil {
			return err
		}

		for _, f := range findings {
			i, ok := issuesMap[f.Summary]
			if !ok {
				i = &issue{
					summary:         f.Summary,
					score:           vulcanreport.ScoreSeverity(vulcanreport.RankSeverity(float32(f.Score))),
					checktype:       f.Checktype,
					unclassified:    !groupie.Classified(f.Summary),
					affectedTargets: make(map[string]bool),
					affectedTeams:   make(map[string]*cli.Team),
				}
			}
			i.affectedTargets[f.Target] = true
			i.affectedTeams[team.Name] = team

			issuesMap[f.Summary] = i
		}
	}

	var i issues
	for _, v := range issuesMap {
		i = append(i, v)
	}

	sort.Sort(i)

	return printIssues(i)
}

func printIssues(i issues) error {
	type printableIssue struct {
		Summary         string   `json:"summary"`
		Severity        string   `json:"severity"`
		Checktype       string   `json:"checktype"`
		Unclassified    bool     `json:"unclassified"`
		AffectedTeams   []string `json:"affected_teams"`
		AffectedTargets []string `json:"affected_targets"`
	}

	var printable []printableIssue
	for _, issue := range i {
		var teams []string
		for _, t := range issue.affectedTeams {
			teams = append(teams, t.Name)
		}
		sort.Strings(teams)

		var targets []string
		for t := range issue.affectedTargets {
			targets = append(targets, t)
		}
		sort.Strings(targets)

		p := printableIssue{
			Summary:         issue.summary,
			Severity:        severityToStr(vulcanreport.RankSeverity(issue.score)),
			Checktype:       issue.checktype,
			Unclassified:    issue.unclassified,
			AffectedTeams:   teams,
			AffectedTargets: targets,
		}

		printable = append(printable, p)
	}

	if csvOutput {
		w := csv.NewWriter(os.Stdout)

		header := []string{
			"Severity",
			"Title",
			"Affected Teams",
			"Affected Targets",
			"Checktype",
			"Unclassified",
		}
		if err := w.Write(header); err != nil {
			return err
		}

		for _, p := range printable {
			record := []string{
				p.Severity,
				p.Summary,
				strings.Join(p.AffectedTeams, "\n"),
				strconv.Itoa(len(p.AffectedTargets)),
				p.Checktype,
				fmt.Sprintf("%t", p.Unclassified),
			}
			if err := w.Write(record); err != nil {
				return err
			}
		}

		w.Flush()

		return w.Error()
	}

	str, err := json.Marshal(printable)
	if err != nil {
		return err
	}

	fmt.Println(string(str))

	return nil
}

func severityToStr(severity vulcanreport.SeverityRank) string {
	switch severity {
	case vulcanreport.SeverityNone:
		return "Info"
	case vulcanreport.SeverityLow:
		return "Low"
	case vulcanreport.SeverityMedium:
		return "Medium"
	case vulcanreport.SeverityHigh:
		return "High"
	case vulcanreport.SeverityCritical:
		return "Critical"
	default:
		return "N/A"
	}
}
