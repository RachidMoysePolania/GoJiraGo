package worker

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	jira "github.com/andygrunwald/go-jira/v2/cloud"
	"github.com/gocarina/gocsv"
	table "github.com/jedib0t/go-pretty/v6/table"
)

type Issue struct {
	ID          string `csv:"id"`
	Key         string `csv:"key"`
	Priority    string `csv:"priority"`
	Summary     string `csv:"summary"`
	Description string `csv:"description"`
	AssetName   string `csv:"asset_name"`
	Assignee    string `csv:"assignee"`
	Comment     string `csv:"comment"`
}

func BasicAuth(server_name, username, token string) *jira.Client {
	jiraUrl := fmt.Sprintf("https://%v.atlassian.net/", server_name)
	tp := jira.BasicAuthTransport{
		Username: username,
		APIToken: token,
	}
	client, err := jira.NewClient(jiraUrl, tp.Client())
	if err != nil {
		panic(err)
	}
	return client
}

func GetTicketsJQL(jql string, client *jira.Client) []Issue {
	var allissues []Issue
	last := 0
	for {
		issues, _, err := client.Issue.Search(context.Background(), jql, &jira.SearchOptions{StartAt: last})
		if err != nil {
			panic(err)
		}
		if len(issues) == 0 {
			break
		}
		for _, i := range issues {
			allissues = append(allissues, Issue{
				ID:          i.ID,
				Key:         i.Key,
				Priority:    i.Fields.Priority.Name,
				Summary:     i.Fields.Summary,
				Description: fmt.Sprintf("%v", i.Fields.Unknowns["customfield_11611"]),
				AssetName:   fmt.Sprintf("%v", i.Fields.Unknowns["customfield_11607"]),
				Assignee:    i.Fields.Assignee.EmailAddress,
				Comment:     "",
			})
		}
		last += len(issues)
	}
	return allissues
}

func PrintIssues(issues []Issue) {
	rowConfigAutoMerge := table.RowConfig{AutoMerge: true}
	t := table.NewWriter()
	t.AppendHeader(table.Row{"Key", "Priority", "AssetName", "Summary"}, rowConfigAutoMerge)

	for _, i := range issues {
		t.AppendRow(table.Row{i.Key, i.Priority, i.AssetName, i.Summary}, rowConfigAutoMerge)
	}
	t.SetStyle(table.StyleColoredBlackOnCyanWhite)
	t.SetAllowedRowLength(100)
	fmt.Println(t.Render())
}

func IssueToCSV(issues []Issue) bool {
	gocsv.TagSeparator = ";"
	csvContent, err := gocsv.MarshalString(&issues)
	if err != nil {
		log.Println("Error converting to CSV..!", err)
	}

	currentTime := time.Now()
	if err := os.WriteFile(fmt.Sprintf("issues-%v.csv", currentTime.Format("2006-January-02")), []byte(csvContent), 0777); err != nil {
		log.Println("Error writing file", err)
		return false
	}

	return true
}

func GetIssuesFromCSV(csvfile string, client *jira.Client) []Issue {
	file, err := os.Open(csvfile)
	if err != nil {
		log.Println("Error opening csv file...", err)
	}
	defer file.Close()

	issues := []Issue{}
	if err := gocsv.UnmarshalFile(file, &issues); err != nil {
		log.Println("Error unmarshaling file...", err)
	}

	return issues
}

func UpdateIssueAssignee(client *jira.Client, issueId string, assignee *jira.User) {
	client.Issue.UpdateAssignee(context.Background(), issueId, assignee)
}

func GetUser(client *jira.Client, user_name string) *jira.User {
	if user, _, _ := client.User.GetCurrentUser(context.Background()); user.EmailAddress == user_name || user_name == "" {
		return user
	}
	user, _, err := client.User.Find(context.Background(), user_name)
	if err != nil {
		log.Println("error getting the user... ", err)
	}
	return &user[0]
}
