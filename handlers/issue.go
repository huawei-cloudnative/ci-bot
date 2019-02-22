package handlers

import (
	"encoding/json"

	"github.com/golang/glog"
	"github.com/google/go-github/github"

	"github.com/Huawei-PaaS/ci-bot/handlers/assign"
	"github.com/Huawei-PaaS/ci-bot/handlers/label"
)

type GithubIssue github.Issue

func (s *Server) handleIssueEvent(body []byte) {
	glog.Info("Received an Issue Event")
}

//function to handle issue comments
func (s *Server) handleIssueCommentEvent(body []byte, client *github.Client) {
	var commentEvent github.IssueCommentEvent

	// Unmarshal
	err := json.Unmarshal(body, &commentEvent)
	if err != nil {
		glog.Errorf("fail to unmarshal: %v", err)
	}

	// assign
	err = assign.Handle(client, commentEvent)
	if err != nil {
		glog.Errorf("fail to handle: %v", err)
	}

	// label
	err = label.Handle(client, commentEvent)
	if err != nil {
		glog.Errorf("fail to handle: %v", err)
	}
}
