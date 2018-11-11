package handlers

import (
	"fmt"
	"github.com/golang/glog"
	"net/http"

	"google/go-github/github"
)

// Server implements http.Handler. It validates incoming GitHub webhooks and
// then dispatches them to the handlers accordingly.
type Server struct {
	WebHookSecret []byte
	GithubClient *github.Client
}

// ServeHTTP validates an incoming webhook and invoke its handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, s.WebHookSecret)
	if err != nil {
		glog.Errorf("Invalid payload")
		return
	}
	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	if err != nil {
		glog.Errorf("Failed to parse webhook")
		return
	}
	fmt.Fprint(w, "Received a webhook event")

	switch event.(type) {
	case *github.IssueEvent:
		go s.handleIssueEvent()
	case *github.IssueCommentEvent:
		go s.handleIssueCommentEvent()
	}
}

func (s *Server) handleIssueEvent() {
	glog.Infof("Received an Issue Event")
}

func (s *Server) handleIssueCommentEvent() {
	glog.Infof("Received an IssueComment Event")
}
