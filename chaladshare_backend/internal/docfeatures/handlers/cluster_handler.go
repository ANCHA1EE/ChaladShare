package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"chaladshare_backend/internal/docfeatures/models"
	"chaladshare_backend/internal/docfeatures/service"

	"github.com/gin-gonic/gin"
)

type FeatureHandler struct {
	svc service.FeatureService
}

func NewFeatureHandler(svc service.FeatureService) *FeatureHandler {
	return &FeatureHandler{svc: svc}
}

func (h *FeatureHandler) GetVectors(c *gin.Context) {
	label := strings.TrimSpace(c.Query("label"))
	onlyUn := c.Query("only_unclustered") == "1" || strings.ToLower(c.Query("only_unclustered")) == "true"

	if label == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing label"})
		return
	}
	if label != "typed" && label != "handwritten" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "label must be typed or handwritten"})
		return
	}

	items, err := h.svc.ListVectors(label, onlyUn)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.VectorsResp{Items: items})
}

func (h *FeatureHandler) BatchUpdateClusters(c *gin.Context) {
	var req models.BatchUpdateClustersReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json", "detail": err.Error()})
		return
	}
	if len(req.Updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "updates is empty"})
		return
	}

	updated, err := h.svc.BatchUpdateClusters(req.Updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.BatchUpdateClustersResp{Updated: updated})
}

func (h *FeatureHandler) RunClustering(c *gin.Context) {
	label := strings.TrimSpace(c.Query("label"))
	onlyUn := c.Query("only_unclustered") == "1" || strings.ToLower(c.Query("only_unclustered")) == "true"

	k := 8
	if ks := strings.TrimSpace(c.Query("k")); ks != "" {
		if v, err := strconv.Atoi(ks); err == nil {
			k = v
		}
	}

	updated, err := h.svc.RunClustering(label, onlyUn, k)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"label":            label,
		"only_unclustered": onlyUn,
		"k":                k,
		"updated":          updated,
	})
}
