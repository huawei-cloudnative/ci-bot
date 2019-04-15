package handlers

import (
	"encoding/json"

	"github.com/Huawei-PaaS/ci-bot/handlers/approve"
	"github.com/Huawei-PaaS/ci-bot/handlers/assign"
	"github.com/Huawei-PaaS/ci-bot/handlers/label"
	"github.com/Huawei-PaaS/ci-bot/handlers/lgtm"
	"github.com/Huawei-PaaS/ci-bot/handlers/repository"
	"github.com/Huawei-PaaS/ci-bot/handlers/retest"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
)

type GithubIssue github.Issue

func (s *Server) handleIssueEvent(body []byte) {
	glog.Info("Received an Issue Event")
}

//function to handle issue comments
func (s *Server) handleIssueCommentEvent(body []byte, client *github.Client, r repository.Interface) {
	var commentEvent github.IssueCommentEvent

	// Unmarshal
	err := json.Unmarshal(body, &commentEvent)
	if err != nil {
		glog.Errorf("Failed to unmarshal commentEvent: %v", err)
	}
	// label
	if label.RegAddLabel.MatchString(*commentEvent.Comment.Body) || label.RegRemoveLabel.MatchString(*commentEvent.Comment.Body) {
		err = label.Handle(client, commentEvent)
		if err != nil {
			glog.Errorf("Failed to handle label: %v", err)
		}
	}
	// assign
	if AssignOrUnassing.MatchString(*commentEvent.Comment.Body) {
		err = assign.Handle(client, commentEvent)
		if err != nil {
			glog.Errorf("Failed to handle assign: %v", err)
		}
	}
	// retest
	if TestReg.MatchString(*commentEvent.Comment.Body) || RetestReg.MatchString(*commentEvent.Comment.Body) {
		err = retest.Handle(client, commentEvent, s.Config.TravisCIToken, s.Config.TravisRepoName)
		if err != nil {
			glog.Errorf("Failed to handle retest: %v", err)
		}
	}

	// approve
	if approve.RegAddApprove.MatchString(*commentEvent.Comment.Body) || approve.RegCancelApprove.MatchString(*commentEvent.Comment.Body) {
		err = approve.Handle(client, r, commentEvent)
		if err != nil {
			glog.Errorf("Failed to handle: %v", err)
		}
	}

	// lgtm
	if lgtm.RegAddLgtm.MatchString(*commentEvent.Comment.Body) || lgtm.RegCancelLgtm.MatchString(*commentEvent.Comment.Body) {
		err = lgtm.Handle(client, r, commentEvent)
		if err != nil {
			glog.Errorf("Failed to handle: %v", err)
		}
	}

	// reviewers
	if assign.CCRegExp.MatchString(*commentEvent.Comment.Body){
		err = assign.ReviewerReqByComment(client, commentEvent)
		if err != nil {
			glog.Errorf("Failed to handle: %v", err)
		}
	}
}
