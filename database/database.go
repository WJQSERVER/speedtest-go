package database

import (
	"speedtest/config"
	"speedtest/database/bolt"
	"speedtest/database/memory"
	"speedtest/database/mysql"
	"speedtest/database/none"
	"speedtest/database/postgresql"
	"speedtest/database/schema"

	log "github.com/sirupsen/logrus"
)

var (
	DB DataAccess
)

type DataAccess interface {
	Insert(*schema.TelemetryData) error
	FetchByUUID(string) (*schema.TelemetryData, error)
	FetchLast100() ([]schema.TelemetryData, error)
}

func SetDBInfo(conf *config.Config) {
	switch conf.DatabaseType {
	case "postgresql":
		DB = postgresql.Open(conf.DatabaseHostname, conf.DatabaseUsername, conf.DatabasePassword, conf.DatabaseName)
	case "mysql":
		DB = mysql.Open(conf.DatabaseHostname, conf.DatabaseUsername, conf.DatabasePassword, conf.DatabaseName)
	case "bolt":
		DB = bolt.Open(conf.DatabaseFile)
	case "memory":
		DB = memory.Open("")
	case "none":
		DB = none.Open("")
	default:
		log.Fatalf("Unsupported database type: %s", conf.DatabaseType)
	}
}
