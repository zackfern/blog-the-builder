package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

	"github.com/google/go-github/github"
	"github.com/gregdel/pushover"
	"golang.org/x/oauth2"
)

// RequiredVars is an internal helper outlining the environment variable expected to be defined for the program to run
var RequiredVars = []string{"GITHUB_ACCESS_TOKEN", "REPO_OWNER", "REPO_NAME"}

// RepoOwner defines the GitHub owner of the repository we'll be downloading.
var RepoOwner string

// RepoName defines the repository we'll be downloading.
var RepoName string

// PushoverClient is a reusable client
var PushoverClient *pushover.Pushover

// PushoverUser is a reusable object used to send notifications
var PushoverUser *pushover.Recipient

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

func setupPushoverClient() {
	pushoverToken := os.Getenv("PUSHOVER_TOKEN")
	if pushoverToken == "" {
		fmt.Println("PUSHOVER_TOKEN is undefined, not setting up Pushover...")
		return
	}
	pushoverUser := os.Getenv("PUSHOVER_USER")
	if pushoverUser == "" {
		fmt.Println("PUSHOVER_USER is undefined, not setting up Pushover...")
		return
	}

	PushoverClient = pushover.New(pushoverToken)
	PushoverUser = pushover.NewRecipient(pushoverUser)
}

func pushoverNotification(message string) {
	if PushoverClient != nil && PushoverUser != nil {
		msg := pushover.NewMessage(message)
		PushoverClient.SendMessage(msg, PushoverUser)
	}
}

func runHugo(buildPath string) error {
	cmd := exec.Command("hugo-0.37")
	cmd.Dir = fmt.Sprintf("./%s", buildPath)
	fmt.Println("cmd.Dir is", cmd.Dir)
	out, err := cmd.Output()
	handleError(err)

	fmt.Println(string(out))
	return err
}

func setupGithubClient() (context.Context, *github.Client) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_ACCESS_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	return ctx, client
}

func handler(w http.ResponseWriter, r *http.Request) {
	ghCtx, ghClient := setupGithubClient()
	url, _, _ := ghClient.Repositories.GetArchiveLink(ghCtx, RepoOwner, RepoName, github.Tarball, &github.RepositoryContentGetOptions{Ref: "master"})

	archivePath := downloadArchive(url)
	extractedPath := extractArchive(archivePath)
	err := runHugo(extractedPath)

	var msg string
	if err != nil {
		msg = "Failed to build site. :("
	} else {
		msg = "Successfully built site!"
	}

	pushoverNotification(msg)
	fmt.Fprintf(w, msg)
}

func main() {
	for _, ev := range RequiredVars {
		if os.Getenv(ev) == "" {
			panic(fmt.Sprintf("Missing required environment variable %s", ev))
		}
	}
	RepoOwner = os.Getenv("REPO_OWNER")
	RepoName = os.Getenv("REPO_NAME")

	setupPushoverClient()

	http.HandleFunc("/build", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
	fmt.Println("Listening on port 8080!")
}
