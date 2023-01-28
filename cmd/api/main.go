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
	"gopkg.in/yaml.v3"
	"os"
	"runtime"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	var cfg config

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	dirname, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}
	bytesOut, err := os.ReadFile(dirname + "/easylist.yaml")

	if err != nil {
		panic(err)
	}
	if err := yaml.Unmarshal(bytesOut, &cfg); err != nil {
		panic(err)
	}
	cfg.Db.Dsn = cfg.Db.Login + ":" + cfg.Db.Password + "@" + cfg.Db.Host + "/" + cfg.Db.Dbname + "?parseTime=true"
	displayVersion := flag.Bool("version", false, "Display version and exit")

	flag.Parse()

	if *displayVersion {
		fmt.Printf("Version:\t%s\n", version)
		fmt.Printf("Build time:\t%s\n", buildTime)
		os.Exit(0)
	}

	data.DomainName = cfg.Domain
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
		mailer: mailer.New(cfg.Smtp.Host, cfg.Smtp.Port, cfg.Smtp.Username, cfg.Smtp.Password, cfg.Smtp.Sender),
	}

	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}
	logger.PrintInfo(fmt.Sprintf("starting %s server on %d", cfg.Env, cfg.Port), nil)
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("mysql", cfg.Db.Dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.Db.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Db.MaxIdleConns)
	duration, err := time.ParseDuration(cfg.Db.MaxIdleTime)
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
