package label

import (
	"context"
	"regexp"
	"strings"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
)

var (
	// regular expression to add label
	RegAddLabel = regexp.MustCompile(`(?mi)^/(kind|priority)\s*(.*)$`)
	// regular expression to remove label
	RegRemoveLabel = regexp.MustCompile(`(?mi)^/remove-(kind|priority)\s*(.*)$`)
)

const (
	kind = "/kind"
	removeLabel = "/remove-kind"
)

// Get Labels from Regexp matches
func getLabelsFromREMatches(matches [][]string) (labels []string) {
	for _, match := range matches {
		for _, label := range strings.Split(match[0], " ")[1:] {
			label = strings.ToLower(match[1] + "/" + strings.TrimSpace(label))
			labels = append(labels, label)
		}
	}
	return
}
//RemoveLabelsToPR function to remove label to the PR
func RemoveLabelsToPR(ctx context.Context, prEvent github.PullRequestEvent, client *github.Client, addLable string) error{
	var Label []string

	LabelRemove := strings.TrimPrefix(addLable, "kind/")
	Label = append(Label, removeLabel, LabelRemove)

	lableBody := strings.Join(Label, " ")

	// map of add labels
	mapOfRemoveLabels := GetLabelsMap(lableBody)
	glog.Infof("map of remove labels: %v", mapOfRemoveLabels)

	// list labels in current issue
	listofIssueLabels, _, err := client.Issues.ListLabelsByIssue(ctx, *prEvent.Repo.Owner.Login, *prEvent.Repo.Name, *prEvent.Number, nil)
	if err != nil {
		glog.Fatalf("Unable to list issue labels. err: %v", err)
		return err
	}
	glog.Infof("list of issue labels: %v", listofIssueLabels)

	// list of remove labels
	listOfRemoveLabels := GetListOfRemoveLabels(mapOfRemoveLabels, listofIssueLabels)
	glog.Infof("list of remove labels: %v", listOfRemoveLabels)

	// invoke github api to remove labels
	if len(listOfRemoveLabels) > 0 {
		for _, l := range listOfRemoveLabels {
			_, err := client.Issues.RemoveLabelForIssue(ctx, *prEvent.Repo.Owner.Login, *prEvent.Repo.Name, *prEvent.Number, l)
			if err != nil {
				glog.Fatalf("unable to remove label: %v err: %v", l, err)
			} else {
				glog.Infof("remove label successfully: %v", l)
			}
		}
	} else {
		glog.Infof("No label to remove for this event")
	}

	return nil
}

//AddLabelsToPR function to add label to the PR
func AddLabelsToPR(ctx context.Context, prEvent github.PullRequestEvent, client *github.Client, addLable string) error{
	var Label []string
	listOfAddLabels := make([]string, 0)

	LabelAdd := strings.TrimPrefix(addLable, "kind/")
	Label = append(Label, kind, LabelAdd)
	lableBody := strings.Join(Label, " ")
	mapOfAddLabels := GetLabelsMap(lableBody)

	listofRepoLabels, _, err:= client.Issues.ListLabels(ctx, *prEvent.Repo.Owner.Login, *prEvent.Repo.Name, nil)
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
	listOfAddLabels = GetListOfAddLabels(mapOfAddLabels, listofRepoLabels, listofIssueLabels)
	_, _, err = client.Issues.AddLabelsToIssue(ctx, *prEvent.Repo.Owner.Login, *prEvent.Repo.Name, *prEvent.Number, listOfAddLabels)
	if err != nil {
		glog.Fatalf("Unable to add labels: %v err: %v", listOfAddLabels, err)
		return err
	} else {
		glog.Infof("Add labels successfully: %v", listOfAddLabels)
	}

	return nil
}
//HandlePRLabels function to handle add or remove label to the PR
func HandlePRLabels(ctx context.Context, prEvent github.PullRequestEvent, client *github.Client)error{
	addLabelMatches := RegAddLabel.FindAllStringSubmatch(*prEvent.PullRequest.Body, -1)
	removeLabelMatches := RegRemoveLabel.FindAllStringSubmatch(*prEvent.PullRequest.Body, -1)
	if len(addLabelMatches) == 0 && len(removeLabelMatches) == 0{
		return nil
	}

	//get all labels from submatch and store in slice
	labelsToAdd := append(getLabelsFromREMatches(addLabelMatches))
	labelsToRemove := append(getLabelsFromREMatches(removeLabelMatches))

	for i,_ := range labelsToAdd {
		err := AddLabelsToPR(ctx, prEvent, client, labelsToAdd[i])
		if err != nil {
			glog.Fatalf("Unable to list issue labels. err: %v", err)
		}
	}

	for i,_ := range labelsToRemove {
		err := RemoveLabelsToPR(ctx, prEvent, client, labelsToRemove[i])
		if err != nil {
			glog.Fatalf("Unable to list issue labels. err: %v", err)
		}
	}
	return nil
}

