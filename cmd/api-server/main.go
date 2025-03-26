package main

import (
	"context"
	"fmt"
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
	return &ServiceClient{
		deploymentClient: &http.Client{Timeout: 30 * time.Second},
		monitoringClient: &http.Client{Timeout: 30 * time.Second},
		configClient:     &http.Client{Timeout: 30 * time.Second},
		logger:           logger,
		config:           NewServiceConfig(),
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

// Deployment Service APIs
func (sc *ServiceClient) registerDeploymentRoutes(r *gin.RouterGroup) {
	// Get all deployments
	r.GET("/deployments", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), sc.config.Timeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", sc.config.DeploymentServiceURL+"/deployments", nil)
		if err != nil {
			sc.logger.Error("Failed to create deployment request", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}

		resp, err := sc.deploymentClient.Do(req)
		if err != nil {
			sc.logger.Error("Deployment service request failed", zap.Error(err))
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Deployment Service Unavailable"})
			return
		}
		defer resp.Body.Close()

		c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
	})

	// Create a new deployment
	r.POST("/deployments", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), sc.config.Timeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "POST", sc.config.DeploymentServiceURL+"/deployments", c.Request.Body)
		if err != nil {
			sc.logger.Error("Failed to create deployment request", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}

		resp, err := sc.deploymentClient.Do(req)
		if err != nil {
			sc.logger.Error("Deployment service request failed", zap.Error(err))
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Deployment Service Unavailable"})
			return
		}
		defer resp.Body.Close()

		c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
	})
}

// Monitoring Service APIs
func (sc *ServiceClient) registerMonitoringRoutes(r *gin.RouterGroup) {
	// Get cluster metrics
	r.GET("/monitoring/metrics", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), sc.config.Timeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", sc.config.MonitoringServiceURL+"/metrics", nil)
		if err != nil {
			sc.logger.Error("Failed to create monitoring request", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}

		resp, err := sc.monitoringClient.Do(req)
		if err != nil {
			sc.logger.Error("Monitoring service request failed", zap.Error(err))
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Monitoring Service Unavailable"})
			return
		}
		defer resp.Body.Close()

		c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
	})
}

// Configuration Service APIs
func (sc *ServiceClient) registerConfigRoutes(r *gin.RouterGroup) {
	// Get configurations
	r.GET("/configs", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), sc.config.Timeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "GET", sc.config.ConfigServiceURL+"/configs", nil)
		if err != nil {
			sc.logger.Error("Failed to create config request", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}

		resp, err := sc.configClient.Do(req)
		if err != nil {
			sc.logger.Error("Config service request failed", zap.Error(err))
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Configuration Service Unavailable"})
			return
		}
		defer resp.Body.Close()

		c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
	})

	// Create a new configuration
	r.POST("/configs", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), sc.config.Timeout)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, "POST", sc.config.ConfigServiceURL+"/configs", c.Request.Body)
		if err != nil {
			sc.logger.Error("Failed to create config request", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}

		resp, err := sc.configClient.Do(req)
		if err != nil {
			sc.logger.Error("Config service request failed", zap.Error(err))
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Configuration Service Unavailable"})
			return
		}
		defer resp.Body.Close()

		c.DataFromReader(resp.StatusCode, resp.ContentLength, resp.Header.Get("Content-Type"), resp.Body, nil)
	})
}

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Load configuration
	config := NewServiceConfig()

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
	apiRoutes := router.Group("/api/v1")
	{
		serviceClient.registerDeploymentRoutes(apiRoutes)
		serviceClient.registerMonitoringRoutes(apiRoutes)
		serviceClient.registerConfigRoutes(apiRoutes)
	}

	// Start server
	server := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      router,
		ReadTimeout:  config.Timeout,
		WriteTimeout: config.Timeout,
	}

	logger.Info("Starting REST API Service", zap.String("port", config.Port))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Server failed", zap.Error(err))
	}
}
