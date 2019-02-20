package handlers

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
)

const (
	Assign   = "/assign"
	Unassign = "/unassign"
)

type GithubIssue github.Issue

func (s *Server) handleIssueEvent(body []byte) {
	glog.Info("Received an Issue Event")

}

//function to handle issue comments
func (s *Server) handleIssueCommentEvent(body []byte, client *github.Client) {
	var commentEvent github.IssueCommentEvent

	err := json.Unmarshal(body, &commentEvent)
	if err != nil {
		glog.Errorf("fail to unmarshal: %v", err)
	}
	ctx := context.Background()
	comment := *commentEvent.Comment.Body
	//split the assignees and operation to be performed.
	substrings := strings.Split(comment, "@")
	//regular expression to Assign or unassign the Assignees
	reg := regexp.MustCompile("(?mi)^/(un)?assign(( @?[-\\w]+?)*)\\s*$")
	matchAssignOrUnassign := reg.MatchString(comment)
	//list of assignees to be assigned for issues/PR
	listOfAssignees := make([]string, 0)
	//range over the substring to get the list of assignees
	for i, assignees := range substrings {
		if i == 0 {
			//first index is the operation to be performed, rest will be the assignees
			continue
		}
		listOfAssignees = append(listOfAssignees, assignees)
	}
	//operation is the assign or unassign check
	operation := strings.Trim(substrings[0], " ")

	if matchAssignOrUnassign == true {
		if operation == Assign {
			_, _, err := client.Issues.AddAssignees(ctx, *commentEvent.Repo.Owner.Login, *commentEvent.Repo.Name, *commentEvent.Issue.Number, listOfAssignees)
			if err != nil {
				glog.Fatalf("Unable to Add Assignees: %v err: %v", listOfAssignees, err)
			} else {
				glog.Infof("Assignee added successfully: %v", listOfAssignees)
			}
		} else if operation == Unassign {
			_, _, err := client.Issues.RemoveAssignees(ctx, *commentEvent.Repo.Owner.Login, *commentEvent.Repo.Name, *commentEvent.Issue.Number, listOfAssignees)
			if err != nil {
				glog.Fatalf("Cannot remove Assignees: %v err: %v", listOfAssignees, err)
			}
			glog.Infof("Removed assignee: %v", listOfAssignees)
		}
	}
}
