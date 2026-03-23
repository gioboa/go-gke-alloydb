package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gioboa/go-gke-alloydb/internal/config"
	"github.com/gioboa/go-gke-alloydb/internal/store"
)

func main() {
	cfg := config.Load()
	r := gin.Default()
	if err := r.SetTrustedProxies(nil); err != nil {
		log.Fatalf("set trusted proxies: %v", err)
	}
	var dbStore *store.Store

	if cfg.DatabaseURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var err error
		dbStore, err = store.Open(ctx, cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("open db: %v", err)
		}
		defer dbStore.Close()
	}

	r.GET("/healthz/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/healthz/ready", func(c *gin.Context) {
		if dbStore == nil {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := dbStore.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "database unavailable"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	r.GET("/db/ping", func(c *gin.Context) {
		if dbStore == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not configured"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		message, err := dbStore.Queries.DatabasePing(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "database query failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": message,
		})
	})
	r.GET("/regions", func(c *gin.Context) {
		if dbStore == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not configured"})
			return
		}

		limit := 100
		if raw := c.Query("limit"); raw != "" {
			n, err := strconv.Atoi(raw)
			if err != nil || n < 1 || n > 500 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 500"})
				return
			}
			limit = n
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		regions, err := dbStore.ListRegions(ctx, limit)
		if err != nil {
			log.Printf("list regions: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "region query failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"count":   len(regions),
			"regions": regions,
		})
	})

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
