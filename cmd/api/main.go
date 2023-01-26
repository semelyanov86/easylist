package main

import (
	"context"
	"database/sql"
	"easylist/internal/data"
	"easylist/internal/jsonlog"
	"easylist/internal/mailer"
	"expvar"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var version string
var buildTime string

type config struct {
	port         int
	env          string
	registration bool
	confirmation bool
	domain       string
	db           struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	cors struct {
		trustedOrigins []string
	}
}

type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	wg     sync.WaitGroup
}

func main() {
	var cfg config

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	flag.IntVar(&cfg.port, "port", 4000, "API server port")

	flag.StringVar(&cfg.db.dsn, "dsn", os.Getenv("EASYLIST_DB_DSN"), "MySQL data source name")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "MySql max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "Mysql maximum idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "MySql max connection idle time")

	flag.StringVar(&cfg.env, "env", "development", "Environment (development,staging or production)")
	flag.BoolVar(&cfg.registration, "registration", true, "Is registration enabled")
	flag.BoolVar(&cfg.confirmation, "confirmation", true, "Is email confirmation enabled")

	flag.StringVar(&cfg.smtp.host, "smtp-host", "192.168.10.10", "SMTP Host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 1025, "SMTP Port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "", "SMTP Username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "", "SMTP Password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "EasyList Admin <admin@sergeyem.ru>", "SMTP Sender")

	flag.StringVar(&cfg.domain, "domain", "http://easylist.sergeyem.ru", "Domain name of server")

	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	flag.Func("cors-trusted-origins", "Trusted cors origins (space separated)", func(val string) error {
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})

	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		fmt.Printf("Build time:\t%s\n", buildTime)
		os.Exit(0)
	}

	data.DomainName = cfg.domain
	db, err := openDB(cfg)
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.PrintError(err, nil)
		}
	}(db)

	expvar.NewString("version").Set(version)
	expvar.Publish("goroutines", expvar.Func(func() any {
		return runtime.NumGoroutine()
	}))
	expvar.Publish("database", expvar.Func(func() any {
		return db.Stats()
	}))
	expvar.Publish("timestamp", expvar.Func(func() any {
		return time.Now().Unix()
	}))

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	logger.PrintInfo(fmt.Sprintf("starting %s server on %s", cfg.env, cfg.port), nil)
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.db.dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err = db.PingContext(ctx); err != nil {
		return nil, err
	}
	return db, nil
}
