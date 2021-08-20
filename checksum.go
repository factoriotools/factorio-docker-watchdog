package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/blang/semver"

	"github.com/sirupsen/logrus"
)

var checksumsFile = path.Join("/usr/watchdog/factorio-checksums.json")

type checksums struct {
	sha1   map[string]string
	loaded bool
}

func (c *checksums) loadChecksums() {
	if c.loaded {
		return
	}

	c.sha1 = map[string]string{}

	f, err := os.Open(checksumsFile)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		c.loaded = true
		return
	}

	err = json.NewDecoder(f).Decode(&c.sha1)
	if err != nil {
		logrus.Errorln(err)
	}

	c.loaded = true
}

func (c *checksums) saveChecksums() error {
	f, err := os.OpenFile(checksumsFile, os.O_WRONLY, os.ModeAppend)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		f, err = os.Create(checksumsFile)
	}
	if err != nil {
		return err
	}
	defer f.Close()

	err = json.NewEncoder(f).Encode(c.sha1)
	if err != nil {
		return err
	}

	return nil
}

func (c *checksums) getChecksum(version semver.Version) (string, error) {
	c.loadChecksums()
	checksum, ok := c.sha1[version.String()]
	if ok {
		return checksum, nil
	}

	checksum, err := factorioGetChecksum(fmt.Sprintf("https://www.factorio.com/get-download/%s/headless/linux64", version))
	if err != nil {
		return "", err
	}
	if checksum == "" {
		return "", nil
	}

	c.sha1[version.String()] = checksum

	err = c.saveChecksums()
	if err != nil {
		return "", err
	}

	return checksum, nil
}
