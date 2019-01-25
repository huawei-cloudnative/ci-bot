package handlers

import (
	"context"
	"bufio"
	"os"
	"strings"
	"syscall"
	"fmt"
//	"golang.org/x/oauth2"
	"github.com/spf13/pflag"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"encoding/json"
	"github.com/golang/glog"
	"github.com/google/go-github/github"
	"golang.org/x/crypto/ssh/terminal"
)

// Server implements http.Handler. It validates incoming GitHub webhooks and
// then dispatches them to the handlers accordingly.
type Server struct {
	Config       Config
	GithubClient *github.Client
	Context      context.Context
}

type Config struct {
	Owner         string `json:"owner"`
	Repo          string `json:"repo"`
	GitHubToken   string `json:"git_hub_token"`
	WebhookSecret string `json:"webhook_secret"`
	CircleCIToken string `json:"circle_ci_token"`
}

type WebHookServer struct {
	Address    string
	Port       int64
	ConfigFile string
}

func NewWebHookServer() *WebHookServer {
	s := WebHookServer{
		Address:    "0.0.0.0",
		Port:       3000,
		//ConfigFile: "/etc/github-robot/config.json",
		ConfigFile: "/root/bot/src/ci-bot/config.json",
	}
	return &s
}

func  AddFlags(fs *pflag.FlagSet,s *WebHookServer) {
	fs.StringVar(&s.Address, "address", s.Address, "IP address to serve, 0.0.0.0 by default")
	fs.Int64Var(&s.Port, "port", s.Port, "Port to listen on, 3000 by default")
	fs.StringVar(&s.ConfigFile, "config-file", s.ConfigFile, "Config file.")
}

// ServeHTTP validates an incoming webhook and invoke its handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	payload, err := github.ValidatePayload(r, []byte(s.Config.WebhookSecret))
	if err != nil {
		glog.Errorf("Invalid payload: %v", err)
		return
	}
	event, err := github.ParseWebHook(github.WebHookType(r), payload)
	fmt.Println("************ event payload **************",event)
	if err != nil {
		glog.Errorf("Failed to parse webhook")
		fmt.Println()
		fmt.Println("******************Inside error condition********************")
		fmt.Println()
		return
	}
	fmt.Fprint(w, "Received a webhook event")

	//glog.Infof("body: %v", string(payload))

	var client http.Client
	client.Do(r)
	switch event.(type) {
	case *github.IssueEvent:
		fmt.Println(" $$$$$$$$$$ Switch IssueEvent $$$$$$$$$$$$$$$")
		go s.handleIssueEvent(payload)
	case *github.IssueCommentEvent:
		// Comments on PRs belong to IssueCommentEvent
		fmt.Println(" $$$$$$$$$$ Switch IssueCommentEvent $$$$$$$$$$$$$$$")
		go s.handleIssueCommentEvent(payload,ClientRepo)
	case *github.PullRequestEvent:
		fmt.Println(" $$$$$$$$$$ Switch Pull Request $$$$$$$$$$$$$$$")
		go s.handlePullRequestEvent(payload,ClientRepo)
	case *github.PullRequestComment:
		fmt.Println(" $$$$$$$$$$ Switch Pull Request Comment $$$$$$$$$$$$$$$")
		go s.handlePullRequestCommentEvent(payload)
	default:
		fmt.Println()
		fmt.Println("**************default payload***********", event)
		fmt.Println()

	}
}

var ClientRepo *github.Client

func  Run(s * WebHookServer) {
	fmt.Println("Inside RUN()")
	configContent, err := ioutil.ReadFile(s.ConfigFile)
	if err != nil {
		glog.Fatal("Could not read config file: %v", err)
	}
	var config Config
	err = json.Unmarshal(configContent, &config)
	if err != nil {
		glog.Fatal("fail to unmarshal: %v", err)
	}
//	oauthSecret := config.GitHubToken
//	fmt.Println("oauthSecret",oauthSecret)
	ctx := context.Background()
	//ts := oauth2.StaticTokenSource(
	//	&oauth2.Token{AccessToken: string(oauthSecret)},
//	)
//	tc := oauth2.NewClient(ctx, ts)
	
	r := bufio.NewReader(os.Stdin)
	fmt.Print("GitHub Username: ")
	username, _ := r.ReadString('\n')

	fmt.Print("GitHub Password: ")
	bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin))
	password := string(bytePassword)

	tp := github.BasicAuthTransport{
		Username: strings.TrimSpace(username),
		Password: strings.TrimSpace(password),
	}

	client := github.NewClient(tp.Client())
	ctx = context.Background()
	user, _, err := client.Users.Get(ctx, "")
	fmt.Println("user",user)
	// Is this a two-factor auth error? If so, prompt for OTP and try again.
	if _, ok := err.(*github.TwoFactorAuthError); ok {
		fmt.Print("\nGitHub OTP: ")
		otp, _ := r.ReadString('\n')
		tp.OTP = strings.TrimSpace(otp)
		user, _, err = client.Users.Get(ctx, "")
	}

	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return
	}

	ClientRepo = client
	fmt.Println("Inside RUN() ", *(ClientRepo.Repositories))
	// return 200 on / for health checks.
	//http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {fmt.Print("hello")})


	webHookHandler := Server{
		Config:       config,
		GithubClient: client,
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

