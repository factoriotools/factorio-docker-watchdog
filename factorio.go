package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

func factorioGetChecksum(url string) (string, error) {
	checksum := ""

	logrus.Debugln("downloading", url, "for checksum check")

	//Open a new SHA256 hash interface to write to
	hash := sha256.New()

	resp, err := http.Get(url)
	if err != nil {
		return checksum, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", nil
	} else if resp.StatusCode != http.StatusOK {
		return "", errors.New(resp.Status)
	}

	_, err = io.Copy(hash, resp.Body)
	if err != nil {
		return checksum, err
	}

	//Get the 20 bytes hash
	hashInBytes := hash.Sum(nil)[:20]

	//Convert the bytes to a string
	checksum = hex.EncodeToString(hashInBytes)

	logrus.Debugln("got checksum", checksum)

	return checksum, nil
}

var myClient = &http.Client{Timeout: 10 * time.Second}

func GetAvailableVersions() (AvailableVersions, error) {
	var availableVersions AvailableVersions
	r, err := myClient.Get("https://www.factorio.com/updater/get-available-versions?apiVersion=2")
	if err != nil {
		return availableVersions, err
	}
	defer r.Body.Close()

	err = json.NewDecoder(r.Body).Decode(&availableVersions)

	return availableVersions, err
}

type AvailableVersions struct {
	CoreLinuxHeadless64 []AvailableVersionsStep `json:"core-linux_headless64"`
}
type AvailableVersionsStep struct {
	From   string
	To     string
	Stable string
}