// Handle event with label
func Handle(client *github.Client, event github.IssueCommentEvent) error {
	// get basic params
	comment := *event.Comment.Body
	glog.Infof("receive event with label. comment: %s", comment)

	// add labels
	if RegAddLabel.MatchString(comment) {
		return Add(client, event)
	}
	// remove labels
	if RegRemoveLabel.MatchString(comment) {
		return Remove(client, event)
	}

	return nil
}

// add labels
func Add(client *github.Client, event github.IssueCommentEvent) error {
	// get basic params
	ctx := context.Background()
	comment := *event.Comment.Body
	owner := *event.Repo.Owner.Login
	repo := *event.Repo.Name
	number := *event.Issue.Number
	glog.Infof("add label started. comment: %s owner: %s repo: %s number: %d", comment, owner, repo, number)

	//	/kind label1
	//	/kind lable2 format handling
	getLables := strings.Split(comment, "\r\n")

	for _, labelToAdd := range getLables{
		// map of add labels
		mapOfAddLabels := GetLabelsMap(labelToAdd)
		glog.Infof("map of add labels: %v", mapOfAddLabels)

		// list labels in current github repository
		listofRepoLabels, _, err := client.Issues.ListLabels(ctx, owner, repo, nil)
		if err != nil {
			glog.Fatalf("unable to list repository labels. err: %v", err)
			return err
		}
		glog.Infof("list of repository labels: %v", listofRepoLabels)

		// list labels in current issue
		listofIssueLabels, _, err := client.Issues.ListLabelsByIssue(ctx, owner, repo, number, nil)
		if err != nil {
			glog.Fatalf("unable to list issue labels. err: %v", err)
			return err
		}
		glog.Infof("list of issue labels: %v", listofIssueLabels)

		// list of add labels
		listOfAddLabels := GetListOfAddLabels(mapOfAddLabels, listofRepoLabels, listofIssueLabels)
		glog.Infof("list of add labels: %v", listOfAddLabels)

		// invoke github api to add labels
		if len(listOfAddLabels) > 0 {
			_, _, err := client.Issues.AddLabelsToIssue(ctx, owner, repo, number, listOfAddLabels)
			if err != nil {
				glog.Fatalf("unable to add labels: %v err: %v", listOfAddLabels, err)
				return err
			} else {
				glog.Infof("add labels successfully: %v", listOfAddLabels)
			}
		} else {
			glog.Infof("No label to add for this event")
		}
	}

	return nil
}

