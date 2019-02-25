package retest

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
)

var(
	retestReg   = regexp.MustCompile("^/[Rr][Ee][Tt][Ee][Ss][Tt]")
	testReg     = regexp.MustCompile("^/[Tt][Ee][Ss][Tt]")
)

// Handle event with label
func Handle(client *github.Client, event github.IssueCommentEvent, token, repoid string) error {

	comment := *event.Comment.Body
	glog.Infof("Receive event with retest. comment: %s", comment)

	// https://github.com/islinwb/test/pull/1
	prID := strings.SplitAfter(event.Issue.PullRequestLinks.GetHTMLURL(), "github.com/")[1]

	//First index contains -islinwb/test/pull/1
	subStrings := strings.Split(prID, "/")
	len := len(subStrings)
	prNum, err:= strconv.Atoi(subStrings[len-1])
	if err != nil {
		glog.Errorf("Fail to handle: %v", err)
		return err
	}

	if retestReg.MatchString(comment) {
		// "/retest"
		err:= SendToCIForRetestAllJobs(prNum, token, repoid)
		if err != nil {
			glog.Errorf("Retest operation failed: %v", err)
			return err
		}
	} else if testReg.MatchString(comment) {
		// trigger particular job(s)
		job:= strings.Split(comment, " ")
		err:= SendToCIForTestJob(prNum, job[1], token, repoid)
		if err != nil {
			glog.Errorf("Test job failed: %v", err)
			return err
		}
	}
	return nil
}
