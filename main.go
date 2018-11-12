package main

import (
	"context"
	"github.com/golang/glog"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github-robot/handlers"

	"github.com/spf13/pflag"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type WebHookServer struct {
	Address      string
	Port         int64
	GitHubTokenFile string
	WebHookKeyFile string
}

func NewWebHookServer() *WebHookServer {
	s := WebHookServer{
		Address:      "0.0.0.0",
		Port:         3000,
		GitHubTokenFile: "/etc/github/token",
		WebHookKeyFile: "/etc/github/webhook",
	}
	return &s
}

func (s *WebHookServer) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.Address, "address", s.Address, "IP address to serve, 0.0.0.0 by default")
	fs.Int64Var(&s.Port, "port", s.Port, "Port to listen on, 3000 by default")
	fs.StringVar(&s.GitHubTokenFile, "github-token-file", s.GitHubTokenFile,"GitHub OAuth secret file.")
	fs.StringVar(&s.WebHookKeyFile, "webhook-key", s.WebHookKeyFile,"GitHub webhook key file.")
}

func (s *WebHookServer) Run() {
	oauthSecret, err := ioutil.ReadFile(s.GitHubTokenFile)
	if err != nil {
		glog.Fatal("Could not read oauth secret file.")
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: string(oauthSecret)},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)
	// Return 200 on / for health checks.
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {})

	webhookSecret, err := ioutil.ReadFile(s.WebHookKeyFile)
	if err != nil {
		glog.Fatal("Could not read webhook secret file.")
	}
	webHookHandler := handlers.Server{
		WebHookSecret: webhookSecret,
		GithubClient: client,
	}
	//setting handler
	http.HandleFunc("/hook", webHookHandler.ServeHTTP)

	address := s.Address + ":" + strconv.FormatInt(s.Port, 10)
	//starting server
	if err := http.ListenAndServe(address, nil); err!= nil{
		log.Println(err)
	}
}

func main() {
	s := NewWebHookServer()
	s.AddFlags(pflag.CommandLine)

	s.Run()
}
