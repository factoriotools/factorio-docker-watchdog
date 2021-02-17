package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/blang/semver"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
)

var latestVersionProceeded map[string]string

func main() {
	latestVersionProceeded = map[string]string{}
	versionFill, err := semver.Make(os.Getenv("VERSION_FIX"))
	if err != nil {
		logrus.Error(err)
	}
	latestVersionProceeded[fmt.Sprintf("%d.%d", versionFill.Major, versionFill.Minor)] = versionFill.String()
	latestVersionProceeded["stable"] = "0.0.0"

	err = gitSetupCredentials()
	if err != nil {
		logrus.Panic(err)
	}

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
	stableVersion := semver.Version{}
	for _, version := range version.CoreLinuxHeadless64 {
		if version.Stable != "" {
			v, err := semver.Make(version.Stable)

			if err != nil {
				logrus.Error(err)
			}

			stableVersion = v
		}
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
	logrus.Debug("Available versions from Factorio.com ", versions)
	logrus.Debug("Stable version", stableVersion)

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
	logrus.Info("last version of each major", lastVersions)

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
		if version.Major == 0 && version.Minor < 16 {
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

		updateVersion(version, false)
	}

	if latestVersionProceeded["stable"] != stableVersion.String() {
		updateVersion(stableVersion, true)
	}
}

func updateVersion(version semver.Version, stable bool) {
	logrus.Info("Start ", version.String())
	latestVersionProceeded[fmt.Sprintf("%d.%d", version.Major, version.Minor)] = version.String()
	pathRepo := fmt.Sprintf("/tmp/factorio-%s-repo", version)

	defer func() {
		noticeDiscord(version)
	}()

	err := gitCloneRepo(pathRepo)
	if err != nil {
		logrus.Panic(err)
	}
	defer os.RemoveAll(pathRepo)
	logrus.Info("Cloned repo")

	//err = gitCheckoutBranch(pathRepo, "update-"+version.String())
	//if err != nil {
	//	logrus.Panic(err)
	//}
	//logrus.Info("Checkout branch ", version)

	checksum, err := factorioGetChecksum(fmt.Sprintf("https://www.factorio.com/get-download/%s/headless/linux64", version))
	if err != nil {
		logrus.Panic(err)
	}
	logrus.Info("Got checksum ", checksum)

	err = editDockerfile(pathRepo, version, checksum, stable)
	if err != nil {
		logrus.Panic(err)
	}
	logrus.Info("Edited Dockerfile")

	if !stable {
		err = editReadme(pathRepo, version)
		if err != nil {
			logrus.Panic(err)
		}
		logrus.Info("Edited README")
	}

	err = gitCreateCommit(pathRepo, "Update to Factorio "+version.String())
	if err != nil && strings.Contains(err.Error(), "nothing to commit, working tree clean") {
		logrus.Warnln("nothing to commit, working tree clean")
		return
	} else if err != nil {
		logrus.Panic(err)
	}
	logrus.Info("Committed")

	err = gitPush(pathRepo)
	if err != nil {
		logrus.Panic(err)
	}
	logrus.Info("Pushed")

	err = gitTagAndPush(pathRepo, version.String())
	if err != nil {
		logrus.Panic(err)
	}
	logrus.Info("Tagged")

	//createPullRequest(version, "update-"+version.String())
}
