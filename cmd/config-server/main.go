package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/n1xreyes/multi-cloud-k8s-platform/pkg/db/postgres" // (+) Import postgres package
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Config holds the configuration for the Config Service
type Config struct {
	Port    string
	DBHost  string
	DBPort  int
	DBUser  string
	DBPass  string
	DBName  string
	SSLMode string
	Timeout time.Duration
}

// loadConfig loads configuration from environment variables
func loadConfig() Config {
	dbPort, _ := strconv.Atoi(os.Getenv("DB_PORT"))
	if dbPort == 0 {
		dbPort = 5432 // Default PostgreSQL port
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082" // Default port for config service
	}

	return Config{
		Port:    port,
		DBHost:  os.Getenv("DB_HOST"),
		DBPort:  dbPort,
		DBUser:  os.Getenv("DB_USER"),
		DBPass:  os.Getenv("DB_PASSWORD"),
		DBName:  os.Getenv("DB_NAME"),
		SSLMode: os.Getenv("DB_SSLMODE"),
		Timeout: 30 * time.Second,
	}
}

// setupMetrics initializes Prometheus metrics
func setupMetrics() (*prometheus.Registry, gin.HandlerFunc) {
	registry := prometheus.NewRegistry()

	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "config_requests_total",
			Help: "Total number of requests processed by the Config Service",
		},
		[]string{"method", "pass", "status"},
	)

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "config_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	registry.MustRegister(requestCounter, requestDuration)

	metricsMiddleware := func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath() // Use FullPath to group dynamic routes

		c.Next()

		duration := time.Since(start).Seconds()
		status := fmt.Sprintf("%d", c.Writer.Status())

		requestCounter.WithLabelValues(c.Request.Method, path, status).Inc()
		requestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}

	return registry, metricsMiddleware
}

// ApplicationConfigCreateRequest represents the data needed to create/update a config
type ApplicationConfigCreateRequest struct {
	Name       string          `json:"name" binding:"required"`
	Namespace  string          `json:"namespace" binding:"required"`
	ConfigData json.RawMessage `json:"configData" binding:"required"` // keep as RawMessage
}

// Handlers struct to hold dependencies like DB client and logger
type Handlers struct {
	dbClient *postgres.Client
	logger   *zap.Logger
}

// createApplicationConfig handles POST /configs
func (h *Handlers) createApplicationConfig(c *gin.Context) {
	var req ApplicationConfigCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Failed to bind JSON for create config", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body" + err.Error()})
		return
	}

	// Extract UserID from header (assuming gateway forwards it)
	userIDStr := c.GetHeader("X-User-ID")
	userID, err := strconv.Atoi(userIDStr) // Convert to int
	if err != nil {
		// Handle case where header is missing or invalid - decide if it's required
		// For now, let's assume it's optional or handle based on auth policy
		h.logger.Warn("Invalid or missing X-USER-ID header", zap.String("value", userIDStr))
		userID = 0 // Or return an error, e.g., http.StatusUnauthorized
	}

	appConfig := &postgres.ApplicationConfig{
		Name:       req.Name,
		Namespace:  req.Namespace,
		UserID:     userID,                 // Use the extracted userID
		ConfigData: string(req.ConfigData), // Store JSON as string
	}

	if err := h.dbClient.CreateApplicationConfig(c.Request.Context(), appConfig); err != nil {
		h.logger.Warn("Failed to create application config", zap.Error(err))
		// Handle potential unique constraint violation
		if postgres.IsUniqueConstraintViolation(err) {
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("Configuration with name %s in namespace '%s' already exists for this user", req.Name, req.Namespace)})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create configuration"})
		}
		return
	}

	h.logger.Info("Successfully created application config", zap.String("name", appConfig.Name), zap.String("namespace", appConfig.Namespace))
	c.JSON(http.StatusCreated, appConfig)
}

