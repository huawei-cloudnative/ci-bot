package handlers

import (
	"github.com/golang/glog"
	"github.com/google/go-github/github"
)

type GithubPR github.PullRequest

func (s *Server) handlePullRequestEvent(body []byte) {
	glog.Infof("Received an PullRequest Event")

}

func (s *Server) handlePullRequestCommentEvent(body []byte) {
	glog.Infof("Received an PullRequestComment Event")

}
