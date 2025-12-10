package api

import (
	"encoding/json"
	"net/http"
	"os"

	"aoa-inventory/models"
	"aoa-inventory/utils"

	"github.com/gin-gonic/gin"
)

var (
	log             = utils.NewLogger("API")
	sampleInventory []models.InventoryItem
)

func SetupEndpoints(apiGroup *gin.RouterGroup) {
	sampleInventory = loadSampleInventory("data.json")

	apiGroup.GET("/inventory", GetInventory)
}

func GetInventory(ginContext *gin.Context) {
	ginContext.JSON(http.StatusOK, sampleInventory)
}

func loadSampleInventory(path string) []models.InventoryItem {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Could not read %s: %v", path, err)
		return []models.InventoryItem{}
	}

	var items []models.InventoryItem
	if err := json.Unmarshal(data, &items); err != nil {
		log.Printf("Could not parse %s: %v", path, err)
		return []models.InventoryItem{}
	}

	return items
}
