package util

import (
	"context"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
)

var (
	// approve label name
	LabelNameApproved = "approved"
	// lgtm label name
	LabelNameLgtm = "lgtm"
)

// MergePullRequest with approved and lgtm label
func MergePullRequest(client *github.Client, owner string, repo string, number int) error {
	glog.Infof("Merge pr started. owner: %s repo: %s number: %d", owner, repo, number)

	// list labels in current pr
	ctx := context.Background()
	listofPrLabels, _, err := client.Issues.ListLabelsByIssue(ctx, owner, repo, number, nil)
	if err != nil {
		glog.Fatalf("Unable to list pr labels. err: %v", err)
		return err
	}
	glog.Infof("List of pr labels: %v", listofPrLabels)

	// check if it has both approved and lgtm label
	hasApproved := false
	hasLgtm := false
	for _, l := range listofPrLabels {
		if *l.Name == LabelNameApproved {
			hasApproved = true
		} else if *l.Name == LabelNameLgtm {
			hasLgtm = true
		}
	}
	glog.Infof("Pr labels have approved: %t lgtm: %t", hasApproved, hasLgtm)

	// ready to merge
	if hasApproved && hasLgtm {
		// get commit message
		pr, _, err := client.PullRequests.Get(ctx, owner, repo, number)
		commitMessage := *pr.Title
		glog.Infof("Commit message: %s", commitMessage)

		// merge pr
		result, _, err := client.PullRequests.Merge(ctx, owner, repo, number, commitMessage, nil)
		if err != nil {
			glog.Errorf("Unable to merge pr: #%d err: %v", number, err)
			return err
		}

		// skip nil
		if result == nil {
			glog.Errorf("The merge result of pr #%d is nil", number)
			return err
		}

		// check merge result
		if !*result.Merged {
			glog.Errorf("Failed to merge pr #%d. Message: %s", number, *result.Message)
		} else {
			glog.Infof("Merge pr #%d successfully", number)
		}
	}

	return nil
}
