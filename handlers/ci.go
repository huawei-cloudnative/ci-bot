package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

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
	glog.Info("going to send test request to circle ci")

	// TODO: the current way to trigger CircleCI is stupid, find a better way if any

	client := &http.Client{}
	// TODO: substitute with specified job name
	circleCIInfo := CircleCIInfo{
		CircleJob: "build",
		Revision:  info["Commit_ID"],
	}
	jsonStr, err := json.Marshal(circleCIInfo)
	if err != nil {
		glog.Errorf("fail to marshal: %v", err)
	}
	url := fmt.Sprintf("%s/%s/%s/pulls/%s", CircleCIGithubURL, s.Config.Owner, s.Config.Repo, info["PR_ID"])
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonStr))
	if err != nil {
		glog.Errorf("%v", err)
	}
	req.Header.Set("Content-Type", ContentTypeJSON)
	req.SetBasicAuth(s.Config.CircleCIToken, "")
	resp, err := client.Do(req)
	var circleCIResp CircleCIResp
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("fail to read resp: %v", err)
	}
	err = json.Unmarshal(body, circleCIResp)
	if err != nil {
		glog.Errorf("fail to unmarshal: %v", err)
	}

	// buildURL is the CircleCI link of the test for PR
	buildURL := circleCIResp.BuildURL
	glog.Infof("the CircleCI test link: %s", buildURL)
}

func (s *Server) SendToCircleCI(b []byte) {

}
