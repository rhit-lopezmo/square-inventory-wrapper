package api

import (
	"errors"
	"net/http"

	"aoa-inventory/squareUtils"
	"aoa-inventory/squareUtils/models"
	"aoa-inventory/utils"

	"github.com/gin-gonic/gin"
)

var log = utils.NewLogger("API")

func SetupEndpoints(apiGroup *gin.RouterGroup) {
	apiGroup.GET("/inventory", GetInventory)
	apiGroup.PUT("/inventory/:sku", UpdateInventoryItem)
}

func GetInventory(ctx *gin.Context) {
	log.Println("Getting inventory...")

	sampleInventory := squareUtils.LoadSampleInventory("data.json")

	ctx.JSON(http.StatusOK, sampleInventory)
}

func UpdateInventoryItem(ctx *gin.Context) {
	sku := ctx.Param("sku")
	if sku == "" {
		log.Println("ERROR: No SKU provided")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "sku is required"})
		return
	}

	var updatePayload models.InventoryItemUpdate
	if err := ctx.ShouldBindJSON(&updatePayload); err != nil {
		log.Printf("ERROR: Failed to bind request body: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	savedItem, err := squareUtils.UpdateInventoryItem("data.json", sku, &updatePayload)
	if err != nil {
		if errors.Is(err, squareUtils.ErrInventoryItemNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
			return
		}

		log.Printf("ERROR: Failed to update inventory item: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not update inventory item"})
		return
	}

	ctx.JSON(http.StatusOK, savedItem)
}
