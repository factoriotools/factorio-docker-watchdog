package main

type BuildInfo struct {
	Versions map[string]BuildInfoVersion
}

type BuildInfoVersion struct {
	SHA1 string   `json:"sha1"`
	Tags []string `json:"tags"`
}
