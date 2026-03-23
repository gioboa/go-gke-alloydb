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

type createRegionRequest struct {
	RegionID   int64  `json:"region_id" binding:"required,gte=1"`
	RegionName string `json:"region_name" binding:"required,max=25"`
}

type updateRegionRequest struct {
	RegionName string `json:"region_name" binding:"required,max=25"`
}

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
	regions := r.Group("/regions")
	regions.GET("", func(c *gin.Context) {
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
	regions.GET("/:id", func(c *gin.Context) {
		if dbStore == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not configured"})
			return
		}

		id, ok := parseRegionID(c)
		if !ok {
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		region, err := dbStore.GetRegion(ctx, id)
		if err != nil {
			if store.IsNotFound(err) {
				c.JSON(http.StatusNotFound, gin.H{"error": "region not found"})
				return
			}
			log.Printf("get region %d: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "region query failed"})
			return
		}

		c.JSON(http.StatusOK, region)
	})
	regions.POST("", func(c *gin.Context) {
		if dbStore == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not configured"})
			return
		}

		var req createRegionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		region, err := dbStore.CreateRegion(ctx, store.Region{
			ID:   req.RegionID,
			Name: req.RegionName,
		})
		if err != nil {
			if store.IsUniqueViolation(err) {
				c.JSON(http.StatusConflict, gin.H{"error": "region already exists"})
				return
			}
			log.Printf("create region %d: %v", req.RegionID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "region create failed"})
			return
		}

		c.JSON(http.StatusCreated, region)
	})
	regions.PUT("/:id", func(c *gin.Context) {
		if dbStore == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not configured"})
			return
		}

		id, ok := parseRegionID(c)
		if !ok {
			return
		}

		var req updateRegionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		region, err := dbStore.UpdateRegion(ctx, id, req.RegionName)
		if err != nil {
			if store.IsNotFound(err) {
				c.JSON(http.StatusNotFound, gin.H{"error": "region not found"})
				return
			}
			log.Printf("update region %d: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "region update failed"})
			return
		}

		c.JSON(http.StatusOK, region)
	})
	regions.DELETE("/:id", func(c *gin.Context) {
		if dbStore == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database not configured"})
			return
		}

		id, ok := parseRegionID(c)
		if !ok {
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		region, err := dbStore.DeleteRegion(ctx, id)
		if err != nil {
			if store.IsNotFound(err) {
				c.JSON(http.StatusNotFound, gin.H{"error": "region not found"})
				return
			}
			log.Printf("delete region %d: %v", id, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "region delete failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"deleted": region})
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

func parseRegionID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid region id"})
		return 0, false
	}
	return id, true
}
