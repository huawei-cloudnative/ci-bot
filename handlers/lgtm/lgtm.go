package lgtm

import (
	"context"
	"regexp"

	"github.com/golang/glog"
	"github.com/google/go-github/github"

	"github.com/Huawei-PaaS/ci-bot/handlers/repository"
	"github.com/Huawei-PaaS/ci-bot/handlers/util"
)

var (
	// lgtm label name
	LabelNameLgtm = util.LabelNameLgtm
	// regular expression to add lgtm
	RegAddLgtm = regexp.MustCompile(`(?mi)^/lgtm\s*$`)
	// regular expression to cancel lgtm
	RegCancelLgtm = regexp.MustCompile(`(?mi)^/lgtm cancel\s*$`)
)

// Handle event with lgtm
func Handle(client *github.Client, r repository.Interface, event github.IssueCommentEvent) error {
	// only handle pr which is open
	if event.Issue.IsPullRequest() && *event.Issue.State == "open" {
		// get basic params
		comment := *event.Comment.Body
		glog.Infof("Receive event with lgtm. comment: %s", comment)

		// add lgtm label
		if RegAddLgtm.MatchString(comment) {
			return Add(client, r, event)
		}
		// remove lgtm label
		if RegCancelLgtm.MatchString(comment) {
			return Cancel(client, r, event)
		}
	}
	return nil
}

// Add lgtm label
func Add(client *github.Client, r repository.Interface, event github.IssueCommentEvent) error {
	// get basic params
	ctx := context.Background()
	comment := *event.Comment.Body
	issueAuthor := *event.Issue.User.Login
	commentAuthor := *event.Comment.User.Login
	owner := *event.Repo.Owner.Login
	repo := *event.Repo.Name
	number := *event.Issue.Number
	glog.Infof("Add lgtm started. Comment: %s issueAuthor: %s commentAuthor: %s owner: %s repo: %s number: %d",
		comment, issueAuthor, commentAuthor, owner, repo, number)

	// can not lgtm on self-own pr
	if issueAuthor == commentAuthor {
		glog.Info("can not lgtm on self-own pr")
		return nil
	}

	// check if current author is collaborator
	IsCollaborator, _, err := client.Repositories.IsCollaborator(ctx, owner, repo, commentAuthor)
	if err != nil {
		glog.Fatalf("Unable to check if current author is collaborator. err: %v", err)
		return err
	}
	// not collaborator
	if !IsCollaborator {
		// list file names in current pr e.g. test/hello.go
		prChangedFiles, _, err := client.PullRequests.ListFiles(ctx, owner, repo, number, nil)
		if err != nil {
			glog.Fatalf("Unable to list pr changed files. err: %v", err)
			return err
		}
		listOfFileNames := make([]string, 0)
		for _, f := range prChangedFiles {
			listOfFileNames = append(listOfFileNames, *f.Filename)
		}
		glog.Infof("List of pr file names: %v", listOfFileNames)

		// e.g. master
		pr, _, err := client.PullRequests.Get(ctx, owner, repo, number)
		glog.Infof("Pr base ref: %v", *pr.Base.Ref)

		// load owners
		err = r.LoadOwners(*pr.Base.Ref)
		if err != nil {
			glog.Fatalf("Unable to load owners. err: %v", err)
			return err
		}

		// get all approvers and reviewers
		mapOfOwners := make(map[string]string)
		for _, path := range listOfFileNames {
			// get all approvers by path
			allApprovers := r.GetAllApprovers(path)
			glog.Infof("Path: %s AllApprovers: %v", path, allApprovers)
			for k, v := range allApprovers {
				mapOfOwners[k] = v
			}
			// get all reviewers by path
			allReviewers := r.GetAllReviewers(path)
			glog.Infof("Path: %s AllReviewers: %v", path, allReviewers)
			for k, v := range allReviewers {
				mapOfOwners[k] = v
			}
		}
		glog.Infof("Map of owners: %v", mapOfOwners)

		// can not find from the owners
		if _, ok := mapOfOwners[commentAuthor]; !ok {
			glog.Infof("can not find %s from owners", commentAuthor)
			return nil
		}
	} else {
		glog.Infof("Current author %s is collaborator", commentAuthor)
	}

	// list labels in current issue
	listofIssueLabels, _, err := client.Issues.ListLabelsByIssue(ctx, owner, repo, number, nil)
	if err != nil {
		glog.Fatalf("Unable to list issue labels. err: %v", err)
		return err
	}
	glog.Infof("List of issue labels: %v", listofIssueLabels)

	// check if it has lgtm
	hasLgtm := false
	for _, l := range listofIssueLabels {
		if *l.Name == LabelNameLgtm {
			hasLgtm = true
			break
		}
	}

	// it has no lgtm
	if !hasLgtm {
		// add label lgtm
		listOfAddLabels := []string{LabelNameLgtm}
		_, _, err := client.Issues.AddLabelsToIssue(ctx, owner, repo, number, listOfAddLabels)
		if err != nil {
			glog.Fatalf("Unable to add label: %v err: %v", listOfAddLabels, err)
			return err
		} else {
			glog.Infof("Add label successfully: %v", listOfAddLabels)
		}
	} else {
		glog.Infof("No label to add: %v", LabelNameLgtm)
	}

	// try to merge pr
	err = util.MergePullRequest(client, owner, repo, number)
	if err != nil {
		glog.Errorf("Unable to merge pr: #%d err: %v", number, err)
		return err
	}

	return nil
}

