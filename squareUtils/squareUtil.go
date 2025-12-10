package squareUtils

import (
	"aoa-inventory/squareUtils/models"
	"aoa-inventory/utils"
	"encoding/json"
	"errors"
	"os"
)

var log = utils.NewLogger("SQUARE-UTILS")

var ErrInventoryItemNotFound = errors.New("inventory item not found")

func LoadSampleInventory(path string) []models.InventoryItem {
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

func UpdateInventoryItem(path string, sku string, update *models.InventoryItemUpdate) (*models.InventoryItem, error) {
	if update == nil {
		return nil, errors.New("update is required")
	}

	if sku == "" {
		return nil, errors.New("sku is required")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var items []models.InventoryItem
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}

	var updatedItem models.InventoryItem
	itemUpdated := false
	for i := range items {
		if items[i].SKU == sku {
			items[i].ApplyUpdate(update)
			updatedItem = items[i]
			itemUpdated = true
			break
		}
	}

	if !itemUpdated {
		return nil, ErrInventoryItemNotFound
	}

	updatedData, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(path, updatedData, 0644); err != nil {
		return nil, err
	}

	return &updatedItem, nil
}
