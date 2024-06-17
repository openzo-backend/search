package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/tanush-128/openzo_backend/search/internal/service"
)

type searchHandler struct {
	searchService service.SearchService
}

func NewSearchHandler(searchService service.SearchService) *searchHandler {
	return &searchHandler{searchService: searchService}
}

func (h *searchHandler) SearchStoresByPincode(ctx *gin.Context) {
	pincode := ctx.Param("pincode")
	term := ctx.Query("term")

	stores, err := h.searchService.SearchStoresByPincode(pincode, term)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, stores)

	// return h.searchService.SearchStoresByPincode(pincode, term)
}

func (h *searchHandler) SearchProductsByPincode(ctx *gin.Context) {
	pincode := ctx.Param("pincode")
	term := ctx.Query("term")

	products, err := h.searchService.SearchProductsByPincode(pincode, term)
	if err != nil {
		ctx.JSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, products)

	// return h.searchService.SearchProductsByPincode(pincode, term)
}
