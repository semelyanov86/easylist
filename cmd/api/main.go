package main

import (
	"context"
	"database/sql"
	"easylist/internal/data"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const version = "1.0.0"

type config struct {
	port         int
	env          string
	registration bool
	confirmation bool
	db           struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

type application struct {
	config config
	logger *log.Logger
	models data.Models
}

func main() {
	var cfg config
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	flag.IntVar(&cfg.port, "port", 4000, "API server port")

	flag.StringVar(&cfg.db.dsn, "dsn", os.Getenv("EASYLIST_DB_DSN"), "MySQL data source name")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "MySql max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "Mysql maximum idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "MySql max connection idle time")

	flag.StringVar(&cfg.env, "env", "development", "Environment (development,staging or production)")
	flag.BoolVar(&cfg.registration, "registration", true, "Is registration enabled")
	flag.BoolVar(&cfg.confirmation, "confirmation", true, "Is email confirmation enabled")
	flag.Parse()

	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			logger.Println(err)
		}
	}(db)

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
	err = srv.ListenAndServe()
	logger.Fatal(err)
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
