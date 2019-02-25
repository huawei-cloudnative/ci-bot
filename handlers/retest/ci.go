package retest

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	. "github.com/Huawei-PaaS/ci-bot/handlers/types"

	"github.com/golang/glog"
)

const (
	TravisCIEndPoint = "https://api.travis-ci.org" //travis-ci endpoint
	ContentTypeJSON  = "application/json"          //content-type
	TravisAPIVersion = "Travis-API-Version"        //API version, a Mandatory header field for Travis-CI V3 APIs call
)

type TravisJobType string

//Travis Build Job Names
const (
	Build       TravisJobType = "build"
	Verify      TravisJobType = "verify"
	Unittest    TravisJobType = "unittest"
	Integration TravisJobType = "integration"
	Crossbuild  TravisJobType = "crossbuild"
)
//Travis Job Id's
const(
	JobBuild        = iota + 1
	JobVerify
	JobUnittest
	JobIntegration
	JobCrossbuild
)

//GetJobIdsFromTravisBuild Function to handle Get Jobs from TravisCI
func GetJobIdsFromTravisBuild(BuildRefId string, Token string)(error, []byte) {
	var Reqbody io.Reader

	url := fmt.Sprintf("%s%s/jobs", TravisCIEndPoint, BuildRefId)
	req, err := http.NewRequest(http.MethodGet, url, Reqbody)
	if err != nil {
		glog.Errorf("HTTP Request failed: %v", err)
		return err,[]byte("")
	}
	err, _, Body := SendHttpReqToCI(req, Token)
	return nil,Body
}

//TriggerJob Function to handle Triggering Jobs build in TravisCI
func TriggerJob(JobBuildId string, Token string) error {
	var Reqbody io.Reader
	var statusCode int
	url := fmt.Sprintf("%s%s/restart", TravisCIEndPoint, JobBuildId)
	req, err := http.NewRequest(http.MethodPost, url, Reqbody)
	if err != nil {
		glog.Errorf("HTTP request failed: %v", err)
		return err
	}
	err, statusCode, _ = SendHttpReqToCI(req, Token)
	if statusCode == http.StatusAccepted {
		glog.Infof("Restart Job Build is successfully triggered !!")
	} else {
		glog.Infof("Restart Job Build is failed to trigger, HttpStatus Code: %v" , http.StatusAccepted)
		return err
	}
	return err
}

//SendHttpReqToCI Function to handle send HTTP request to TravisCI
func SendHttpReqToCI(req *http.Request, Token string) (error, int, []byte) {
	client := &http.Client{}

	req.Header.Set("Content-Type", ContentTypeJSON)
	req.Header.Set("Authorization", "token "+Token)
	req.Header.Set(TravisAPIVersion, "3")
	//Http Do req, For getting the PR Request information for particular repo
	resp, err := client.Do(req)
	if err != nil {
		glog.Errorf("HTTP Do request failed: %v", err)
		return nil, resp.StatusCode, []byte("")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("Failed to read resp: %v", err)
	}
	return err, resp.StatusCode, body
}
//StartToTriggerJob Function to handle travis build job trigger
func StartToTriggerJob(TravisJobRespBody TravisJobRespStruct, jobname TravisJobType, token string) error {
	var err error
	//Range over all jobs to match the JonID and Trigger Build
	for _, job := range TravisJobRespBody.Jobs {
		JobId := strings.Split(job.Number, ".")[1]
		BuildJobId, _ := strconv.Atoi(JobId)
		if Build == jobname && BuildJobId == JobBuild || Verify == jobname && BuildJobId == JobVerify ||
			Unittest == jobname && BuildJobId == JobUnittest || Integration == jobname && BuildJobId == JobIntegration ||
			Crossbuild == jobname && BuildJobId == JobCrossbuild {
			err = TriggerJob(job.Href, token)
			break
		}
	}
	if err != nil {
		glog.Errorf("Failed to Trigger Job: %v", err)
	}
	return err
}

//SendToCIForRetestAllJobs Function to handle RETEST all jobs on PR changes
func SendToCIForRetestAllJobs(prID int, token, repoid string) error {
	glog.Info("Going to send test request to travis ci")

	var Reqbody io.Reader
	var TravisRespBody TravisRespStruct

	url := fmt.Sprintf("%s/repo/%s/requests", TravisCIEndPoint, repoid)
	req, err := http.NewRequest(http.MethodGet, url, Reqbody)
	if err != nil {
		glog.Errorf("HTTP request is failed: %v", err)
		return err
	}
	//Mandatory Header sets to be included for Travis-CI API's
	err, _, Body := SendHttpReqToCI(req, token)
	if err != nil {
		glog.Errorf("Failed to send HTTP request: %v", err)
		return err
	}
	err = json.Unmarshal(Body, &TravisRespBody)
	if err != nil {
		glog.Errorf("Failed to unmarshal TravisRespBody: %v", err)
		return err
	}
	//Fetch the Build id from Requests, then send rebuild request to Travis-CI
	for _, PrRequest := range TravisRespBody.Requests {
		for i, _ := range PrRequest.Builds {
			if PrRequest.Builds[i].PullRequestNumber == prID {
				url := fmt.Sprintf("%s%s/restart", TravisCIEndPoint, PrRequest.Builds[i].Href)
				req, err := http.NewRequest(http.MethodPost, url, Reqbody)
				if err != nil {
					glog.Errorf("HTTP request failed: %v", err)
					return err
				}
				err, statusCode, _ := SendHttpReqToCI(req, token)
				if err != nil && statusCode == http.StatusAccepted {
					glog.Info("Restart Build is successfully triggered !!")
				}
			}
		}
	}
	return nil
}

//SendToCIForTestJob Function to handle TEST on particular JOB
func SendToCIForTestJob(prID int, jobname, token, repoid string) error {
	glog.Info("Going to send test request to travis ci")

	var Reqbody io.Reader
	var TravisRespBody TravisRespStruct
	var TravisJobRespBody TravisJobRespStruct

	url := fmt.Sprintf("%s/repo/%s/requests", TravisCIEndPoint, repoid)
	req, err := http.NewRequest(http.MethodGet, url, Reqbody)
	if err != nil {
		glog.Errorf("HTTP request failed: %v", err)
		return err
	}
	err, _, Body := SendHttpReqToCI(req, token)
	//Mandatory Header sets to be included for Travis-CI API's
	err = json.Unmarshal(Body, &TravisRespBody)
	if err != nil {
		glog.Errorf("Failed to unmarshal TravisRespBody: %v", err)
		return err
	}
	//Fetch the Build id from Requests
	for _, PrRequest := range TravisRespBody.Requests {
		for i, _ := range PrRequest.Builds {
			if PrRequest.Builds[i].PullRequestNumber == prID {
				err, RespBody := GetJobIdsFromTravisBuild(PrRequest.Builds[i].Href, token)
				if err != nil {
					glog.Errorf("Failed to get JobID from Travis-CI: %v", err)
					return err
				}
				err = json.Unmarshal(RespBody, &TravisJobRespBody)
				if err != nil {
					glog.Errorf("Failed to unmarshal TravisJobRespBody: %v", err)
					return err
				}
				err = StartToTriggerJob(TravisJobRespBody, TravisJobType(jobname), token)
				if err != nil {
					glog.Errorf("Failed to unmarshal TravisJobType: %v", err)
					return err
				}
			}
		}
	}
	return err
}
