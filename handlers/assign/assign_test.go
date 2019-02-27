package assign

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-github/github"
)

//github client
var gitclient *github.Client

//test structure
type IssueCommentTests struct {
	name    string
	client  *github.Client
	event   github.IssueCommentEvent
	wantErr error
}

// RewriteTransport rewrites https requests to http to avoid Aunthentication and cert issues
// during testing.
type RewriteTransport struct {
	//Transport http.RoundTripper
	Response *http.Response
	Err      error
}

//RoundTrip rewrites the request scheme to http and calls through to the
// composed RoundTripper
func (t *RewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.Response, t.Err
}

//TestHandleAssign function tests to AddAssignee
func TestHandleAssign(t *testing.T) {
	transport := &RewriteTransport{Response: &http.Response{}}
	//http client
	cli := &http.Client{Transport: transport}
	client := github.NewClient(cli)
	gitclient = client
	//Variables which are pointer to IssueCommentEvent
	var action string
	action = "created"
	var id int64
	id = 10
	var number int
	number = 56
	var payload string
	payload = "/assign @abcd"
	var owner string
	owner = "abcd"
	var name string
	name = "test-ci-bot"
	//Values of the RewriteTransport Response
	transport.Response.StatusCode = 201
	transport.Response.Status = "201 Created"
	transport.Response.Header.Get("Content-Type")
	transport.Response.Body = ioutil.NopCloser(strings.NewReader(""))
	//Error is nil, for true scenarios
	transport.Err = nil
	var tests = IssueCommentTests{
		name:   "Test Add Assignee",
		client: gitclient,
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
		if err := Handle(gitclient, tests.event); err != tests.wantErr {
			t.Errorf("HandleAssignee() error = %v, wantErr %v", err, tests.wantErr)
		}
	})
}

//TestHandleUnAssign function tests to Handle Unassign (Remove assignees)
func TestHandleUnAssign(t *testing.T) {
	transport := &RewriteTransport{Response: &http.Response{}}
	cli := &http.Client{Transport: transport}
	client := github.NewClient(cli)
	gitclient = client
	//Values for the RewriteTransport Response
	transport.Response.StatusCode = 201
	transport.Response.Status = "201 Created"
	transport.Response.Header.Get("Content-Type")
	transport.Response.Body = ioutil.NopCloser(strings.NewReader(""))
	//Variables which are pointer to IssueCommentEvent
	var action string
	action = "created"
	var id int64 = 10
	var number int = 56
	var payload string
	payload = "/unassign @xyzs"
	var owner string
	owner = "owner"
	var name string
	name = "test-ci-bot"
	tests := IssueCommentTests{
		name:   "Test Remove Assignee",
		client: gitclient,
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
		if err := Handle(gitclient, tests.event); err != tests.wantErr {
			t.Errorf("HandleUnAssign() error = %v, wantErr %v", err, tests.wantErr)
		}
	})
}
