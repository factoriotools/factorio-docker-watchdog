package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"

	"github.com/sirupsen/logrus"

	"github.com/blang/semver"
)

var (
	reDockerfileVersion = regexp.MustCompile(`(VERSION=)(\d+\.\d+\.\d+)`)
	reDockerfileSHA1    = regexp.MustCompile(`(SHA1=)([a-z0-9]+)`)
)

func gitSetupCredentials() error {
	cmd := exec.Command("git", "config", "--global", "user.email", os.Getenv("GIT_EMAIL"))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(output))
	}

	cmd = exec.Command("git", "config", "--global", "user.name", os.Getenv("GIT_NAME"))
	output, err = cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(output))
	}

	return nil
}

func gitCloneRepo(path string) error {
	logrus.Debugln("git clone", fmt.Sprintf("https://github.com/%s/%s.git", githubRepoOwner, githubRepoName))
	cmd := exec.Command("git", "clone", fmt.Sprintf("https://%s:%s@github.com/%s/%s.git", githubUser, githubToken, githubRepoOwner, githubRepoName), path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(output))
	}
	return nil
}

func gitCheckoutBranch(path string, branch string) error {
	cmd := exec.Command("git", "checkout", "-b", branch)
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(output))
	}
	return nil
}

func gitCreateCommit(path string, commitMessage string) error {
	logrus.Debugln("git add")
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(output))
	}

	logrus.Debugln("git commit", commitMessage)
	cmd = exec.Command("git", "commit", "-am", commitMessage)
	cmd.Dir = path
	output, err = cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(output))
	}
	return nil
}

func gitPush(path string, branch string) error {
	args := []string{"push"}
	if branch != "" {
		args = append(args, "--set-upstream", "origin", branch)
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(output))
	}
	return nil
}

func gitPushBranch(path string, branch string) error {
	cmd := exec.Command("git", "push", "--set-upstream", "origin", branch)
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(output))
	}
	return nil
}

func gitTagAndPush(path string, tagName string) error {
	cmd := exec.Command("git", "tag", tagName)
	cmd.Dir = path
	output, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(output))
	}

	cmd = exec.Command("git", "push", "origin", tagName)
	cmd.Dir = path
	output, err = cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(output))
	}
	return nil

}

func editDockerfile(path string, version semver.Version, checksum string) error {
	filename := fmt.Sprintf("%s/%d.%d/Dockerfile", path, version.Major, version.Minor)

	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	file = reDockerfileVersion.ReplaceAll(file, []byte("${1}"+version.String()))
	file = reDockerfileSHA1.ReplaceAll(file, []byte("${1}"+checksum))

	err = ioutil.WriteFile(filename, file, 0666)
	if err != nil {
		return err
	}

	return nil
}

func editReadme(path string, version semver.Version) error {
	file, err := ioutil.ReadFile(path + "/README.md")
	if err != nil {
		return err
	}

	reReadmeVersion, err := regexp.Compile(fmt.Sprintf("(`)%d\\.%d\\.\\d+(`)", version.Major, version.Minor))
	if err != nil {
		return err
	}

	file = reReadmeVersion.ReplaceAll(file, []byte("${1}"+version.String()+"${2}"))

	err = ioutil.WriteFile(path+"/README.md", file, 0666)
	if err != nil {
		return err
	}

	return nil
}

func editBuildinfo(path string, buildinfo BuildInfo) error {
	f, err := os.OpenFile(path+"/buildinfo.json", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return encoder.Encode(buildinfo.Versions)
}
