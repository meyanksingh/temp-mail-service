package server

import (
	"math/rand"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	db "github.com/meyanksingh/smtp-server/db"
	"github.com/meyanksingh/smtp-server/logger"
	"github.com/meyanksingh/smtp-server/models"
)

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	logger.Debug("Generated random string of length %d", length)
	return string(b)
}

func getRandomEmail(email string) []models.Message {
	logger.Info("Fetching emails for: %s", email)
	var messages []models.Message
	result := db.DB.Where(`"to" = ?`, email).Find(&messages)
	if result.Error != nil {
		logger.Error("Error fetching emails: %v", result.Error)
		return []models.Message{}
	}

	logger.Info("Found %d messages for %s", len(messages), email)
	return messages
}

func getEmailHandler(c *gin.Context) {
	email := c.Param("email")
	logger.Info("Handling request for email: %s", email)
	messages := getRandomEmail(string(email))
	c.JSON(200, gin.H{"messages": messages})
	logger.Info("Returned %d messages for %s", len(messages), email)
}

func HandleTempMail(c *gin.Context) {
	logger.Info("Handling request for temporary email generation")
	host := os.Getenv("HOST")
	if host == "" {
		host = "meyank.me"
		logger.Warn("HOST environment variable not set, using default: meyank.me")
	}
	randomEmail := generateRandomString(10) + "@" + host
	logger.Info("Generated temporary email: %s", randomEmail)
	c.JSON(200, gin.H{"email": randomEmail})
}

func StartServer(port string) {
	logger.Info("Initializing HTTP server on port %s", port)

	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	logger.Info("CORS middleware configured to allow all origins")

	router.Use(gin.Recovery())

	router.Use(func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		latency := time.Since(startTime)
		status := c.Writer.Status()
		logger.Info("[HTTP] %s | %d | %s | %s", c.Request.Method, status, c.Request.URL.Path, latency)
	})

	router.GET("/tempmail", HandleTempMail)
	router.GET("/tempmail/:email", getEmailHandler)
	logger.Info("HTTP server routes configured, starting on port %s", port)

	if err := router.Run("0.0.0.0:" + port); err != nil {
		logger.Fatal("Failed to start HTTP server: %v", err)
	}
}
