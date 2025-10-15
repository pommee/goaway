package main

import (
	"embed"
	"goaway/backend/logging"
	"goaway/cmd"
)

var (
	version, commit, date string

	//go:embed client/dist/*
	content embed.FS

	log = logging.GetLogger()
)

func main() {
	if err := cmd.Execute(version, commit, date, content); err != nil {
		log.Fatal("Command execution failed: %s", err)
	}
}
