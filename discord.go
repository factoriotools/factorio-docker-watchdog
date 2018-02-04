package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/blang/semver"
	"github.com/sirupsen/logrus"
)

func noticeDiscord(version semver.Version) error {
	webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	if webhookURL == "" {
		return nil
	}

	// Create struct mapping to Discord webhook object
	var data = map[string]interface{}{
		"embeds": []map[string]interface{}{},
	}
	var embed = map[string]interface{}{
		"title":       "Factorio update released",
		"description": version.String(),
	}
	var fields = []map[string]interface{}{}

	// Merge fields and embed into data
	embed["fields"] = fields
	data["embeds"] = []map[string]interface{}{embed}

	asd, err := json.Marshal(data)
	if err != nil {
		logrus.Panicln(err)
	}

	rsp, err := http.Post(
		webhookURL,
		"application/json; charset=utf-8",
		bytes.NewBuffer(asd),
	)
	if err != nil {
		logrus.Panicln(err)
	}

	a, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		logrus.Panicln(err)
	}

	logrus.Print(string(a))

	return nil
}
