package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gioboa/go-gke-alloydb/internal/config"
	"github.com/gioboa/go-gke-alloydb/internal/store"
)

func main() {
	cfg := config.Load()
	r := gin.Default()
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

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
