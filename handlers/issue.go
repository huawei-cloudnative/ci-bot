package handlers

import (
	"encoding/json"
	"fmt"
	"context"
//	"io/ioutil"
//	"net/http"
//	"regexp"
//	"strings"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
)

type GithubIssue github.Issue

func (s *Server) handleIssueEvent(body []byte) {
	glog.Infof("Received an Issue Event")

}

func (s *Server) handleIssueCommentEvent(body []byte, client * github.Client) {
	glog.Infof("Received an IssueComment Event")

	var prc github.IssueCommentEvent
	err := json.Unmarshal(body, &prc)
	if err != nil {
		glog.Errorf("fail to unmarshal: %v", err)
	}
	glog.Infof("prc: %v", prc)
/*	comment := *prc.Comment.Body

	 //https://github.com/islinwb/test/pull/1
	prID := strings.SplitAfter(prc.Issue.PullRequestLinks.GetHTMLURL(), "github.com/")[1]
	 //https://github.com/islinwb/test/pull/1.patch
	 //From <commit ID> MON ...
	patchURL := prc.Issue.PullRequestLinks.GetPatchURL()
	resp, err := http.Get(patchURL)
	if err != nil {
		fmt.Println(err)
	}

	resp1, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	patchDetail := string(resp1)
	reg := regexp.MustCompile(`From [A-Za-z0-9]{40}`)
	commitIDstr := reg.FindString(patchDetail)
	commitID := strings.TrimPrefix(commitIDstr, "From ")

	var info map[string]string
	info["PR_ID"] = prID
	info["Commit_ID"] = commitID

	if labelReg.MatchString(comment) {
		labelSlice := strings.Split(comment, " ")
		if len(labelSlice) > 0 {
		}
	}
	
	if retestReg.MatchString(comment) {
		// "/retest"
		s.SendToCI(info)
	} else if testReg.MatchString(comment) {
		// TODO: trigger particular job(s)
		s.SendToCircleCI(body)
	}*/ 
	
	ctx := context.Background()

	list,_,err := client.Repositories.ListCollaborators(ctx,"swx457056","test-ci-bot",nil)
	if err != nil{
		glog.Fatal("Cannot List the Collaborators",err)
	}
	fmt.Println("list",list)

	assign,_,err := client.Repositories.IsCollaborator(ctx, "swx457056", "test-ci-bot", "sids-b")
	fmt.Println("assign",assign)
	if err != nil {
		glog.Fatal("Not the collaborator",err)

	}

//	var assignees github.IssueRequest
//	get := assignees.GetAssignees()
	get:=make([]string,0)
	get = append(get,"sids-b")
	fmt.Println("***********get***************",get)


	if assign {
		 fmt.Println("Add Assignee")

			issue,_,err := client.Issues.AddAssignees(ctx,"swx457056", "test-ci-bot",1,get)
			fmt.Println("err",err)
			fmt.Println("issue",issue)
		


	}


}