// listApplicationConfigs handles GET /configs
func (h *Handlers) listApplicationConfigs(c *gin.Context) {
	namespace := c.DefaultQuery("namespace", "default")
	userIDStr := c.GetHeader("X-User-ID") // Filter by user if needed
	userID, _ := strconv.Atoi(userIDStr)  // Ignore error for now, treat 0 as "all users" if needed

	configs, err := h.dbClient.ListApplicationConfigs(c.Request.Context(), namespace, userID) // Pass userID
	if err != nil {
		h.logger.Warn("Failed to list application configs", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list application configs"})
		return
	}

	// Convert ConfigData back to JSON object for response
	type ResponseConfig struct {
		ID         int             `json:"id"`
		Name       string          `json:"name"`
		Namespace  string          `json:"namespace"`
		UserID     int             `json:"user_id"`
		ConfigData json.RawMessage `json:"config_data"`
		CreatedAt  time.Time       `json:"created_at"`
		UpdatedAt  time.Time       `json:"updated_at"`
	}

	responseConfigs := make([]ResponseConfig, len(configs))
	for i, cfg := range configs {
		responseConfigs[i] = ResponseConfig{
			ID:         cfg.ID,
			Name:       cfg.Name,
			Namespace:  cfg.Namespace,
			UserID:     cfg.UserID,
			ConfigData: json.RawMessage(cfg.ConfigData), // Convert stringback to RawMessage
			CreatedAt:  cfg.CreatedAt,
			UpdatedAt:  cfg.UpdatedAt,
		}
	}

	h.logger.Info("Successfully listed application configs", zap.String("namespace", namespace), zap.Int("count", len(responseConfigs)))
	c.JSON(http.StatusOK, gin.H{"items": responseConfigs})
}

// getApplicationConfig handles GET /configs/:name
func (h *Handlers) getApplicationConfig(c *gin.Context) {
	name := c.Param("name")
	namespace := c.DefaultQuery("namespace", "default")
	userIDStr := c.GetHeader("X-User-ID") // Get user ID for potential authz check
	userID, _ := strconv.Atoi(userIDStr)

	config, err := h.dbClient.GetApplicationConfigByNameAndNamespace(c.Request.Context(), name, namespace, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			h.logger.Warn("Application config not found", zap.String("name", name), zap.String("namespace", namespace))
			c.JSON(http.StatusNotFound, gin.H{"error": "Application config not found"})
		} else {
			h.logger.Warn("Failed to get application config", zap.Error(err), zap.String("name", name), zap.String("namespace", namespace))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get application config"})
		}
		return
	}

	// Convert ConfigData back to JSON object for response
	type ResponseConfig struct {
		ID         int             `json:"id"`
		Name       string          `json:"name"`
		Namespace  string          `json:"namespace"`
		UserID     int             `json:"user_id"`
		ConfigData json.RawMessage `json:"config_data"`
		CreatedAt  time.Time       `json:"created_at"`
		UpdatedAt  time.Time       `json:"updated_at"`
	}
	response := ResponseConfig{
		ID:         config.ID,
		Name:       config.Name,
		Namespace:  config.Namespace,
		UserID:     config.UserID,
		ConfigData: json.RawMessage(config.ConfigData),
		CreatedAt:  config.CreatedAt,
		UpdatedAt:  config.UpdatedAt,
	}

	h.logger.Info("Successfully retrieved application config", zap.String("name", name), zap.String("namespace", namespace), zap.Int("count", len(response.ConfigData)))
	c.JSON(http.StatusOK, response)
}

