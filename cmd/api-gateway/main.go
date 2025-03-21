package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// Configuration struct for API Gateway
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

// Middleware for authentication
func authMiddleware(authServiceURL string, logger *zap.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if token == "" {
				logger.Warn("Missing authentication token")
				http.Error(w, "Unauthorized: Missing token", http.StatusUnauthorized)
				return
			}

			// Create auth service request context with timeout
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()

			// Validate token with Auth Service
			req, err := http.NewRequestWithContext(ctx, "POST", authServiceURL+"/validate", nil)
			if err != nil {
				logger.Error("Failed to create auth request", zap.Error(err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			req.Header.Set("Authorization", "Bearer "+token)

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				logger.Error("Auth service request failed", zap.Error(err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				logger.Warn("Authentication failed", zap.Int("status_code", resp.StatusCode))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Extract user claims and add to request context
			var claims map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&claims); err != nil {
				logger.Error("Failed to decode auth response", zap.Error(err))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Set user info in context for downstream handlers
			ctx = context.WithValue(r.Context(), "user", claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Middleware for rate limiting
func rateLimitMiddleware(rps int, interval time.Duration) mux.MiddlewareFunc {
	limiter := rate.NewLimiter(rate.Every(interval), rps)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Middleware for request logging
func loggingMiddleware(logger *zap.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a custom response writer to capture status code
			crw := &customResponseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(crw, r)

			duration := time.Since(start)

			// Log request details
			logger.Info("API Request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.Int("status", crw.statusCode),
				zap.Duration("duration", duration),
			)
		})
	}
}

// Custom response writer to capture status code
type customResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (crw *customResponseWriter) WriteHeader(code int) {
	crw.statusCode = code
	crw.ResponseWriter.WriteHeader(code)
}

// Create a reverse proxy handler for service routes
func createProxyHandler(route ServiceRoute, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract path without the base path
		path := strings.TrimPrefix(r.URL.Path, route.PathBase)

		// Create the target URL
		targetURL := fmt.Sprintf("%s%s", route.URL, path)
		if r.URL.RawQuery != "" {
			targetURL = fmt.Sprintf("%s?%s", targetURL, r.URL.RawQuery)
		}

		// Create the outgoing request
		outReq, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)
		if err != nil {
			logger.Error("Failed to create proxy request",
				zap.String("target", targetURL),
				zap.Error(err),
			)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Copy headers
		for name, values := range r.Header {
			for _, value := range values {
				outReq.Header.Add(name, value)
			}
		}

		// Forward user context if available
		if user, ok := r.Context().Value("user").(map[string]interface{}); ok {
			if userID, ok := user["sub"].(string); ok {
				outReq.Header.Set("X-User-ID", userID)
			}
		}

		// Add X-Forwarded headers
		outReq.Header.Set("X-Forwarded-For", r.RemoteAddr)
		outReq.Header.Set("X-Forwarded-Proto", r.URL.Scheme)
		outReq.Header.Set("X-Forwarded-Host", r.Host)

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
			http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()

		// Copy response headers
		for name, values := range resp.Header {
			for _, value := range values {
				w.Header().Add(name, value)
			}
		}

		// Set status code
		w.WriteHeader(resp.StatusCode)

		// Copy response body
		if _, err := fmt.Fprintf(w, "%s", resp.Body); err != nil {
			logger.Error("Failed to write response", zap.Error(err))
		}
	}
}

// Set up Prometheus metrics
func setupMetrics() *prometheus.Registry {
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
	return registry
}

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Load configuration
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

	// Set up Prometheus registry
	registry := setupMetrics()

	// Create router
	router := mux.NewRouter()

	// Apply global middleware
	router.Use(loggingMiddleware(logger))
	router.Use(rateLimitMiddleware(config.RateLimit, time.Duration(config.RateLimitInterval)*time.Second))

	// Health check endpoint (no auth required)
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	}).Methods("GET")

	// Metrics endpoint (for Prometheus)
	router.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))

	// API routes with authentication
	apiRouter := router.PathPrefix("/api").Subrouter()
	apiRouter.Use(authMiddleware(config.AuthServiceURL, logger))

	// Register service routes
	for _, route := range routes {
		logger.Info("Registering route",
			zap.String("name", route.Name),
			zap.String("path", route.PathBase),
			zap.Strings("methods", route.Methods),
		)

		handler := createProxyHandler(route, logger)
		apiRouter.PathPrefix(route.PathBase).Handler(handler).Methods(route.Methods...)
	}

	// Start server
	server := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      router,
		ReadTimeout:  time.Duration(config.Timeout) * time.Second,
		WriteTimeout: time.Duration(config.Timeout) * time.Second,
	}

	logger.Info("Starting API Gateway", zap.String("port", config.Port))
	if err := server.ListenAndServe(); err != nil {
		logger.Fatal("Server failed", zap.Error(err))
	}
}
