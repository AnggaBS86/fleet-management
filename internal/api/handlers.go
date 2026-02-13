package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"leet-management/internal/db"
)

type Server struct {
	store *db.Store
}

func NewServer(store *db.Store) *Server {
	return &Server{store: store}
}

func (s *Server) RegisterRoutes(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/vehicles/:vehicle_id/location", s.getLocation)
	r.GET("/vehicles/:vehicle_id/history", s.getHistory)
}

func (s *Server) getLocation(c *gin.Context) {
	vehicleID := c.Param("vehicle_id")
	ctx, cancel := db.WithTimeout(c.Request.Context())
	defer cancel()

	loc, err := s.store.GetLatest(ctx, vehicleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "location not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"vehicle_id": loc.VehicleID,
		"latitude":   loc.Latitude,
		"longitude":  loc.Longitude,
		"timestamp":  loc.Timestamp,
	})
}

func (s *Server) getHistory(c *gin.Context) {
	vehicleID := c.Param("vehicle_id")
	startStr := c.Query("start")
	endStr := c.Query("end")

	start, err := strconv.ParseInt(startStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start"})
		return
	}
	end, err := strconv.ParseInt(endStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end"})
		return
	}

	ctx, cancel := db.WithTimeout(c.Request.Context())
	defer cancel()

	locations, err := s.store.GetHistory(ctx, vehicleID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"vehicle_id": vehicleID,
		"items":      locations,
	})
}
