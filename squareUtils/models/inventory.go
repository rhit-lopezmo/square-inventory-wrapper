package models

type InventoryItem struct {
	ID                string `json:"id"`
	Name              string `json:"name"`
	Description       string `json:"description"`
	SKU               string `json:"sku"`
	CurrentStock      int    `json:"currentStock"`
	ImageURL          string `json:"imageUrl"`
	Category          string `json:"category"`
	ReportingCategory string `json:"reportingCategory"`
}

// InventoryItemUpdate represents optional updates for an inventory item.
type InventoryItemUpdate struct {
	ID                *string `json:"id"`
	Name              *string `json:"name"`
	Description       *string `json:"description"`
	CurrentStock      *int    `json:"currentStock"`
	ImageURL          *string `json:"imageUrl"`
	Category          *string `json:"category"`
	ReportingCategory *string `json:"reportingCategory"`
}

// ApplyUpdate merges provided fields onto an InventoryItem without overwriting missing values.
func (i *InventoryItem) ApplyUpdate(update *InventoryItemUpdate) {
	if update == nil {
		return
	}

	if update.ID != nil {
		i.ID = *update.ID
	}

	if update.Name != nil {
		i.Name = *update.Name
	}

	if update.Description != nil {
		i.Description = *update.Description
	}

	if update.CurrentStock != nil {
		i.CurrentStock = *update.CurrentStock
	}

	if update.ImageURL != nil {
		i.ImageURL = *update.ImageURL
	}

	if update.Category != nil {
		i.Category = *update.Category
	}

	if update.ReportingCategory != nil {
		i.ReportingCategory = *update.ReportingCategory
	}
}
