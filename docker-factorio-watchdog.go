package main

import (
	"fmt"
	"os"

	"github.com/blang/semver"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
)

var latestVersionProceeded map[string]string

func main() {
	latestVersionProceeded = map[string]string{}
	checkVersion()
	c := cron.New()
	c.AddFunc("@every 5m", func() { checkVersion() })
	c.Run()
}

func checkVersion() {
	version, err := GetAvailableVersions()
	if err != nil {
		logrus.Panic(err)
	}

	// Get all available versions
	versions := semver.Versions{}
	for _, version := range version.CoreLinuxHeadless64 {
		if version.To == "" {
			continue
		}
		v, err := semver.Make(version.To)
		if err != nil {
			logrus.Error(err)
			continue
		}

		versions = append(versions, v)
	}
	semver.Sort(versions)

	// Filter only latest version based on minor
	var lastVersion semver.Version
	firstRun := true
	lastVersions := semver.Versions{}
	for i, version := range versions {
		if firstRun {
			lastVersion = version
			firstRun = false
		}

		if lastVersion.Major != version.Major || lastVersion.Minor != version.Minor {
			lastVersions = append(lastVersions, lastVersion)
		} else if i+1 == len(versions) {
			lastVersions = append(lastVersions, version)
		}

		lastVersion = version
	}

	// Remove all tags which are published on docker
	tags, err := getTags()
	for _, tag := range tags {
		tagVersion, err := semver.Make(tag.Name)
		if err != nil {
			continue
		}
		for i, version := range lastVersions {
			if version.String() == tag.Name {
				lastVersions = append(lastVersions[:i], lastVersions[i+1:]...)
			}
		}

		key := fmt.Sprintf("%d.%d", tagVersion.Major, tagVersion.Minor)
		if val, ok := latestVersionProceeded[key]; ok {
			if val == tagVersion.String() {
				logrus.Info("Delete ", tagVersion.String(), " from latest proceeded")
				delete(latestVersionProceeded, key)
			}
		}
	}

	for _, version := range lastVersions {
		if version.Major == 0 && version.Minor < 13 {
			continue
		}
		key := fmt.Sprintf("%d.%d", version.Major, version.Minor)

		logrus.Info(latestVersionProceeded)
		if val, ok := latestVersionProceeded[key]; ok {
			logrus.Info("OK ", val)
			if val == version.String() {
				logrus.Info("Version exists in cache SKIP")
				continue
			}
		}

		updateVersion(version)
	}
}

func updateVersion(version semver.Version) {
	logrus.Info("Start ", version.String())
	latestVersionProceeded[fmt.Sprintf("%d.%d", version.Major, version.Minor)] = version.String()
	pathRepo := fmt.Sprintf("/tmp/factorio-%s-repo", version)

	err := gitCloneRepo(pathRepo)
	if err != nil {
		logrus.Panic(err)
	}
	defer os.RemoveAll(pathRepo)
	logrus.Info("Cloned repo")

	err = gitCheckoutBranch(pathRepo, "update-"+version.String())
	if err != nil {
		logrus.Panic(err)
	}
	logrus.Info("Checkout branch ", version)

	checksum, err := factorioGetChecksum(fmt.Sprintf("https://www.factorio.com/get-download/%s/headless/linux64", version))
	if err != nil {
		logrus.Panic(err)
	}
	logrus.Info("Got checksum ", checksum)

	err = editDockerfile(pathRepo, version, checksum)
	if err != nil {
		logrus.Panic(err)
	}
	logrus.Info("Edited Dockerfile")
	err = editReadme(pathRepo, version)
	if err != nil {
		logrus.Panic(err)
	}
	logrus.Info("Edited README")

	err = gitCreateCommit(pathRepo, "update to "+version.String())
	if err != nil {
		logrus.Panic(err)
	}
	logrus.Info("Committed")

	err = gitPushBranch(pathRepo, "update-"+version.String())
	if err != nil {
		logrus.Panic(err)
	}
	logrus.Info("Pushed")

	createPullRequest(version, "update-"+version.String())
}
