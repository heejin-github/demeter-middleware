package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"log"
	"time"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"
)

var (
	requestCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latencies in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	rateLimiter = rate.NewLimiter(rate.Every(time.Second), 10000) // 10000 requests per second
)

func init() {
	prometheus.MustRegister(requestCounter)
	prometheus.MustRegister(requestDuration)
}

func startMiddleware() error {
	r := gin.Default()

	r.Use(metricsMiddleware())
	r.Use(rateLimitMiddleware())

	r.Any("/auth/:apikey", handleAuth)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.GET("/health", handleHealth)

	return r.Run(":3000")
}

func metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)

		status := strconv.Itoa(c.Writer.Status())
		requestCounter.WithLabelValues(c.Request.Method, c.FullPath(), status).Inc()
		requestDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration.Seconds())
	}
}

func rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rateLimiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func handleAuth(c *gin.Context) {
	apiKey := c.Param("apikey")
	backend := c.Query("backend")
	path := c.Query("path")

	log.Printf("Received request - API Key: %s, Backend: %s, Path: %s", apiKey, backend, path)

	if !isValidAPIKey(apiKey) {
		log.Printf("Invalid API key attempt: %s", apiKey)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
		return
	}

	username, protocol, err := getAPIKeyDetails(apiKey)
	if err != nil {
		log.Printf("Error retrieving API key details: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	log.Printf("API Key details - Username: %s, Protocol: %s", username, protocol)

	var targetURL string
	switch backend {
	case "8545":
		targetURL = "http://localhost:8545"
	case "5052":
		targetURL = "http://localhost:5052"
	default:
		log.Printf("Invalid backend request: %s", backend)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid backend"})
		return
	}

	remote, err := url.Parse(targetURL)
	if err != nil {
		log.Printf("Error parsing target URL: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if path != "" && !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	log.Printf("Proxying request to %s%s", targetURL, path)
	proxy := httputil.NewSingleHostReverseProxy(remote)
	c.Request.URL.Path = path
	proxy.ServeHTTP(c.Writer, c.Request)
}

func handleHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
