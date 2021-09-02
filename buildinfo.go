package main

type BuildInfo struct {
	Versions map[string]BuildInfoVersion
}

type BuildInfoVersion struct {
	SHA256 string   `json:"sha256"`
	Tags   []string `json:"tags"`
}