// updateApplicationConfig handles PUT /configs/:name
func (h *Handlers) updateApplicationConfig(c *gin.Context) {
	name := c.Param("name")
	namespace := c.DefaultQuery("namespace", "default")
	userIDStr := c.GetHeader("X-User-ID")
	userID, _ := strconv.Atoi(userIDStr)

	var req ApplicationConfigCreateRequest // Reuse create request struct for update
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Warn("Failed to bind JSON for update config", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body" + err.Error()})
		return
	}

	// Ensure the name/namespace in URL match the body (or ignore body values)
	if req.Name != name || req.Namespace != namespace {
		// c.JSON(http.StatusBadRequest, gin.H{"error": "Name/Namespace in URL and body must match"})
		// return
		// Or prioritize URL params
		req.Name = name
		req.Namespace = namespace
	}

	appConfig := &postgres.ApplicationConfig{
		Name:       req.Name,
		Namespace:  req.Namespace,
		UserID:     userID, // Use the extracted userID for check
		ConfigData: string(req.ConfigData),
	}

	err := h.dbClient.UpdateApplicationConfig(c.Request.Context(), appConfig)
	if err != nil {
		if err == sql.ErrNoRows {
			h.logger.Warn("Attempted to update non-existent config", zap.String("name", name), zap.String("namespace", namespace))
			c.JSON(http.StatusNotFound, gin.H{"error": "Application config not found"})
		} else {
			h.logger.Warn("Failed to update application config", zap.Error(err), zap.String("name", name), zap.String("namespace", namespace))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update application config"})
		}
		return
	}

	// Fetch the updated record to return it
	updatedConfig, err := h.dbClient.GetApplicationConfigByNameAndNamespace(c.Request.Context(), name, namespace, userID)
	if err != nil {
		// This shouldn't happen if update succeeded, but handle defensively
		h.logger.Error("Failed to get application config after update", zap.Error(err), zap.String("name", name), zap.String("namespace", namespace))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve updated application config"})
		return
	}

	// Convert ConfigData back to JSON object for response
	type ResponseConfig struct {
		ID         int             `json:"id"`
		Name       string          `json:"name"`
		Namespace  string          `json:"namespace"`
		UserID     int             `json:"userId"`
		ConfigData json.RawMessage `json:"configData"` // Use RawMessage
		CreatedAt  time.Time       `json:"createdAt"`
		UpdatedAt  time.Time       `json:"updatedAt"`
	}

	response := ResponseConfig{
		ID:         updatedConfig.ID,
		Name:       updatedConfig.Name,
		Namespace:  updatedConfig.Namespace,
		UserID:     updatedConfig.UserID,
		ConfigData: json.RawMessage(updatedConfig.ConfigData),
		CreatedAt:  updatedConfig.CreatedAt,
		UpdatedAt:  updatedConfig.UpdatedAt,
	}

	h.logger.Info("Successfully updated application config", zap.String("name", name), zap.String("namespace", namespace))
	c.JSON(http.StatusOK, response)
}

// deleteApplicationConfig handles DELETE /configs/:name
func (h *Handlers) deleteApplicationConfig(c *gin.Context) {
	name := c.Param("name")
	namespace := c.DefaultQuery("namespace", "default")
	userIDStr := c.GetHeader("X-User-ID")
	userID, _ := strconv.Atoi(userIDStr)

	err := h.dbClient.DeleteApplicationConfig(c.Request.Context(), name, namespace, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			h.logger.Warn("Application config not found", zap.String("name", name), zap.String("namespace", namespace))
			c.JSON(http.StatusNotFound, gin.H{"error": "Application config not found"})
		} else {
			h.logger.Warn("Failed to delete application config", zap.String("name", name), zap.String("namespace", namespace))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete application config"})
		}
		return
	}

	h.logger.Info("Successfully deleted application config", zap.String("name", name), zap.String("namespace", namespace))
	c.JSON(http.StatusOK, gin.H{"message": "Successfully deleted application config"})
}

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("can't initialize zap logger: %v", err)
	}
	defer logger.Sync() // Flushes buffer, if any

	// Load configuration
	cfg := loadConfig()
	logger.Info("Configuration loaded successfully", zap.String("port", cfg.Port), zap.String("db_host", cfg.DBHost))

	// Connect to Postgres
	pgConfig := postgres.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPass,
		DBName:   cfg.DBName,
		SSLMode:  cfg.SSLMode,
	}

	dbClient, err := postgres.NewClient(context.Background(), pgConfig)
	if err != nil {
		logger.Fatal("Failed to connect to postgres", zap.Error(err))
	}
	defer dbClient.Close()
	logger.Info("Successfully connected to postgres")

	// Setup Prometheus registry and middleware
	registry, metricsMiddleware := setupMetrics()

	// Set Gin to release mode for production
	gin.SetMode(gin.ReleaseMode)

	// Create Gin router
	router := gin.New()
	router.Use(gin.Recovery())    // Recover from panics
	router.Use(metricsMiddleware) // Use metrics middleware

	// Simple logging middleware
	router.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		logger.Info("Request handled,",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.String("client_IP", c.ClientIP()))
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		// Add DB ping check
		ctx, cancel := context.WithTimeout(c.Request.Context(), 1*time.Second)
		defer cancel()
		if err := dbClient.PingContext(ctx); err != nil {
			logger.Error("Health check failed: DB ping error", zap.Error(err))
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "reason": "database connection failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})
}
