package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// ServiceConfig holds the configuration for the REST API service
type ServiceConfig struct {
	Port                 string
	DeploymentServiceURL string
	MonitoringServiceURL string
	ConfigServiceURL     string
	Timeout              time.Duration
}

// ServiceClient represents a client to interact with microservices
type ServiceClient struct {
	deploymentClient *http.Client
	monitoringClient *http.Client
	configClient     *http.Client
	logger           *zap.Logger
	config           ServiceConfig
}

// NewServiceConfig loads configuration from environment variables
func NewServiceConfig() ServiceConfig {
	return ServiceConfig{
		Port:                 os.Getenv("PORT"),
		DeploymentServiceURL: os.Getenv("DEPLOYMENT_SERVICE_URL"),
		MonitoringServiceURL: os.Getenv("MONITORING_SERVICE_URL"),
		ConfigServiceURL:     os.Getenv("CONFIG_SERVICE_URL"),
		Timeout:              30 * time.Second,
	}
}

// NewServiceClient creates clients for interacting with microservices
func NewServiceClient(logger *zap.Logger) *ServiceClient {
	config := NewServiceConfig()

	// Set default URLs if env vars are empty
	if config.DeploymentServiceURL == "" {
		config.DeploymentServiceURL = os.Getenv("DEPLOYMENT_SERVICE_URL")
	}
	if config.MonitoringServiceURL == "" {
		config.MonitoringServiceURL = os.Getenv("MONITORING_SERVICE_URL")
	}
	if config.ConfigServiceURL == "" {
		config.ConfigServiceURL = os.Getenv("CONFIG_SERVICE_URL")
	}

	logger.Info("Service URLs",
		zap.String("deployment", config.DeploymentServiceURL),
		zap.String("monitoring", config.MonitoringServiceURL),
		zap.String("config", config.ConfigServiceURL),
	)

	return &ServiceClient{
		deploymentClient: &http.Client{Timeout: config.Timeout},
		monitoringClient: &http.Client{Timeout: config.Timeout},
		configClient:     &http.Client{Timeout: config.Timeout},
		logger:           logger,
		config:           config,
	}
}

// setupMetrics initializes Prometheus metrics for the API service
func setupMetrics() (*prometheus.Registry, gin.HandlerFunc) {
	registry := prometheus.NewRegistry()

	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_requests_total",
			Help: "Total number of requests processed by the REST API service",
		},
		[]string{"service", "method", "status"},
	)

	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method"},
	)

	registry.MustRegister(requestCounter, requestDuration)

	metricsMiddleware := func(c *gin.Context) {
		start := time.Now()
		service := extractServiceFromPath(c.FullPath())

		c.Next()

		duration := time.Since(start).Seconds()
		status := fmt.Sprintf("%d", c.Writer.Status())

		requestCounter.WithLabelValues(service, c.Request.Method, status).Inc()
		requestDuration.WithLabelValues(service, c.Request.Method).Observe(duration)
	}

	return registry, metricsMiddleware
}

// extractServiceFromPath determines the service based on the request path
func extractServiceFromPath(path string) string {
	switch {
	case path == "/metrics":
		return "metrics"
	case path == "/health":
		return "health"
	case path == "/api/deployments":
		return "deployment"
	case path == "/api/monitoring":
		return "monitoring"
	case path == "/api/configs":
		return "configuration"
	default:
		return "unknown"
	}
}

// Generic Proxy Handler
func (sc *ServiceClient) proxyRequest(targetBaseUrl string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract the rest of the path
		proxyPath := c.Param("proxyPath") // Assuming path is like /deployments/*proxyPath

		// Construct the target URL
		targetURL := fmt.Sprintf("%s%s", targetBaseUrl, proxyPath)
		if c.Request.URL.RawQuery != "" {
			targetURL = fmt.Sprintf("%s?%s", targetBaseUrl, c.Request.URL.RawQuery)
		}

		sc.logger.Debug("Proxying request",
			zap.String("method", c.Request.Method),
			zap.String("originalPath", c.Request.URL.Path),
			zap.String("targetURL", targetURL))

		ctx, cancel := context.WithTimeout(c.Request.Context(), sc.config.Timeout)
		defer cancel()

		// Create new request to target service
		req, err := http.NewRequestWithContext(ctx, c.Request.Method, targetURL, c.Request.Body)
		if err != nil {
			sc.logger.Error("Failed to create proxy request", zap.Error(err), zap.String("targetURL", targetURL))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error creating proxy request"})

			return
		}

		// Copy headers (including potential user/auth headers from gateway
		req.Header = c.Request.Header.Clone()

		// Execute the request
		client := &http.Client{Timeout: sc.config.Timeout}
		resp, err := client.Do(req)
		if err != nil {
			sc.logger.Error("Proxy request failed", zap.Error(err), zap.String("target", targetURL))
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Service Unavailable: %s", targetBaseUrl)})

			return
		}
		defer resp.Body.Close()

		// Copy response header back to client
		for key, values := range resp.Header {
			for _, value := range values {
				c.Header(key, value)
			}
		}

		// Send response back
		c.Status(resp.StatusCode)
		_, err = io.Copy(c.Writer, resp.Body)
		if err != nil {
			sc.logger.Error("Failed to copy response body", zap.Error(err))
			// Status is already set, difficult to change now. Log the error.
		}
	}
}

// Deployment Service APIs
func (sc *ServiceClient) registerDeploymentRoutes(r *gin.RouterGroup) {
	handler := sc.proxyRequest(sc.config.DeploymentServiceURL)
	r.Any("/deployments/*proxyPath", handler) // Capture all methods and subpaths
}

// Monitoring Service APIs
func (sc *ServiceClient) registerMonitoringRoutes(r *gin.RouterGroup) {
	handler := sc.proxyRequest(sc.config.MonitoringServiceURL)
	r.Any("/monitoring/*proxyPath", handler)
}

// Configuration Service APIs
func (sc *ServiceClient) registerConfigRoutes(r *gin.RouterGroup) {
	handler := sc.proxyRequest(sc.config.ConfigServiceURL)
	r.Any("/configs/*proxyPath", handler)
}

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Set up Prometheus registry and middleware
	registry, metricsMiddleware := setupMetrics()

	// Set Gin to release mode for production
	gin.SetMode(gin.ReleaseMode)

	// Create Gin router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(metricsMiddleware)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Metrics endpoint for Prometheus
	router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))

	// Service client for interacting with microservices
	serviceClient := NewServiceClient(logger)

	// API routes group
	apiRoutes := router.Group("/api/v1") // Base path for API Server's own endpoints if any, or just groups
	{
		serviceClient.registerDeploymentRoutes(apiRoutes)
		serviceClient.registerMonitoringRoutes(apiRoutes)
		serviceClient.registerConfigRoutes(apiRoutes)
	}

	// Start server
	server := &http.Server{
		Addr:         ":" + serviceClient.config.Port,
		Handler:      router,
		ReadTimeout:  serviceClient.config.Timeout,
		WriteTimeout: serviceClient.config.Timeout,
	}

	logger.Info("Starting REST API Service", zap.String("port", serviceClient.config.Port))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Server failed", zap.Error(err))
	}
}
