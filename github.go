package main

import (
	"context"

	"fmt"

	"os"

	"github.com/blang/semver"
	"github.com/google/go-github/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

var (
	githubUser      = os.Getenv("GITHUB_USER")
	githubToken     = os.Getenv("GITHUB_TOKEN")
	githubRepoOwner = os.Getenv("GITHUB_REPO_OWNER") // current organization factoriotools
	githubRepoName  = os.Getenv("GITHUB_REPO_NAME")  // current repo factorio-docker
)

func createPullRequest(version semver.Version, branch string) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: githubToken},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	title := fmt.Sprintf("Updated to version %s", version.String())
	base := "master"
	modify := true
	_, _, err := client.PullRequests.Create(ctx, githubRepoOwner, githubRepoName, &github.NewPullRequest{
		Title:               &title,
		Head:                &branch,
		Base:                &base,
		MaintainerCanModify: &modify,
	})
	if err != nil {
		logrus.Panic(err)
	}
}
