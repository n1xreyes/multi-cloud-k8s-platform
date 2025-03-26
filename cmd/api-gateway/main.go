package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

const internalServerErrorMessage = "Internal Server Error"

// Config struct for API Gateway
type Config struct {
	Port              string `json:"port"`
	AuthServiceURL    string `json:"auth_service_url"`
	APIServiceURL     string `json:"api_service_url"`
	RateLimit         int    `json:"rate_limit"`
	RateLimitInterval int    `json:"rate_limit_interval"`
	Timeout           int    `json:"timeout"`
}

// ServiceRoute defines a route to be proxied through the gateway
type ServiceRoute struct {
	Name     string
	PathBase string
	URL      string
	Methods  []string
}

// Initialize and return configuration from environment variables
func loadConfig() Config {
	config := Config{
		Port:              "8080",
		AuthServiceURL:    "http://auth-service:8080",
		APIServiceURL:     "http://api-service:8080",
		RateLimit:         100,
		RateLimitInterval: 1,
		Timeout:           30,
	}

	// Override with environment variables if provided
	if port := os.Getenv("PORT"); port != "" {
		config.Port = port
	}
	if authURL := os.Getenv("AUTH_SERVICE_URL"); authURL != "" {
		config.AuthServiceURL = authURL
	}
	if apiURL := os.Getenv("API_SERVICE_URL"); apiURL != "" {
		config.APIServiceURL = apiURL
	}

	return config
}

// Middleware for authentication
func authMiddleware(authServiceURL string, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if token == "" {
			logger.Warn("Missing authentication token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Missing token"})
			return
		}

		// Create auth service request context with timeout
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		// Validate token with Auth Service
		req, err := http.NewRequestWithContext(ctx, "POST", authServiceURL+"/validate", nil)
		if err != nil {
			logger.Error("Failed to create auth request", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": internalServerErrorMessage})
			return
		}
		req.Header.Set("Authorization", "Bearer "+token)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			logger.Error("Auth service request failed", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": internalServerErrorMessage})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logger.Warn("Authentication failed", zap.Int("status_code", resp.StatusCode))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		// Extract user claims and add to request context
		var claims map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
			logger.Error("Failed to decode auth response", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": internalServerErrorMessage})
			return
		}

		// Set user info in context for downstream handlers
		c.Set("user", claims)
		c.Next()
	}
}

// Middleware for rate limiting
func rateLimitMiddleware(rps int, interval time.Duration) gin.HandlerFunc {
	limiter := rate.NewLimiter(rate.Every(interval), rps)

	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			return
		}
		c.Next()
	}
}

// Middleware for request logging
func loggingMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// After request is processed
		duration := time.Since(start)

		// Log request details
		logger.Info("API Request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("remote_addr", c.ClientIP()),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("duration", duration),
		)
	}
}

// Create a reverse proxy handler for service routes
func createProxyHandler(route ServiceRoute, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract path without the base path
		path := strings.TrimPrefix(c.Request.URL.Path, route.PathBase)

		// Create the target URL
		targetURL := fmt.Sprintf("%s%s", route.URL, path)
		if c.Request.URL.RawQuery != "" {
			targetURL = fmt.Sprintf("%s?%s", targetURL, c.Request.URL.RawQuery)
		}

		// Create the outgoing request
		outReq, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, targetURL, c.Request.Body)
		if err != nil {
			logger.Error("Failed to create proxy request",
				zap.String("target", targetURL),
				zap.Error(err),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": internalServerErrorMessage})
			return
		}

		// Copy headers
		copyHeaders(c.Request.Header, outReq.Header)

		// Forward user context if available
		forwardUserContext(c, outReq)

		// Add X-Forwarded headers
		addForwardedHeaders(c, outReq)

		// Send the request to the target service
		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		resp, err := client.Do(outReq)
		if err != nil {
			logger.Error("Proxy request failed",
				zap.String("service", route.Name),
				zap.String("target", targetURL),
				zap.Error(err),
			)
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Service Unavailable"})
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		copyHeaders(resp.Header, c.Writer.Header())

		// Set status code and copy response body
		c.Status(resp.StatusCode)
		io.Copy(c.Writer, resp.Body)
	}
}

// Helper function to copy headers
func copyHeaders(src, dst http.Header) {
	for name, values := range src {
		for _, value := range values {
			dst.Add(name, value)
		}
	}
}

