package handlers

import (
	"context"
	"encoding/json"

	"github.com/huawei-cloudnative/ci-bot/handlers/assign"
	"github.com/huawei-cloudnative/ci-bot/handlers/label"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
)

type GithubPR github.PullRequestEvent

func (s *Server) handlePullRequestEvent(body []byte, client *github.Client) {
	glog.Infof("Received an PullRequest Event")
	// get basic params
	ctx := context.Background()

	var prEvent github.PullRequestEvent

	// Unmarshal
	err := json.Unmarshal(body, &prEvent)
	if err != nil {
		glog.Errorf("Failed to unmarshal prEvent: %v", err)
	}
	//PR assignees
	err = assign.HandlePRAssign(ctx, prEvent, client)
	if err != nil {
		glog.Fatalf("HandlePRAssign is failed. err: %v", err)
	}
	//PR Reviewers
	err = assign.HandlePRReviewer(ctx, prEvent, client)
	if err != nil {
		glog.Fatalf("HandlePRReviewer is failed. err: %v", err)
	}
	//PR Labels
	err = label.HandlePRLabels(ctx, prEvent, client)
	if err != nil {
		glog.Fatalf("HandlePRLabels is failed. err: %v", err)
	}
}

func (s *Server) handlePullRequestCommentEvent(body []byte) {
	glog.Infof("Received an PullRequestComment Event")

}
