package handlers

import (
	"github.com/golang/glog"
)

type CircleCIInfo struct {
	CircleJob string `json:"build_parameters[CIRCLE_JOB]"`
	Revision  string `json:"revision"`
}

type CircleCIResp struct {
	BuildURL string `json:"build_url"`
}

const (
	CircleCIGithubURL = "https://circleci.com/api/v1.1/project/github"
	ContentTypeJSON   = "application/json"
)

func (s *Server) SendToCI(info map[string]string) {
	glog.Info("going to send test request to ci")

}

func (s *Server) SendToTravisCI(b []byte) {

}
