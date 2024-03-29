package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/blang/semver"

	"github.com/sirupsen/logrus"
)

const checksumsFile = "/usr/watchdog/factorio-checksums.json"

type checksums struct {
	sha256 map[string]string
	loaded bool
}

func (c *checksums) loadChecksums() {
	if c.loaded {
		return
	}

	c.sha256 = map[string]string{}

	f, err := os.Open(checksumsFile)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		c.loaded = true
		return
	}

	err = json.NewDecoder(f).Decode(&c.sha256)
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

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(c.sha256)
	if err != nil {
		return err
	}

	return nil
}

func (c *checksums) getChecksum(version semver.Version) (string, error) {
	c.loadChecksums()
	checksum, ok := c.sha256[version.String()]
	if ok && checksum != "" {
		return checksum, nil
	}

	checksum, err := factorioGetChecksum(fmt.Sprintf("https://www.factorio.com/get-download/%s/headless/linux64", version))
	if err != nil {
		return "", err
	}
	if checksum == "" {
		return "", nil
	}

	c.sha256[version.String()] = checksum

	err = c.saveChecksums()
	if err != nil {
		return "", err
	}

	return checksum, nil
}
