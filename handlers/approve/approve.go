package approve

import (
	"context"
	"regexp"

	"github.com/golang/glog"
	"github.com/google/go-github/github"

	"github.com/Huawei-PaaS/ci-bot/handlers/repository"
	"github.com/Huawei-PaaS/ci-bot/handlers/util"
)

var (
	// approve label name
	LabelNameApproved = util.LabelNameApproved
	// regular expression to add approve
	RegAddApprove = regexp.MustCompile(`(?mi)^/approve\s*$`)
	// regular expression to cancel approve
	RegCancelApprove = regexp.MustCompile(`(?mi)^/approve cancel\s*$`)
)

// Handle event with approve
func Handle(client *github.Client, r repository.Interface, event github.IssueCommentEvent) error {
	// only handle pr which is open
	if event.Issue.IsPullRequest() && *event.Issue.State == "open" {
		// get basic params
		comment := *event.Comment.Body
		glog.Infof("Receive event with approve. Comment: %s", comment)

		// add approved label
		if RegAddApprove.MatchString(comment) {
			return Add(client, r, event)
		}
		// remove approved label
		if RegCancelApprove.MatchString(comment) {
			return Cancel(client, r, event)
		}
	}
	return nil
}

// Add approved label
func Add(client *github.Client, r repository.Interface, event github.IssueCommentEvent) error {
	// get basic params
	ctx := context.Background()
	comment := *event.Comment.Body
	commentAuthor := *event.Comment.User.Login
	owner := *event.Repo.Owner.Login
	repo := *event.Repo.Name
	number := *event.Issue.Number
	glog.Infof("Add approve started. Comment: %s commentAuthor: %s owner: %s repo: %s number: %d",
		comment, commentAuthor, owner, repo, number)

	// check if current author is collaborator
	IsCollaborator, _, err := client.Repositories.IsCollaborator(ctx, owner, repo, commentAuthor)
	if err != nil {
		glog.Fatalf("Unable to check if current author is collaborator. err: %v", err)
		return err
	}
	// not collaborator
	if !IsCollaborator {
		issuelistCommentsOptions := &github.IssueListCommentsOptions{Sort: "created", Direction: "asc"}
		issueComments, _, err := client.Issues.ListComments(ctx, owner, repo, number, issuelistCommentsOptions)
		if err != nil {
			glog.Fatalf("Unable to list issue comments. err: %v", issueComments)
			return err
		}

		// init current approvers map
		mapOfApprovers := map[string]string{}
		for _, ic := range issueComments {
			// add approvers
			if RegAddApprove.MatchString(*ic.Body) {
				// add approver in map
				mapOfApprovers[*ic.User.Login] = *ic.User.Login
			}
			// cancel approvers
			if RegCancelApprove.MatchString(*ic.Body) {
				// delete approver in map
				if _, ok := mapOfApprovers[*ic.User.Login]; ok {
					delete(mapOfApprovers, *ic.User.Login)
				}
			}
		}
		// the last comment event does not including in the result of list issue comments
		// so it can be added here
		mapOfApprovers[commentAuthor] = commentAuthor
		glog.Infof("Current map of approvers: %v", mapOfApprovers)

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

		// init approved path map
		mapOfApprovedPath := make(map[string]map[string]string)
		listOfUnapprovedPath := make([]string, 0)
		for _, path := range listOfFileNames {
			// get all approvers by path
			allApprovers := r.GetAllApprovers(path)
			glog.Infof("Path: %s AllApprovers: %v", path, allApprovers)

			// init map by path
			if mapOfApprovedPath[path] == nil {
				mapOfApprovedPath[path] = make(map[string]string)
			}

			// the current approvers are in the approvers of path
			for k, v := range mapOfApprovers {
				if _, ok := allApprovers[k]; ok {
					mapOfApprovedPath[path][k] = v
				}
			}

			// could not find the approver of path
			if len(mapOfApprovedPath[path]) == 0 {
				listOfUnapprovedPath = append(listOfUnapprovedPath, path)
			}
		}
		glog.Infof("Map of approved path: %v", mapOfApprovedPath)

		// unapproved path is existing
		if len(listOfUnapprovedPath) > 0 {
			glog.Infof("Unapproved path is existing: %v", listOfUnapprovedPath)
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

	// check if it is approved
	hasApproved := false
	for _, l := range listofIssueLabels {
		if *l.Name == LabelNameApproved {
			hasApproved = true
			break
		}
	}

	// not approved
	if !hasApproved {
		// add label approved
		listOfAddLabels := []string{LabelNameApproved}
		_, _, err := client.Issues.AddLabelsToIssue(ctx, owner, repo, number, listOfAddLabels)
		if err != nil {
			glog.Fatalf("Unable to add label: %v err: %v", listOfAddLabels, err)
			return err
		} else {
			glog.Infof("Add label successfully: %v", listOfAddLabels)
		}
	} else {
		glog.Infof("No label to add: %v", LabelNameApproved)
	}

	// try to merge pr
	err = util.MergePullRequest(client, owner, repo, number)
	if err != nil {
		glog.Errorf("Unable to merge pr: #%d err: %v", number, err)
		return err
	}
	return nil
}

// Cancel removes approved label
func Cancel(client *github.Client, r repository.Interface, event github.IssueCommentEvent) error {
	// get basic params
	ctx := context.Background()
	comment := *event.Comment.Body
	commentAuthor := *event.Comment.User.Login
	owner := *event.Repo.Owner.Login
	repo := *event.Repo.Name
	number := *event.Issue.Number
	glog.Infof("Cancel approve started. Comment: %s commentAuthor: %s owner: %s repo: %s number: %d",
		comment, commentAuthor, owner, repo, number)

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

		// init owners
		err = r.LoadOwners(*pr.Base.Ref)
		if err != nil {
			glog.Fatalf("Unable to load owners. err: %v", err)
			return err
		}

		// check if the owner is existing in the approvers of path
		IsApprover := false
		for _, path := range listOfFileNames {
			// get all approvers by path
			allApprovers := r.GetAllApprovers(path)
			glog.Infof("Path: %s AllApprovers: %v", path, allApprovers)

			// owner is existing in the approvers of path
			if _, ok := allApprovers[commentAuthor]; ok {
				IsApprover = true
				glog.Infof("Approver commentAuthor: %s path: %s", commentAuthor, path)
				break
			}
		}

		// not approver
		if !IsApprover {
			glog.Infof("Owner is not existing in the approvers of path")
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

	// check if it is approved
	hasApproved := false
	for _, l := range listofIssueLabels {
		if *l.Name == LabelNameApproved {
			hasApproved = true
			break
		}
	}

	// approved
	if hasApproved {
		// remove label approved
		_, err := client.Issues.RemoveLabelForIssue(ctx, owner, repo, number, LabelNameApproved)
		if err != nil {
			glog.Fatalf("Unable to remove label: %v err: %v", LabelNameApproved, err)
		} else {
			glog.Infof("Remove label successfully: %v", LabelNameApproved)
		}
	} else {
		glog.Infof("No label to remove: %v", LabelNameApproved)
	}

	return nil
}
