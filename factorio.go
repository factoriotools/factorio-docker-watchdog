package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

func factorioGetChecksum(url string) (string, error) {
	var returnSHA1String string

	//Open a new SHA1 hash interface to write to
	hash := sha1.New()

	resp, err := http.Get(url)
	if err != nil {
		return returnSHA1String, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New(resp.Status)
	}

	_, err = io.Copy(hash, resp.Body)
	if err != nil {
		return returnSHA1String, err
	}

	//Get the 20 bytes hash
	hashInBytes := hash.Sum(nil)[:20]

	//Convert the bytes to a string
	returnSHA1String = hex.EncodeToString(hashInBytes)

	return returnSHA1String, nil
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
