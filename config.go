package main

import (
	"os"
)

// PopulateSettings adds some default settings to config under `key`, including:
// - `cwd`: Current working directory
func PopulateSettings(config map[string]interface{}, key string) {
	settings := make(map[string]string)
	if cwd, err := os.Getwd(); err == nil {
		settings["cwd"] = cwd
	} else {
		panic(err)
	}

	config[key] = settings
}

// SiftSettings adds all keys in `config` to each map in the slice of maps
// accessed by config[under] -> []map[string]interface{} if it doesn't already
// exist.
func SiftSettings(config map[string]interface{}, under string) {
	for _, tree := range config[under].([]interface{}) {
		tree := tree.(map[string]interface{})
		for k, v := range config {
			if _, exists := tree[k]; !(exists || k == under) {
				tree[k] = v
			}
		}
	}
}