// remove labels
func Remove(client *github.Client, event github.IssueCommentEvent) error {
	// get basic params
	ctx := context.Background()
	comment := *event.Comment.Body
	owner := *event.Repo.Owner.Login
	repo := *event.Repo.Name
	number := *event.Issue.Number
	glog.Infof("remove label started. comment: %s owner: %s repo: %s number: %d", comment, owner, repo, number)

	getLables := strings.Split(comment, "\r\n")

	for _, labelToRemove := range getLables{
		// map of add labels
		mapOfRemoveLabels := GetLabelsMap(labelToRemove)
		glog.Infof("map of remove labels: %v", mapOfRemoveLabels)

		// list labels in current issue
		listofIssueLabels, _, err := client.Issues.ListLabelsByIssue(ctx, owner, repo, number, nil)
		if err != nil {
			glog.Fatalf("unable to list issue labels. err: %v", err)
			return err
		}
		glog.Infof("list of issue labels: %v", listofIssueLabels)

		// list of remove labels
		listOfRemoveLabels := GetListOfRemoveLabels(mapOfRemoveLabels, listofIssueLabels)
		glog.Infof("list of remove labels: %v", listOfRemoveLabels)

		// invoke github api to remove labels
		if len(listOfRemoveLabels) > 0 {
			for _, l := range listOfRemoveLabels {
				_, err := client.Issues.RemoveLabelForIssue(ctx, owner, repo, number, l)
				if err != nil {
					glog.Fatalf("unable to remove label: %v err: %v", l, err)
				} else {
					glog.Infof("remove label successfully: %v", l)
				}
			}
		} else {
			glog.Infof("No label to remove for this event")
		}
	}

	return nil
}

// getListOfAddLabels return the exact list of add labels
func GetListOfAddLabels(mapOfAddLabels map[string]string, listofRepoLabels []*github.Label, listofIssueLabels []*github.Label) []string {
	// init
	listOfAddLabels := make([]string, 0)
	// range over the map to filter the list of labels
	for l := range mapOfAddLabels {
		// check if the label is existing in current github repository
		existingInRepo := false
		for _, repoLabel := range listofRepoLabels {
			if l == *repoLabel.Name {
				existingInRepo = true
				break
			}
		}
		// the label is not existing in current github repository so it can not add this label
		if !existingInRepo {
			glog.Infof("label %s is not existing in repository", l)
			continue
		}

		// check if the label is existing in current issue
		existingInIssue := false
		for _, issueLabel := range listofIssueLabels {
			if l == *issueLabel.Name {
				existingInIssue = true
				break
			}
		}
		// the label is already existing in current issue so it is no need to add this label
		if existingInIssue {
			glog.Infof("label %s is already existing in current issue", l)
			continue
		}

		// append
		listOfAddLabels = append(listOfAddLabels, l)
	}
	return listOfAddLabels
}

// getListOfRemoveLabels return the exact list of remove labels
func GetListOfRemoveLabels(mapOfRemoveLabels map[string]string, listofIssueLabels []*github.Label) []string {
	// init
	listOfRemoveLabels := make([]string, 0)
	// range over the map to filter the list of labels
	for l := range mapOfRemoveLabels {
		// check if the label is existing in current issue
		existingInIssue := false
		for _, issueLabel := range listofIssueLabels {
			if l == *issueLabel.Name {
				existingInIssue = true
				break
			}
		}
		// the label is not existing in current issue so it is no need to remove this label
		if !existingInIssue {
			glog.Infof("label %s is not existing in current issue", l)
			continue
		}

		// append
		listOfRemoveLabels = append(listOfRemoveLabels, l)
	}
	return listOfRemoveLabels
}

// getLabelsMap for add or remove labels
func GetLabelsMap(comment string) map[string]string {
	// init labels map
	mapOfLabels := map[string]string{}
	// split with blank space
	substrings := strings.Split(strings.TrimSpace(comment), " ")
	// init label group
	labelGroup := ""
	// range over the substrings to get the map of labels
	for i, l := range substrings {
		if i == 0 {
			// first index is the operation to be performed, rest will be the labels
			// the label group. e.g kind, priority
			labelGroup = strings.Replace(strings.Replace(l, "/", "", 1), "remove-", "", 1)
		} else {
			// the whole label = label group + / + label. e.g kind/feature
			wholeLabel := labelGroup + "/" + l
			// use map to avoid the reduplicate label
			mapOfLabels[wholeLabel] = wholeLabel
		}
	}
	return mapOfLabels
}
