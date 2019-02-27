package handlers

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Huawei-PaaS/ci-bot/handlers/label"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
)

const (
	kind = "/kind"
)

type GithubPR github.PullRequestEvent

func (s *Server) handlePullRequestEvent(body []byte, client *github.Client) {
	glog.Infof("Received an PullRequest Event")
	// get basic params
	ctx := context.Background()
	var Label []string
	var prEvent github.PullRequestEvent
	listOfAddLabels := make([]string, 0)
	// Unmarshal
	err := json.Unmarshal(body, &prEvent)
	if err != nil {
		glog.Errorf("Failed to unmarshal prEvent: %v", err)
	}
	//Get the Lable to add
	eventBody := strings.SplitAfter(*prEvent.PullRequest.Body, kind)

	if strings.Contains(eventBody[0], kind) {
		if len(eventBody) >= 1 {
			//Trimming /kind lable operation
			labelOp := strings.Split(eventBody[1], "\r\n")
			Op := strings.Trim(labelOp[0], " ")
			Label = append(Label, kind, Op)
			lableBody := strings.Join(Label, " ")

			mapOfAddLabels := label.GetLabelsMap(lableBody)
			listofRepoLabels, _, err := client.Issues.ListLabels(ctx, *prEvent.Repo.Owner.Login, *prEvent.Repo.Name, nil)
			if err != nil {
				glog.Fatalf("Unable to list repository labels. err: %v", err)
			}
			// list labels in current issue
			listofIssueLabels, _, err := client.Issues.ListLabelsByIssue(ctx, *prEvent.Repo.Owner.Login, *prEvent.Repo.Name, *prEvent.Number, nil)
			if err != nil {
				glog.Fatalf("Unable to list issue labels. err: %v", err)
			}
			glog.Infof("List of issue labels: %v", listofIssueLabels)
			//Get the list of add labels
			listOfAddLabels = label.GetListOfAddLabels(mapOfAddLabels, listofRepoLabels, listofIssueLabels)
			_, _, err = client.Issues.AddLabelsToIssue(ctx, *prEvent.Repo.Owner.Login, *prEvent.Repo.Name, *prEvent.Number, listOfAddLabels)
			if err != nil {
				glog.Fatalf("Unable to add labels: %v err: %v", listOfAddLabels, err)
			} else {
				glog.Infof("Add labels successfully: %v", listOfAddLabels)
			}
		}
	} else {
		glog.Infof("No /kind is added to this PR !!")
	}

}

func (s *Server) handlePullRequestCommentEvent(body []byte) {
	glog.Infof("Received an PullRequestComment Event")

}