// Helper function to forward user context
func forwardUserContext(c *gin.Context, req *http.Request) {
	if user, exists := c.Get("user"); exists {
		if userMap, ok := user.(map[string]interface{}); ok {
			if userID, ok := userMap["sub"].(string); ok {
				req.Header.Set("X-User-ID", userID)
			}
		}
	}
}

// Helper function to add forwarded headers
func addForwardedHeaders(c *gin.Context, req *http.Request) {
	req.Header.Set("X-Forwarded-For", c.ClientIP())
	req.Header.Set("X-Forwarded-Proto", c.Request.URL.Scheme)
	req.Header.Set("X-Forwarded-Host", c.Request.Host)
}

// Set up Prometheus metrics
func setupMetrics() (*prometheus.Registry, gin.HandlerFunc) {
	registry := prometheus.NewRegistry()

	// Request counter
	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_requests_total",
			Help: "Total number of requests processed by the API Gateway",
		},
		[]string{"method", "path", "status"},
	)

	// Request duration
	requestDuration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gateway_request_duration_seconds",
			Help:    "Request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	registry.MustRegister(requestCounter, requestDuration)

	// Create Gin middleware to update metrics
	metricMiddleware := func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()

		c.Next()

		duration := time.Since(start).Seconds()
		status := fmt.Sprintf("%d", c.Writer.Status())

		requestCounter.WithLabelValues(c.Request.Method, path, status).Inc()
		requestDuration.WithLabelValues(c.Request.Method, path).Observe(duration)
	}

	return registry, metricMiddleware
}

// Register the service routes
func registerRoutes(engine *gin.Engine, routes []ServiceRoute, logger *zap.Logger) {
	// API routes with authentication
	api := engine.Group("/api")
	{
		// Register service routes
		for _, route := range routes {
			logger.Info("Registering route",
				zap.String("name", route.Name),
				zap.String("path", route.PathBase),
				zap.Strings("methods", route.Methods),
			)

			handler := createProxyHandler(route, logger)
			routePath := strings.TrimPrefix(route.PathBase, "/api") + "/*path"

			for _, method := range route.Methods {
				switch method {
				case "GET":
					api.GET(routePath, handler)
				case "POST":
					api.POST(routePath, handler)
				case "PUT":
					api.PUT(routePath, handler)
				case "DELETE":
					api.DELETE(routePath, handler)
				case "PATCH":
					api.PATCH(routePath, handler)
				}
			}
		}
	}
}

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Load configuration
	config := loadConfig()

	// Define service routes
	routes := []ServiceRoute{
		{
			Name:     "API Service",
			PathBase: "/api/v1",
			URL:      config.APIServiceURL,
			Methods:  []string{"GET", "POST", "PUT", "DELETE"},
		},
		{
			Name:     "Deployment Service",
			PathBase: "/api/v1/deployments",
			URL:      "http://deployment-service:8080",
			Methods:  []string{"GET", "POST", "PUT", "DELETE"},
		},
		{
			Name:     "Monitoring Service",
			PathBase: "/api/v1/monitoring",
			URL:      "http://monitoring-service:8080",
			Methods:  []string{"GET"},
		},
		{
			Name:     "Configuration Service",
			PathBase: "/api/v1/configs",
			URL:      "http://configuration-service:8080",
			Methods:  []string{"GET", "POST", "PUT", "DELETE"},
		},
	}

	// Set up Prometheus registry and middleware
	registry, metricsMiddleware := setupMetrics()

	// Set Gin to release mode for production
	gin.SetMode(gin.ReleaseMode)

	// Create Gin router
	router := gin.New()

	// Apply global middleware
	router.Use(gin.Recovery())
	router.Use(loggingMiddleware(logger))
	router.Use(rateLimitMiddleware(config.RateLimit, time.Duration(config.RateLimitInterval)*time.Second))
	router.Use(metricsMiddleware)

	// Health check endpoint (no auth required)
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Metrics endpoint (for Prometheus)
	router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(registry, promhttp.HandlerOpts{})))

	// Apply authentication middleware to API endpoints
	authGroup := router.Group("/api")
	authGroup.Use(authMiddleware(config.AuthServiceURL, logger))

	// Register service routes
	registerRoutes(router, routes, logger)

	// Start server
	server := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      router,
		ReadTimeout:  time.Duration(config.Timeout) * time.Second,
		WriteTimeout: time.Duration(config.Timeout) * time.Second,
	}

	logger.Info("Starting API Gateway", zap.String("port", config.Port))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("Server failed", zap.Error(err))
	}
}
