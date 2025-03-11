package database

import (
	"log"
	"os"
	"time"

	customlogger "github.com/meyanksingh/smtp-server/logger"
	"github.com/meyanksingh/smtp-server/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDB() *gorm.DB {
	customlogger.Info("[DB] Initializing database connection")

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		customlogger.Fatal("[DB] DATABASE_URL environment variable not set")
	}

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			LogLevel:      logger.Info,
			Colorful:      true,
		},
	)
	customlogger.Info("[DB] GORM logger configured")

	var err error

	// connecting to the database
	customlogger.Info("[DB] Attempting to connect to PostgreSQL database")
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		customlogger.Fatal("[DB] Failed to connect to the database: %v", err)
	}
	customlogger.Info("[DB] Database connection established successfully")

	sqlDB, err := DB.DB()
	if err != nil {
		customlogger.Warn("[DB] Warning: Failed to get SQL DB handle: %v", err)
	} else {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)
		customlogger.Info("[DB] Connection pool configured with MaxIdleConns=10, MaxOpenConns=100")
	}

	customlogger.Info("[DB] Running database migrations")
	err = DB.AutoMigrate(&models.Message{})
	if err != nil {
		customlogger.Fatal("[DB] Failed to migrate database models: %v", err)
	}
	customlogger.Info("[DB] Database migrations completed successfully")

	return DB
}
