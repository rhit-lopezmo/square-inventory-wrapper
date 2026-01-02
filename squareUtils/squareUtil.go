package squareUtils

import (
	"aoa-inventory/config"
	"aoa-inventory/squareUtils/client"
	"aoa-inventory/squareUtils/models"
	"aoa-inventory/utils"
	"context"
	"encoding/json"
	"errors"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	square "github.com/square/square-go-sdk"
)

var log = utils.NewLogger("SQUARE-UTILS")

var ErrInventoryItemNotFound = errors.New("inventory item not found")

func LoadSampleInventory(path string) []models.InventoryItem {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Could not read %s: %v", path, err)
		return []models.InventoryItem{}
	}

	items := []models.InventoryItem{}
	if err := json.Unmarshal(data, &items); err != nil {
		log.Printf("Could not parse %s: %v", path, err)
		return []models.InventoryItem{}
	}

	return items
}

func LoadInventory() []models.InventoryItem {
	sqClient := client.SquareClient

	if sqClient == nil {
		log.Println("ERROR: Square client is not initialized")
		return []models.InventoryItem{}
	}

	ctx := context.Background()
	items := []models.InventoryItem{}

	countReq := &square.BatchGetInventoryCountsRequest{
		LocationIDs: []string{config.SquareLocationID},
		States:      []square.InventoryState{square.InventoryStateInStock},
	}

	countResp, err := sqClient.Inventory.BatchGetCounts(ctx, countReq)
	if err != nil {
		log.Printf("ERROR: Failed to fetch inventory from Square: %v", err)
		return []models.InventoryItem{}
	}

	variationCounts := map[string]int{}

	for _, count := range countResp.Counts {
		if count == nil || count.CatalogObjectID == nil || count.Quantity == nil {
			continue
		}

		qty, err := strconv.ParseFloat(*count.Quantity, 64)
		if err != nil {
			log.Printf("ERROR: Could not parse quantity for catalog object %s: %v", *count.CatalogObjectID, err)
			continue
		}

		variationCounts[*count.CatalogObjectID] = int(math.Round(qty))
	}

	// Fetch catalog metadata to enrich counts with names/descriptions/SKUs/categories/images.
	listReq := &square.ListCatalogRequest{
		Types: square.String("ITEM,ITEM_VARIATION,IMAGE,CATEGORY"),
	}

	catalogPage, err := sqClient.Catalog.List(ctx, listReq)
	if err != nil {
		log.Printf("ERROR: Failed to fetch catalog items from Square: %v", err)
		return []models.InventoryItem{}
	}

	type itemMeta struct {
		name         string
		desc         string
		categoryID   string
		categoryName string
		imageIDs     []string
	}

	type variationMeta struct {
		itemID        string
		variationName string
		sku           string
		imageID       string
	}

	imageURLs := map[string]string{}
	categoryNames := map[string]string{}
	itemsMeta := map[string]itemMeta{}
	variationDetails := map[string]variationMeta{}

	// Build maps from the catalog objects we received.
	for _, obj := range catalogPage.Results {
		if obj == nil {
			continue
		}

		switch obj.GetType() {
		case "IMAGE":
			if obj.Image != nil && obj.Image.ImageData != nil && obj.Image.ImageData.URL != nil {
				imageURLs[obj.Image.ID] = *obj.Image.ImageData.URL
			}
		case "CATEGORY":
			if obj.Category != nil && obj.Category.CategoryData != nil && obj.Category.CategoryData.Name != nil && obj.Category.ID != nil {
				categoryNames[*obj.Category.ID] = *obj.Category.CategoryData.Name
			}
		case "ITEM":
			if obj.Item != nil && obj.Item.ItemData != nil {
				itemData := obj.Item.ItemData
				meta := itemMeta{}
				if itemData.Name != nil {
					meta.name = *itemData.Name
				}
				if itemData.Description != nil {
					meta.desc = *itemData.Description
				}
				if itemData.CategoryID != nil {
					meta.categoryID = *itemData.CategoryID
				} else if len(itemData.Categories) > 0 && itemData.Categories[0] != nil && itemData.Categories[0].CategoryData != nil && itemData.Categories[0].CategoryData.Name != nil {
					if name := itemData.Categories[0].CategoryData.GetName(); name != nil {
						meta.categoryName = *name
					}
				}
				if len(itemData.ImageIDs) > 0 {
					meta.imageIDs = itemData.ImageIDs
				}
				itemsMeta[obj.Item.ID] = meta
			}
		case "ITEM_VARIATION":
			if obj.ItemVariation != nil && obj.ItemVariation.ItemVariationData != nil {
				vData := obj.ItemVariation.ItemVariationData
				meta := variationMeta{}
				if vData.ItemID != nil {
					meta.itemID = *vData.ItemID
				}
				if vData.Name != nil {
					meta.variationName = *vData.Name
				}
				if vData.Sku != nil && *vData.Sku != "" {
					meta.sku = *vData.Sku
				} else {
					meta.sku = obj.ItemVariation.ID
				}
				if obj.ItemVariation.ImageID != nil {
					meta.imageID = *obj.ItemVariation.ImageID
				}
				variationDetails[obj.ItemVariation.ID] = meta
			}
		}
	}

	for variationID, stock := range variationCounts {
		meta := variationDetails[variationID]
		parent := itemsMeta[meta.itemID]

		name := parent.name
		if name == "" {
			name = meta.variationName
		}
		if name == "" {
			name = variationID
		}

		// Prefer variation image, then item image list.
		imageURL := ""
		if meta.imageID != "" {
			imageURL = imageURLs[meta.imageID]
		}
		if imageURL == "" && len(parent.imageIDs) > 0 {
			if url, ok := imageURLs[parent.imageIDs[0]]; ok {
				imageURL = url
			}
		}

		categoryName := ""
		if parent.categoryID != "" {
			categoryName = categoryNames[parent.categoryID]
		} else if parent.categoryName != "" {
			categoryName = parent.categoryName
		}

		items = append(items, models.InventoryItem{
			ID:           variationID,
			Name:         name,
			Description:  parent.desc,
			SKU:          meta.sku,
			CurrentStock: stock,
			ImageURL:     imageURL,
			Category:     categoryName,
		})
	}

	log.Printf("Loaded %d inventory items from Square", len(items))

	return items
}

