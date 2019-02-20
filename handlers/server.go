package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/golang/glog"
	"github.com/google/go-github/github"
	"github.com/spf13/pflag"
	"golang.org/x/oauth2"
)

//Github client
var ClientRepo *github.Client
var c = Config{}

// Server implements http.Handler. It validates incoming GitHub webhooks and
// then dispatches them to the handlers accordingly.
type Server struct {
	Config       Config
	GithubClient *github.Client
	Context      context.Context
}

//config structure
type Config struct {
	Repo          string `json:"repo"`
	GitHubToken   string `json:"git_hub_token"`
	WebhookSecret string `json:"webhook_secret"`
}

//webhook server
type WebHookServer struct {
	Address    string
	Port       int64
	ConfigFile string
}

//webhook handler
func NewWebHookServer() *WebHookServer {
	s := WebHookServer{
		Address: "0.0.0.0",
		Port:    3000,
	}
	return &s
}

//function to add flags
func AddFlags(fs *pflag.FlagSet, s *WebHookServer) {
	fs.StringVar(&s.Address, "address", s.Address, "IP address to serve, 0.0.0.0 by default")
	fs.Int64Var(&s.Port, "port", s.Port, "Port to listen on, 3000 by default")
	fs.StringVar(&c.Repo, "repo", c.Repo, "refers to the project repo address")
	fs.StringVar(&c.GitHubToken, "github-token", c.GitHubToken, "contains the githubtoken info")
	fs.StringVar(&c.WebhookSecret, "webhook-secret", c.WebhookSecret, "contains the webhooksecret key")
	fs.Parse(os.Args[1:])
}

// ServeHTTP validates an incoming webhook and invoke its handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, []byte(c.WebhookSecret))
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

	var client http.Client
	client.Do(r)
	switch event.(type) {
	case *github.IssueEvent:
		go s.handleIssueEvent(payload)
	case *github.IssueCommentEvent:
		// Comments on PRs belong to IssueCommentEvent
		go s.handleIssueCommentEvent(payload, ClientRepo)
	case *github.PullRequestEvent:
		go s.handlePullRequestEvent(payload)
	case *github.PullRequestComment:
		go s.handlePullRequestCommentEvent(payload)
	}
}

//function to run
func Run(s *WebHookServer) {
	oauthSecret := c.GitHubToken
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: string(oauthSecret)},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	ClientRepo = client
	webHookHandler := Server{
		Config:       c,
		GithubClient: ClientRepo,
		Context:      ctx,
	}
	//setting handler
	http.HandleFunc("/hook", webHookHandler.ServeHTTP)

	address := s.Address + ":" + strconv.FormatInt(s.Port, 10)
	//starting server
	if err := http.ListenAndServe(address, nil); err != nil {
		log.Println(err)
	}
}
