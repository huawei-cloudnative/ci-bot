package assign

import (
	"context"
	"regexp"
	"strings"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
)

const (
	Assign   = "/assign"
	Unassign = "/unassign"
)

// Handle event with assign
func Handle(client *github.Client, event github.IssueCommentEvent) error {
	comment := *event.Comment.Body
	//regular expression to Assign or unassign the Assignees
	reg := regexp.MustCompile("(?mi)^/(un)?assign(( @?[-\\w]+?)*)\\s*$")

	if reg.MatchString(comment) {
		ctx := context.Background()
		//split the assignees and operation to be performed.
		substrings := strings.Split(comment, "@")
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
		if operation == Assign {
			_, _, err := client.Issues.AddAssignees(ctx, *event.Repo.Owner.Login, *event.Repo.Name, *event.Issue.Number, listOfAssignees)
			if err != nil {
				glog.Fatalf("Unable to Add Assignees: %v err: %v", listOfAssignees, err)
				return err
			} else {
				glog.Infof("Assignee added successfully: %v", listOfAssignees)
			}
		} else if operation == Unassign {
			_, _, err := client.Issues.RemoveAssignees(ctx, *event.Repo.Owner.Login, *event.Repo.Name, *event.Issue.Number, listOfAssignees)
			if err != nil {
				glog.Fatalf("Cannot remove Assignees: %v err: %v", listOfAssignees, err)
				return err
			}
			glog.Infof("Removed assignee: %v", listOfAssignees)
		}
	}
	return nil
}
