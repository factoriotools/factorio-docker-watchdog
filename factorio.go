package main

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"net/http"
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
