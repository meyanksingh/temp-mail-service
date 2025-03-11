package main

import (
	"os"
	"time"

	"github.com/joho/godotenv"
	db "github.com/meyanksingh/smtp-server/db"
	"github.com/meyanksingh/smtp-server/logger"
	"github.com/meyanksingh/smtp-server/server"
	"github.com/meyanksingh/smtp-server/smtp"
)

func init() {
	logger.Initialize()
	logger.Info("Initializing Temporary Email Service")
}

func main() {
	startTime := time.Now()

	logger.Info("Loading environment variables")
	err := godotenv.Load()
	if err != nil {
		logger.Fatal("Error loading environment file: %v", err)
	}
	logger.Info("Environment variables loaded successfully")

	logger.Info("Setting up database connection")
	db.ConnectDB()
	logger.Info("Database connection established")

	logger.Info("Reading configuration from environment")
	host := os.Getenv("HOST")
	smtpPort := os.Getenv("PORT")
	httpPort := os.Getenv("HTTP_PORT")

	if host == "" || smtpPort == "" {
		logger.Fatal("HOST or PORT environment variables are not set")
	}

	if httpPort == "" {
		httpPort = "8000"
		logger.Info("HTTP_PORT not specified, using default: %s", httpPort)
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "DEBUG" {
		logger.SetLevel(logger.LevelDebug)
		logger.Debug("Debug logging enabled")
	}

	logger.Info("Configuration - Host: %s, SMTP Port: %s, HTTP Port: %s", host, smtpPort, httpPort)
	logger.Info("Application initialized in %v", time.Since(startTime))
	logger.Info("Starting servers...")

	go func() {
		logger.Info("Starting HTTP server...")
		server.StartServer(httpPort)
	}()

	logger.Info("Starting SMTP server...")
	smtp.StartSMTPServer(host, smtpPort)
}