// Cancel removes lgtm label
func Cancel(client *github.Client, r repository.Interface, event github.IssueCommentEvent) error {
	// get basic params
	ctx := context.Background()
	comment := *event.Comment.Body
	issueAuthor := *event.Issue.User.Login
	commentAuthor := *event.Comment.User.Login
	owner := *event.Repo.Owner.Login
	repo := *event.Repo.Name
	number := *event.Issue.Number
	glog.Infof("Cancel lgtm started. Comment: %s issueAuthor: %s commentAuthor: %s owner: %s repo: %s number: %d",
		comment, issueAuthor, commentAuthor, owner, repo, number)

	// can cancel lgtm on self-own pr
	if issueAuthor != commentAuthor {
		// check if current author is collaborator
		IsCollaborator, _, err := client.Repositories.IsCollaborator(ctx, owner, repo, commentAuthor)
		if err != nil {
			glog.Fatalf("Unable to check if current author is collaborator. err: %v", err)
			return err
		}
		// Not collaborator
		if !IsCollaborator {
			// list file names in current pr e.g. test/hello.go
			prChangedFiles, _, err := client.PullRequests.ListFiles(ctx, owner, repo, number, nil)
			if err != nil {
				glog.Fatalf("Unable to list pr changed files. err: %v", err)
				return err
			}
			listOfFileNames := make([]string, 0)
			for _, f := range prChangedFiles {
				listOfFileNames = append(listOfFileNames, *f.Filename)
			}
			glog.Infof("List of pr file names: %v", listOfFileNames)

			// e.g. master
			pr, _, err := client.PullRequests.Get(ctx, owner, repo, number)
			glog.Infof("Pr base ref: %v", *pr.Base.Ref)

			// load owners
			err = r.LoadOwners(*pr.Base.Ref)
			if err != nil {
				glog.Fatalf("Unable to load owners. err: %v", err)
				return err
			}

			// get all approvers and reviewers
			mapOfOwners := make(map[string]string)
			for _, path := range listOfFileNames {
				// get all approvers by path
				allApprovers := r.GetAllApprovers(path)
				glog.Infof("Path: %s AllApprovers: %v", path, allApprovers)
				for k, v := range allApprovers {
					mapOfOwners[k] = v
				}
				// get all reviewers by path
				allReviewers := r.GetAllReviewers(path)
				glog.Infof("Path: %s AllReviewers: %v", path, allReviewers)
				for k, v := range allReviewers {
					mapOfOwners[k] = v
				}
			}
			glog.Infof("Map of owners: %v", mapOfOwners)

			// can not find from the owners
			if _, ok := mapOfOwners[commentAuthor]; !ok {
				glog.Infof("can not find %s from owners", commentAuthor)
				return nil
			}
		} else {
			glog.Infof("Current author %s is collaborator", commentAuthor)
		}
	} else {
		glog.Infof("Cancel lgtm on %s self-own pr", commentAuthor)
	}

	// list labels in current issue
	listofIssueLabels, _, err := client.Issues.ListLabelsByIssue(ctx, owner, repo, number, nil)
	if err != nil {
		glog.Fatalf("Unable to list issue labels. err: %v", err)
		return err
	}
	glog.Infof("List of issue labels: %v", listofIssueLabels)

	// check if it has lgtm
	hasLgtm := false
	for _, l := range listofIssueLabels {
		if *l.Name == LabelNameLgtm {
			hasLgtm = true
			break
		}
	}

	// it has no lgtm
	if hasLgtm {
		// remove label lgtm
		_, err := client.Issues.RemoveLabelForIssue(ctx, owner, repo, number, LabelNameLgtm)
		if err != nil {
			glog.Fatalf("Unable to remove label: %v err: %v", LabelNameLgtm, err)
		} else {
			glog.Infof("Remove label successfully: %v", LabelNameLgtm)
		}
	} else {
		glog.Infof("No label to remove: %v", LabelNameLgtm)
	}

	return nil
}
