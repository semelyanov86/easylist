package main

import (
	"easylist/internal/data"
	"easylist/internal/jsonlog"
	"easylist/internal/mailer"
	"sync"
)

var version string
var buildTime string

type config struct {
	Port         int
	Env          string
	Registration bool
	Confirmation bool
	Statistics   bool
	Domain       string
	Frontend     string
	Db           *database
	AppName      string
	Smtp         *smtp
	Limiter      *limiter `yaml:"limiter"`
	Cors         struct {
		TrustedOrigins []string `yaml:"trustedOrigins"`
	}
}

type database struct {
	Dsn          string
	Host         string
	Login        string
	Password     string
	Dbname       string
	MaxOpenConns int    `yaml:"maxOpenConns"`
	MaxIdleConns int    `yaml:"maxIdleConns"`
	MaxIdleTime  string `yaml:"maxIdleTime"`
}

type smtp struct {
	Host     string
	Port     int
	Username string
	Password string
	Sender   string
}

type limiter struct {
	Rps     float64
	Burst   int
	Enabled bool
}

type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}
