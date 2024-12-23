package main

import "github.com/allape/goenv"

const (
	HbadDatabaseFilename = "HBAD_DATABASE_FILENAME"
	HbadHttpAddress      = "HBAD_HTTP_ADDRESS"
)

var (
	DatabaseFilename = goenv.Getenv(HbadDatabaseFilename, "data/hbad.db")
	HttpAddress      = goenv.Getenv(HbadHttpAddress, ":8080")
)
