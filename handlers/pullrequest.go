package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/google/go-github/github"
)

type GithubPR github.PullRequestEvent

var client github.Client

func (s *Server) handlePullRequestEvent(body []byte, client *github.Client) {
	ctx := context.Background()
	glog.Infof("***********Received an PullRequest Event *****************")
	var pull github.PullRequestEvent
	err := json.Unmarshal(body, &pull)
	if err != nil {
		glog.Errorf("fail to unmarshal: %v", err)
	}
	glog.Infof("pull: %v", pull)
	fmt.Println(" @@@@@@@@@@@@@@@@ pull request @@@@@@@@@@@@",pull.PullRequest)
	PRList, _, err := client.Repositories.ListCollaborators(ctx, "swx457056", "test-ci-bot", nil)
	fmt.Println("*********** err ***************", err)
	fmt.Println("&&&&&&&&&&&& PRLIst Collaborators", PRList)
	fmt.Println()
	fmt.Println("pull request event", pull)

	contributors, resp, err := client.Repositories.ListContributors(ctx, "swx457056", "test-ci-bot", nil)
	fmt.Println("*******contributors**************", &contributors)
	fmt.Println()
	fmt.Println("resp", resp)
	fmt.Println("err", err)
	fmt.Println()

	var reviewreq github.ReviewersRequest
	reviewreq.Reviewers = []string{"sids-b", "swx457056"}
	reviewreq.TeamReviewers = []string{"sids-b", "swx457056"}
	fmt.Println("######## reviewreq.Reviewers ##############", reviewreq.Reviewers)

	rr, _, _ := client.PullRequests.RequestReviewers(ctx, "swx457056", "test-ci-bot", 39, reviewreq)
	fmt.Println(" $$$$$$$$$$$$$ rr merged $$$$$$$$$$$$$$",rr.Merged)

	fmt.Println(" %%%%%%%%%%% rr %%%%%%%%%%%", rr)
	
	if !*rr.Merged{
		merged,_,_ := client.PullRequests.Merge(ctx,"swx457056","test-ci-bot",39,"TEST",nil)
		fmt.Println("************ Merged ***************",merged)

	}

	testmerge,_,err := client.PullRequests.IsMerged(ctx,"swx457056","test-ci-bot",37)
	fmt.Println(" ########### test merge ###########",testmerge)
}

func (s *Server) handlePullRequestCommentEvent(body []byte) {
	glog.Infof("Received an PullRequestComment Event")
	fmt.Println("**************** handlePullRequestCommentEvent *********")

}
