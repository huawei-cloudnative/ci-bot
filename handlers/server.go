package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
)

// Server implements http.Handler. It validates incoming GitHub webhooks and
// then dispatches them to the handlers accordingly.
type Server struct {
	Config       Config
	GithubClient *github.Client
	Context      context.Context
}

type Config struct {
	Owner         string `json:"owner"`
	Repo          string `json:"repo"`
	GitHubToken   string `json:"git_hub_token"`
	WebhookSecret string `json:"webhook_secret"`
	CircleCIToken string `json:"circle_ci_token"`
}

// ServeHTTP validates an incoming webhook and invoke its handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, []byte(s.Config.WebhookSecret))
	if err != nil {
		glog.Errorf("Invalid payload: %v", err)
		return
	}
	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		glog.Errorf("Failed to parse webhook")
		return
	}
	fmt.Fprint(w, "Received a webhook event")

	glog.Infof("body: %v", string(payload))

	var client http.Client
	client.Do(r)

	switch event.(type) {
	case *github.IssueEvent:
		go s.handleIssueEvent(payload)
	case *github.IssueCommentEvent:
		// Comments on PRs belong to IssueCommentEvent
		go s.handleIssueCommentEvent(payload)
	case *github.PullRequest:
		go s.handlePullRequestEvent(payload)
	case *github.PullRequestComment:
		go s.handlePullRequestCommentEvent(payload)

	}
}
