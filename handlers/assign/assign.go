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

var (
	AssignRegExp = regexp.MustCompile(`(?mi)^/(un)?assign(( @?[-\w]+?)*)\s*$`)
	// ccRegexp parses and validates /cc commands, also used by blunderbuss
	CCRegExp = regexp.MustCompile(`(?mi)^/(un)?cc(( +@?[-/\w]+?)*)\s*$`)
)
//parseLogins function to parse the login id's
func parseLogins(text string) []string {
	var parts []string
	for _, p := range strings.Split(text, " ") {
		t := strings.Trim(p, "@ ")
		if t == "" {
			continue
		}
		parts = append(parts, t)
	}
	return parts
}
//AddAssignee function to add assignee to the PR
func AddAssignee(ctx context.Context, prEvent github.PullRequestEvent, client *github.Client, listOfAssignees []string) error {
	_, _, err := client.Issues.AddAssignees(ctx, *prEvent.Repo.Owner.Login, *prEvent.Repo.Name, *prEvent.Number, listOfAssignees)
	if err != nil {
		glog.Fatalf("Unable to Add Assignees: %v err: %v", listOfAssignees, err)
		return err
	} else {
		glog.Infof("Assignee added successfully: %v", listOfAssignees)
	}

	return nil
}
//RemoveReviewer function to remove the reviewer to the PR
func RemoveReviewer(ctx context.Context, login, repoName string, prNum int, client *github.Client, listOfAssignees []string) error {
	var reviewersList github.ReviewersRequest
	reviewersList.Reviewers = listOfAssignees
	  _, err := client.PullRequests.RemoveReviewers(ctx, login, repoName, prNum, reviewersList)
	if err != nil {
		glog.Fatalf("Cannot remove Reviewers: %v err: %v", listOfAssignees, err)
		return err
	}
	glog.Infof("Removed Reviewers: %v", listOfAssignees)
	return nil
}
//AddReviewer function to add the reviewer to the PR
func AddReviewer(ctx context.Context, login, repoName string, prNum int, client *github.Client, listOfAssignees []string) error {
	var reviewersList github.ReviewersRequest
	var listOpt github.ListOptions
	var ExistingList, revieweReqList []string

	ListRepoReviewers, _, err := client.PullRequests.ListReviewers(ctx, login, repoName, prNum, &listOpt)
	if err != nil {
		glog.Fatalf("Unable to get the review list : err: %v", err)
		return err
	}
	//check if the requested reviewer is already been assigned as reviewer
	if len(ListRepoReviewers.Users) > 0{
		for _, repoReviewer := range ListRepoReviewers.Users {
			for i, _ := range listOfAssignees {
				if listOfAssignees[i] == *repoReviewer.Login{
					ExistingList = append(ExistingList, listOfAssignees[i])
				}else{
					revieweReqList = append(revieweReqList, listOfAssignees[i])
				}
			}
		}
		reviewersList.Reviewers = revieweReqList
		if len(revieweReqList) == 0{
			glog.Infof("Reviewers already added to this PR: %v", ExistingList)
			return nil
		}
	}else{
		reviewersList.Reviewers = listOfAssignees
	}

	_, _, err = client.PullRequests.RequestReviewers(ctx, login, repoName, prNum, reviewersList)
	if err != nil {
		glog.Fatalf("Unable to Add Reviewers: %v err: %v", listOfAssignees, err)
		return err
	} else {
		glog.Infof("Reviewers added successfully: %v", listOfAssignees)
	}

	return nil
}
//RemoveAssignee function to remove the assignee to the PR
func RemoveAssignee(ctx context.Context, prEvent github.PullRequestEvent, client *github.Client, listOfAssignees []string) error {
	_, _, err := client.Issues.RemoveAssignees(ctx, *prEvent.Repo.Owner.Login, *prEvent.Repo.Name, *prEvent.Number, listOfAssignees)
	if err != nil {
		glog.Fatalf("Cannot remove Assignees: %v err: %v", listOfAssignees, err)
		return err
	}
	glog.Infof("Removed assignee: %v", listOfAssignees)
	return nil
}
//GetMatchList to get the list of add and remove assignees
func GetMatchList(login string, matchesList [][]string)([]string, []string){
	users := make(map[string]bool)
	for _, re := range matchesList {
		add := re[1] != "un" // un<cmd> == !add
		if re[2] == "" {
			users[login] = add
		} else {
			for _, login := range parseLogins(re[2]) {
				users[login] = add
			}
		}
	}
	var toAdd, toRemove []string
	for login, add := range users {
		if add {
			toAdd = append(toAdd, login)
		} else {
			toRemove = append(toRemove, login)
		}
	}

	return toAdd, toRemove
}
//HandlePRAssign function to add assignee to the PR
func HandlePRAssign(ctx context.Context, prEvent github.PullRequestEvent, client *github.Client) error {
	//Get all matching assignee list for the PR Body
	assigneeMatches := AssignRegExp.FindAllStringSubmatch(*prEvent.PullRequest.Body, -1)
	toAdd, toRemove := GetMatchList(*prEvent.PullRequest.User.Login, assigneeMatches)

	if len(toAdd) > 0 {
		glog.Infof("Going to add assignees %v:", toAdd)
		err := AddAssignee(ctx, prEvent, client, toAdd)
		if err != nil {
			glog.Fatalf("AddAssignee is failed err: %v", err)
			return err
		}
	}
	if len(toRemove) > 0 {
		glog.Infof("Going to add remove assignees %v:", toRemove)
		err := RemoveAssignee(ctx, prEvent, client, toRemove)
		if err != nil {
			glog.Fatalf("RemoveAssignee is failed err: %v", err)
			return err
		}
	}
	return nil
}
//HandlePRReviewer to handle add and remove reviewers to the PR
func HandlePRReviewer(ctx context.Context, prEvent github.PullRequestEvent, client *github.Client) error {
	//Get all matching assignee list for the PR Body
	reviewMatches := CCRegExp.FindAllStringSubmatch(*prEvent.PullRequest.Body, -1)
	toAdd, toRemove := GetMatchList(*prEvent.PullRequest.User.Login, reviewMatches)

	login := *prEvent.Repo.Owner.Login
	repoName := *prEvent.Repo.Name
	prNum:= *prEvent.Number

	if len(toAdd) > 0 {
		glog.Infof("Going to add assign reviewer %v:", toAdd)
		err := AddReviewer(ctx, login, repoName, prNum, client, toAdd)
		if err != nil {
			glog.Fatalf("Adding reviewer is failed err: %v", err)
			return err
		}
	}
	if len(toRemove) > 0 {
		glog.Infof("Going to add remove assign reviewer %v:", toRemove)
		err := RemoveReviewer(ctx, login, repoName, prNum, client, toRemove)
		if err != nil {
			glog.Fatalf("Remove reviewer is failed err: %v", err)
			return err
		}
	}
	return nil
}
//HandlePRReviewer to handle add and remove reviewers to the PR
func ReviewerReqByComment(client *github.Client, event github.IssueCommentEvent) error{
	ctx := context.Background()
	login := *event.Repo.Owner.Login
	repoName := *event.Repo.Name
	IssueNum:= *event.Issue.Number

	assigneeMatches := CCRegExp.FindAllStringSubmatch(*event.Comment.Body, -1)
	toAdd, toRemove := GetMatchList(*event.Comment.User.Login, assigneeMatches)

	if len(toAdd) > 0 {
		glog.Infof("Going to add reviewer from comment section%v:", toAdd)
		err := AddReviewer(ctx, login, repoName, IssueNum, client, toAdd)
		if err != nil {
			glog.Fatalf("Adding reviewer is failed err: %v", err)
			return err
		}
	}
	if len(toRemove) > 0 {
		glog.Infof("Going to remove reviewer from comment section %v:", toRemove)
		err := RemoveReviewer(ctx, login, repoName, IssueNum, client, toRemove)
		if err != nil {
			glog.Fatalf("Remove reviewer is failed err: %v", err)
			return err
		}
	}
	return nil
}


// Handle event with assign
func Handle(client *github.Client, event github.IssueCommentEvent) error {
	comment := *event.Comment.Body
	//regular expression to Assign or unassign the Assignees
	reg := regexp.MustCompile("(?mi)^/(un)?assign(( @?[-\\w]+?)*)\\s*$")

	if reg.MatchString(comment) {
		ctx := context.Background()
		//check if multiple /assign are exist
		getAssignees := strings.Split(comment, "\r\n")
		for _, assignee := range getAssignees{
			//split the assignees and operation to be performed.
			substrings := strings.Split(assignee, "@")
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

	}
	return nil
}
