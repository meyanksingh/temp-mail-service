package server

import (
	"context"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/meyanksingh/smtp-server/logger"
	"github.com/meyanksingh/smtp-server/models"
	"github.com/meyanksingh/smtp-server/redis"
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

	ctx := context.Background()
	messageStrs, err := redis.RedisClient.LRange(ctx, email, 0, -1).Result()
	if err != nil {
		logger.Error("Error fetching emails from Redis: %v", err)
		return []models.Message{}
	}

	if len(messageStrs) == 0 {
		logger.Info("No messages found for: %s", email)
		return []models.Message{}
	}

	var messages []models.Message
	for i, messageStr := range messageStrs {
		message := models.Message{
			ID:        uint(i),
			From:      "",
			To:        "",
			Subject:   "",
			Body:      "",
			CreatedAt: time.Now(),
		}

		lines := strings.Split(messageStr, "\n")
		var bodyStarted bool
		var bodyLines []string

		for _, line := range lines {
			if bodyStarted {
				bodyLines = append(bodyLines, line)
				continue
			}
			if strings.HasPrefix(line, "Body:") {
				bodyStarted = true
				bodyContent := strings.TrimPrefix(line, "Body:")
				bodyLines = append(bodyLines, bodyContent)
				continue
			}
			switch {
			case strings.HasPrefix(line, "From:"):
				message.From = strings.TrimSpace(strings.TrimPrefix(line, "From:"))
			case strings.HasPrefix(line, "To:"):
				message.To = strings.TrimSpace(strings.TrimPrefix(line, "To:"))
			case strings.HasPrefix(line, "Subject:"):
				message.Subject = strings.TrimSpace(strings.TrimPrefix(line, "Subject:"))
			case strings.HasPrefix(line, "Time:"):
				timeStr := strings.TrimSpace(strings.TrimPrefix(line, "Time:"))
				t, _ := time.Parse(time.RFC3339, timeStr)
				message.CreatedAt = t
			}
		}

		if len(bodyLines) > 0 {
			message.Body = strings.Join(bodyLines, "\n")
		}

		messages = append(messages, message)
	}

	logger.Info("Retrieved %d messages from Redis for %s", len(messages), email)
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
	hosts := strings.Split(host, ",")
	randomHost := hosts[rand.Intn(len(hosts))]
	randomEmail := generateRandomString(10) + "@" + randomHost
	logger.Info("Generated temporary email: %s", randomEmail)
	c.JSON(200, gin.H{"email": randomEmail})
}

func StartServer(port string) {
	logger.Info("Initializing HTTP server on port %s", port)

	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	if allowedOrigins == "" {
		allowedOrigins = "*"
	}
	router.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Split(allowedOrigins, ","),
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
