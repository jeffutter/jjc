/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/andygrunwald/go-jira"
	"os"
	"strings"
	"time"
)

var SourceStatus string
var DestStatus string
var Repeat bool

func transitionIssue(issue jira.Issue, client *jira.Client) {
	transitions, _, err := client.Issue.GetTransitions(issue.ID)
	if err != nil {
		panic(err)
	}

	transitioned := false

	for ti := range transitions {
		if transitions[ti].Name == DestStatus {
			fmt.Printf("\nTransitioning: %s\n", issue.Key)

			_, err := client.Issue.DoTransition(issue.ID, transitions[ti].ID)
			if err != nil {
				panic(err)
			}

			transitioned = true

			fmt.Printf("\n%s Transitioned to %s\n", issue.Fields.Summary, DestStatus)
			break
		}
	}

	if transitioned == false {
		possibleTransitions := make([]string, len(transitions))
		for i, transition := range transitions {
			possibleTransitions[i] = transition.Name
		}
		fmt.Printf("\n%s can't transition to status: %s. Available statuses: %s\n", issue.Key, DestStatus, strings.Join(possibleTransitions, ", "))
	}
}

func bulkTransition(client *jira.Client) {
	issues, _, err := client.Issue.Search(fmt.Sprintf("assignee=currentUser() and status=\"%s\"", SourceStatus), nil)
	if err != nil {
		panic(err)
	}

	if len(issues) == 0 {
		fmt.Printf("\nThere are no issues in %s Status\n", SourceStatus)
	}

	for ii := range issues {
		transitionIssue(issues[ii], client)
	}
}

var bulkTransitionCmd = &cobra.Command{
	Use:   "bulk-transition",
	Short: "Change the status of all jira tickets for a given user from source to dest",
	Run: func(cmd *cobra.Command, args []string) {
		jiraUsername := strings.TrimSpace(os.Getenv("JIRA_USERNAME"))
		jiraToken := strings.TrimSpace(os.Getenv("JIRA_TOKEN"))
		jiraHost := strings.TrimSpace(os.Getenv("JIRA_HOST"))

		fmt.Printf("\nTransitioning all tickets for %s in %s status to %s", jiraUsername, SourceStatus, DestStatus)

		tp := jira.BasicAuthTransport{
			Username: jiraUsername,
			Password: jiraToken,
		}

		client, err := jira.NewClient(tp.Client(), jiraHost)
		if err != nil {
			panic(err)
		}

		if Repeat {
			for true {
				bulkTransition(client)
				time.Sleep(time.Second * 60)
			}
		} else {
			bulkTransition(client)
		}
	},
}

func init() {
	rootCmd.AddCommand(bulkTransitionCmd)

	bulkTransitionCmd.Flags().StringVarP(&SourceStatus, "source", "s", "", "Source status to transition from")
	bulkTransitionCmd.Flags().StringVarP(&DestStatus, "dest", "d", "", "Dest status to transition to")
	bulkTransitionCmd.Flags().BoolVarP(&Repeat, "repeat", "r", false, "Repeat operation indefinitely")
	bulkTransitionCmd.MarkFlagRequired("source")
	bulkTransitionCmd.MarkFlagRequired("dest")
}
