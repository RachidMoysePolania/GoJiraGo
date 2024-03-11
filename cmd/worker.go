/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	worker "github.com/RachidMoysePolania/JiraWorker/worker_code"
	"github.com/spf13/cobra"
)

var (
	server_name     string
	username        string
	token           string
	jql             string
	export_csv      bool
	change_assignee bool
	csv_file_path   string
	issues          []worker.Issue
)

// workerCmd represents the worker command
var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "do the dirty work",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		client := worker.BasicAuth(server_name, username, token)
		issues = worker.GetTicketsJQL(jql, client)
		switch export_csv {
		case true:
			worker.IssueToCSV(issues)
		}

		switch change_assignee {
		case true:
			issues = worker.GetIssuesFromCSV(csv_file_path, client)
			for _, issue := range issues {
				worker.UpdateIssueAssignee(client, issue.ID, worker.GetUser(client, issue.Assignee))
			}
		}
		worker.PrintIssues(issues)
	},
}

func init() {
	rootCmd.AddCommand(workerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// workerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// workerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	workerCmd.Flags().StringVarP(&server_name, "server_name", "s", "", "define your company for jira cloud")
	workerCmd.Flags().StringVarP(&username, "username", "u", "", "user@company.com")
	workerCmd.Flags().StringVarP(&token, "token", "t", "", "put here your Jira Token")
	workerCmd.MarkFlagRequired("server_name")
	workerCmd.MarkFlagRequired("username")
	workerCmd.MarkFlagRequired("token")

	workerCmd.Flags().StringVarP(&jql, "jql", "q", "", "put here a JQL to get the result")
	workerCmd.Flags().BoolVarP(&export_csv, "csv", "e", false, "this will export an CSV file with all the issues in your JQL.")

	workerCmd.Flags().BoolVarP(&change_assignee, "change-assignee", "c", false, "this will allow you to update the assignee of the issues.")
	workerCmd.Flags().StringVarP(&csv_file_path, "csv-file", "f", "", "the path of the csv file containing all the issues to modify.")
	workerCmd.MarkFlagsRequiredTogether("change-assignee", "csv-file")
}
