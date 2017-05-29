package main

import (
	"os"

	"fmt"

	"github.com/Sirupsen/logrus"
)

func main() {
	logrus.Info("Start")
	version := "0.15.16"
	pathRepo := fmt.Sprintf("/tmp/factorio-%s-repo", version)

	err := gitCloneRepo(pathRepo)
	if err != nil {
		logrus.Panic(err)
	}
	defer os.RemoveAll(pathRepo)
	logrus.Info("Cloned repo")

	err = gitCheckoutBranch(pathRepo, "update-"+version)
	if err != nil {
		logrus.Panic(err)
	}
	logrus.Info("Checkout branch", version)

	checksum, err := factorioGetChecksum(fmt.Sprintf("https://www.factorio.com/get-download/%s/headless/linux64", version))
	if err != nil {
		logrus.Panic(err)
	}
	logrus.Info("Got checksum", checksum)

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

	err = gitCreateCommit(pathRepo, "update")
	if err != nil {
		logrus.Panic(err)
	}
	logrus.Info("Commited")

	err = gitPushBranch(pathRepo, "update-"+version)
	if err != nil {
		logrus.Panic(err)
	}
	logrus.Info("Pushed")
}
