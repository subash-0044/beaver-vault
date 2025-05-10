package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/subash-0044/beaver-vault/pkg/handler"
)

// Server represents the HTTP server
type Server struct {
	handler *handler.Handler
	router  *gin.Engine
}

// NewGinServer creates a new HTTP server instance
func NewGinServer(h *handler.Handler) *Server {
	s := &Server{
		handler: h,
		router:  gin.Default(),
	}
	s.setupRoutes()
	return s
}

// setupRoutes configures all the routes for the server
func (s *Server) setupRoutes() {
	// Health check
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API routes
	v1 := s.router.Group("/api/v1")
	{
		// Key-Value operations
		v1.GET("/kv/:key", s.handleGet)
		v1.PUT("/kv/:key", s.handleSet)
		v1.DELETE("/kv/:key", s.handleDelete)
	}
}

// Run starts the HTTP server
func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}

// handleGet handles GET requests for key-value pairs
func (s *Server) handleGet(c *gin.Context) {
	key := c.Param("key")
	value, err := s.handler.Get(key)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if value == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"key": key, "value": value})
}

// handleSet handles PUT requests for key-value pairs
func (s *Server) handleSet(c *gin.Context) {
	key := c.Param("key")
	var value interface{}
	if err := c.BindJSON(&value); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	err := s.handler.Store(c.Request.Context(), handler.RequestStore{
		Key:   key,
		Value: value,
	})
	if err != nil {
		if err.Error() == "not the leader" {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "not the leader"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// handleDelete handles DELETE requests for key-value pairs
func (s *Server) handleDelete(c *gin.Context) {
	key := c.Param("key")
	err := s.handler.Delete(key)
	if err != nil {
		if err.Error() == "not the leader" {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "not the leader"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