func UpdateSampleInventoryItem(path string, sku string, update *models.InventoryItemUpdate) (*models.InventoryItem, error) {
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

func UpdateInventoryItem(path string, sku string, update *models.InventoryItemUpdate) (*models.InventoryItem, error) {
	if update == nil {
		return nil, errors.New("update is required")
	}

	if sku == "" {
		return nil, errors.New("sku is required")
	}

	if update.CurrentStock == nil {
		return nil, errors.New("currentStock is required for inventory update")
	}

	sqClient := client.SquareClient
	if sqClient == nil {
		return nil, errors.New("square client is not initialized")
	}

	ctx := context.Background()

	// Fetch catalog to map SKU -> variation and enrich response.
	listReq := &square.ListCatalogRequest{
		Types: square.String("ITEM,ITEM_VARIATION,IMAGE,CATEGORY"),
	}

	catalogPage, err := sqClient.Catalog.List(ctx, listReq)
	if err != nil {
		return nil, err
	}

	type itemMeta struct {
		name         string
		desc         string
		categoryID   string
		categoryName string
		imageIDs     []string
	}

	type variationMeta struct {
		itemID        string
		variationName string
		sku           string
		imageID       string
	}

	imageURLs := map[string]string{}
	categoryNames := map[string]string{}
	itemsMeta := map[string]itemMeta{}
	variationDetails := map[string]variationMeta{}

	for _, obj := range catalogPage.Results {
		if obj == nil {
			continue
		}

		switch obj.GetType() {
		case "IMAGE":
			if obj.Image != nil && obj.Image.ImageData != nil && obj.Image.ImageData.URL != nil {
				imageURLs[obj.Image.ID] = *obj.Image.ImageData.URL
			}
		case "CATEGORY":
			if obj.Category != nil && obj.Category.CategoryData != nil && obj.Category.CategoryData.Name != nil && obj.Category.ID != nil {
				categoryNames[*obj.Category.ID] = *obj.Category.CategoryData.Name
			}
		case "ITEM":
			if obj.Item != nil && obj.Item.ItemData != nil {
				itemData := obj.Item.ItemData
				meta := itemMeta{}
				if itemData.Name != nil {
					meta.name = *itemData.Name
				}
				if itemData.Description != nil {
					meta.desc = *itemData.Description
				}
				if itemData.CategoryID != nil {
					meta.categoryID = *itemData.CategoryID
				} else if len(itemData.Categories) > 0 && itemData.Categories[0] != nil && itemData.Categories[0].CategoryData != nil && itemData.Categories[0].CategoryData.Name != nil {
					if name := itemData.Categories[0].CategoryData.GetName(); name != nil {
						meta.categoryName = *name
					}
				}
				if len(itemData.ImageIDs) > 0 {
					meta.imageIDs = itemData.ImageIDs
				}
				itemsMeta[obj.Item.ID] = meta
			}
		case "ITEM_VARIATION":
			if obj.ItemVariation != nil && obj.ItemVariation.ItemVariationData != nil {
				vData := obj.ItemVariation.ItemVariationData
				meta := variationMeta{}
				if vData.ItemID != nil {
					meta.itemID = *vData.ItemID
				}
				if vData.Name != nil {
					meta.variationName = *vData.Name
				}
				if vData.Sku != nil && *vData.Sku != "" {
					meta.sku = *vData.Sku
				} else {
					meta.sku = obj.ItemVariation.ID
				}
				if obj.ItemVariation.ImageID != nil {
					meta.imageID = *obj.ItemVariation.ImageID
				}
				variationDetails[obj.ItemVariation.ID] = meta
			}
		}
	}

	// Find the variation ID for this SKU.
	var targetVariationID string
	for variationID, meta := range variationDetails {
		if meta.sku == sku {
			targetVariationID = variationID
			break
		}
	}

	if targetVariationID == "" {
		return nil, ErrInventoryItemNotFound
	}

	// Fetch current count for that variation to compute delta.
	countReq := &square.BatchGetInventoryCountsRequest{
		CatalogObjectIDs: []string{targetVariationID},
		LocationIDs:      []string{config.SquareLocationID},
		States:           []square.InventoryState{square.InventoryStateInStock},
	}

	countResp, err := sqClient.Inventory.BatchGetCounts(ctx, countReq)
	if err != nil {
		return nil, err
	}

	currentQty := 0
	if len(countResp.Counts) > 0 && countResp.Counts[0] != nil && countResp.Counts[0].Quantity != nil {
		parsed, err := strconv.ParseFloat(*countResp.Counts[0].Quantity, 64)
		if err == nil {
			currentQty = int(math.Round(parsed))
		}
	}

	newQty := *update.CurrentStock
	delta := newQty - currentQty

	if delta == 0 {
		// Nothing to change; return current state.
		meta := variationDetails[targetVariationID]
		parent := itemsMeta[meta.itemID]

		name := parent.name
		if name == "" {
			name = meta.variationName
		}
		if name == "" {
			name = targetVariationID
		}

		imageURL := ""
		if meta.imageID != "" {
			imageURL = imageURLs[meta.imageID]
		}
		if imageURL == "" && len(parent.imageIDs) > 0 {
			if url, ok := imageURLs[parent.imageIDs[0]]; ok {
				imageURL = url
			}
		}

		categoryName := ""
		if parent.categoryID != "" {
			categoryName = categoryNames[parent.categoryID]
		} else if parent.categoryName != "" {
			categoryName = parent.categoryName
		}

		return &models.InventoryItem{
			ID:           targetVariationID,
			Name:         name,
			Description:  parent.desc,
			SKU:          meta.sku,
			CurrentStock: currentQty,
			ImageURL:     imageURL,
			Category:     categoryName,
		}, nil
	}

	absDelta := int(math.Abs(float64(delta)))
	quantityStr := strconv.Itoa(absDelta)
	fromState := square.InventoryStateNone
	toState := square.InventoryStateInStock

	if delta < 0 {
		fromState = square.InventoryStateInStock
		toState = square.InventoryStateSold
	}

	adjustment := &square.InventoryAdjustment{
		CatalogObjectID: square.String(targetVariationID),
		LocationID:      square.String(config.SquareLocationID),
		FromState:       &fromState,
		ToState:         &toState,
		Quantity:        square.String(quantityStr),
		OccurredAt:      square.String(time.Now().UTC().Format(time.RFC3339)),
	}

	changeType := square.InventoryChangeTypeAdjustment
	change := &square.InventoryChange{
		Type:       &changeType,
		Adjustment: adjustment,
	}

	batchReq := &square.BatchChangeInventoryRequest{
		IdempotencyKey: uuid.NewString(),
		Changes:        []*square.InventoryChange{change},
	}

	if _, err := sqClient.Inventory.BatchCreateChanges(ctx, batchReq); err != nil {
		return nil, err
	}

	// Return the updated item with new stock.
	meta := variationDetails[targetVariationID]
	parent := itemsMeta[meta.itemID]

	name := parent.name
	if name == "" {
		name = meta.variationName
	}
	if name == "" {
		name = targetVariationID
	}

	imageURL := ""
	if meta.imageID != "" {
		imageURL = imageURLs[meta.imageID]
	}
	if imageURL == "" && len(parent.imageIDs) > 0 {
		if url, ok := imageURLs[parent.imageIDs[0]]; ok {
			imageURL = url
		}
	}

	categoryName := ""
	if parent.categoryID != "" {
		categoryName = categoryNames[parent.categoryID]
	} else if parent.categoryName != "" {
		categoryName = parent.categoryName
	}

	return &models.InventoryItem{
		ID:           targetVariationID,
		Name:         name,
		Description:  parent.desc,
		SKU:          meta.sku,
		CurrentStock: newQty,
		ImageURL:     imageURL,
		Category:     categoryName,
	}, nil
}
