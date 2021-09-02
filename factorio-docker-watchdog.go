package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/blang/semver"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
)

func main() {
	err := gitSetupCredentials()
	if err != nil {
		logrus.Panic(err)
	}

	logrus.SetLevel(logrus.DebugLevel)

	checkVersion()
	c := cron.New()
	c.AddFunc("@every 5m", func() { checkVersion() })
	c.Run()
}

func checkVersion() {
	availableVersions, err := GetAvailableVersions()
	if err != nil {
		logrus.Panic(err)
	}

	// Get all available versions
	versions := semver.Versions{}
	var stableVersion *semver.Version
	for _, version := range availableVersions.CoreLinuxHeadless64 {
		if version.Stable != "" {
			v, err := semver.Make(version.Stable)
			if err != nil {
				logrus.Error(err)
				continue
			}

			stableVersion = &v
		}
		if version.To != "" {
			v, err := semver.Make(version.To)
			if err != nil {
				logrus.Error(err)
				continue
			}

			versions = append(versions, v)
		}
	}
	semver.Sort(versions)
	logrus.Debug("Available versions from Factorio.com ", versions)

	// Filter only latest version based on minor
	var lastVersion semver.Version
	highestMajor := make(map[uint64]semver.Version)
	highestMinor := make(map[string]semver.Version)
	firstRun := true
	lastVersions := semver.Versions{}

	for i, version := range versions {
		if firstRun {
			lastVersion = version
			firstRun = false
		}
		if _, ok := highestMajor[version.Major]; !ok {
			highestMajor[version.Major] = version
		}
		highestMinorKey := fmt.Sprintf("%d.%d", version.Major, version.Minor)
		if _, ok := highestMinor[highestMinorKey]; !ok {
			highestMinor[highestMinorKey] = version
		}

		if lastVersion.Major != version.Major || lastVersion.Minor != version.Minor {
			lastVersions = append(lastVersions, lastVersion)
		} else if i+1 == len(versions) {
			lastVersions = append(lastVersions, version)
		}
		lastVersion = version

		if highestMajor[version.Major].LT(version) {
			highestMajor[version.Major] = version
		}
		if highestMinor[highestMinorKey].LT(version) {
			highestMinor[highestMinorKey] = version
		}
	}
	logrus.Info("last version of each major", lastVersions)

	buildinfo := BuildInfo{
		Versions: map[string]BuildInfoVersion{},
	}
	stableFound := false
	checks := checksums{}
	for _, v := range lastVersions {
		minorKey := fmt.Sprintf("%d.%d", v.Major, v.Minor)

		// stable version exists and was not found in previous records
		if !stableFound && stableVersion != nil && v.LT(*stableVersion) {
			checksum, err := checks.getChecksum(*stableVersion)
			if err != nil {
				logrus.Panicln(err)
			}
			version := BuildInfoVersion{
				SHA1: checksum,
				Tags: []string{
					stableVersion.String(),
					"stable",
				},
			}
			buildinfo.Versions[stableVersion.String()] = version
		}

		checksum, err := checks.getChecksum(v)
		if err != nil {
			logrus.Panicln(err)
		}
		version := BuildInfoVersion{
			SHA1: checksum,
			Tags: []string{
				v.String(),
			},
		}
		if highestMajor[v.Major].EQ(v) {
			version.Tags = append(version.Tags, fmt.Sprintf("%d", v.Major))
		}
		if highestMinor[minorKey].EQ(v) {
			version.Tags = append(version.Tags, minorKey)
		}
		if v.EQ(lastVersion) {
			version.Tags = append(version.Tags, "latest")
		}
		if stableVersion != nil && v.EQ(*stableVersion) {
			version.Tags = append(version.Tags, "stable")
			stableFound = true
		}
		buildinfo.Versions[v.String()] = version
	}

	updateVersion(buildinfo)
}

func updateVersion(buildinfo BuildInfo) {
	logrus.Info("Start ", buildinfo)
	pathRepo := "/tmp/factorio-repo"

	//defer func() {
	//	noticeDiscord(version)
	//}()

	err := gitCloneRepo(pathRepo)
	if err != nil {
		logrus.Panic(err)
	}
	defer os.RemoveAll(pathRepo)
	logrus.Info("Cloned repo")

	//err = gitCheckoutBranch(pathRepo, "update-buildinfo")
	//if err != nil {
	//	logrus.Panic(err)
	//}
	//logrus.Info("Checkout branch")

	err = editReadme(pathRepo, buildinfo)
	if err != nil {
		logrus.Panic(err)
	}
	logrus.Info("Edited README")

	err = editBuildinfo(pathRepo, buildinfo)
	if err != nil {
		logrus.Panic(err)
	}
	logrus.Info("Edited Buildinfo")

	err = gitCreateCommit(pathRepo, "Update to Factorio version")
	if err != nil && strings.Contains(err.Error(), "nothing to commit, working tree clean") {
		logrus.Warnln("nothing to commit, working tree clean")
		return
	} else if err != nil {
		logrus.Panic(err)
	}
	logrus.Info("Committed")

	err = gitPush(pathRepo, "")
	if err != nil {
		logrus.Panic(err)
	}
	logrus.Info("Pushed")

	//err = gitTagAndPush(pathRepo, version.String())
	//if err != nil {
	//	logrus.Panic(err)
	//}
	//logrus.Info("Tagged")

	//createPullRequest(version, "update-"+version.String())
}
