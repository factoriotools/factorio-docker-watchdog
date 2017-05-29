package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var (
	reDockerfileVersion = regexp.MustCompile(`(VERSION=)(\d+\.\d+\.\d+)`)
	reDockerfileSHA1    = regexp.MustCompile(`(SHA1=)([a-z0-9]+)`)
)

func gitCloneRepo(path string) error {
	cmd := exec.Command("git", "clone", fmt.Sprintf("https://%s:%s@github.com/Fank/docker_factorio_server.git", os.Getenv("GITHUB_USER"), os.Getenv("GITHUB_TOKEN")), path)
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
	cmd := exec.Command("git", "commit", "-am", "fk_update")
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

func editDockerfile(path string, version string, checksum string) error {
	parts := strings.Split(version, ".")
	filename := fmt.Sprintf("%s/%s.%s/Dockerfile", path, parts[0], parts[1])

	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	file = reDockerfileVersion.ReplaceAll(file, []byte("${1}"+version))
	file = reDockerfileSHA1.ReplaceAll(file, []byte("${1}"+checksum))

	err = ioutil.WriteFile(filename, file, 0666)
	if err != nil {
		return err
	}

	return nil
}

func editReadme(path string, version string) error {
	parts := strings.Split(version, ".")

	file, err := ioutil.ReadFile(path + "/README.md")
	if err != nil {
		return err
	}

	reReadmeVersion, err := regexp.Compile(fmt.Sprintf("(`)%s\\.%s\\.\\d+(`)", parts[0], parts[1]))
	if err != nil {
		return err
	}

	file = reReadmeVersion.ReplaceAll(file, []byte("${1}"+version+"${2}"))

	err = ioutil.WriteFile(path+"/README.md", file, 0666)
	if err != nil {
		return err
	}

	return nil
}
