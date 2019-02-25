package assign

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

//http client
var client *http.Client

//defaultAuthToken to Authenticate github account
const (
	defaultAuthToken = "18a91453fd541adbb47676a6d5ea80f17b79f9d4"
)

//test structure
type IssueCommentTests struct {
	name    string
	client  *github.Client
	event   github.IssueCommentEvent
	wantErr error
}

//Test function to Handle AddAssignee
func TestHandleAssign(t *testing.T) {
	//Authtoken for authentication
	oauthSecret := defaultAuthToken
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: string(oauthSecret)},
	)
	//Call to the http Client with token
	tc := oauth2.NewClient(ctx, ts)
	//github client
	client := github.NewClient(tc)
	//Variables which are pointer to IssueCommentEvent
	var action string
	action = "created"
	var id int64 = 10
	id = 10
	var number int
	number = 56
	var payload string
	payload = "/assign @abcd"
	var owner string
	owner = "swx457056"
	var name string
	name = "test-ci-bot"

	var tests = IssueCommentTests{
		name:   "Test Assign with correct values",
		client: client,
		event: github.IssueCommentEvent{
			Action: &action,
			Issue: &github.Issue{
				ID:     &id,
				Number: &number,
			},
			Comment: &github.IssueComment{
				ID:   &id,
				Body: &payload,
			},
			Repo: &github.Repository{
				ID: &id,
				Owner: &github.User{
					Login: &owner,
				},
				Name: &name,
			},
		},
		wantErr: nil,
	}
	t.Run(tests.name, func(t *testing.T) {
		if err := Handle(tests.client, tests.event); err != tests.wantErr {
			t.Errorf("Handle() error = %v, wantErr %v", err, tests.wantErr)
		}
	})

}

//Test function to Handle Unassign (Remove assignees)
func TestHandleUnAssign(t *testing.T) {
	oauthSecret := defaultAuthToken
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: string(oauthSecret)},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	//Variables which are pointer to IssueCommentEvent
	var action string
	action = "created"
	var id int64 = 10
	var number int = 56
	var payload string
	payload = "/unassign @abcd"
	var owner string
	owner = "swx457056"
	var name string
	name = "test-ci-bot"
	tests := IssueCommentTests{
		name:   "Test UnAssign with correct values",
		client: client,
		event: github.IssueCommentEvent{
			Action: &action,
			Issue: &github.Issue{
				ID:     &id,
				Number: &number,
			},
			Comment: &github.IssueComment{
				ID:   &id,
				Body: &payload,
			},
			Repo: &github.Repository{
				ID: &id,
				Owner: &github.User{
					Login: &owner,
				},
				Name: &name,
			},
		},
		wantErr: nil,
	}
	t.Run(tests.name, func(t *testing.T) {
		if err := Handle(tests.client, tests.event); err != tests.wantErr {
			t.Errorf("HandleUnAssign() error = %v, wantErr %v", err, tests.wantErr)
		}
	})
}
