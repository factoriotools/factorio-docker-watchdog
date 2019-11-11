package main

import "encoding/json"

type imageTag struct {
	Layer string
	Name  string
}

func getTags() ([]imageTag, error) {
	var tags []imageTag
	r, err := myClient.Get("https://index.docker.io/v1/repositories/factoriotools/factorio/tags")
	if err != nil {
		return tags, err
	}
	defer r.Body.Close()

	err = json.NewDecoder(r.Body).Decode(&tags)

	return tags, err
}
