package models

type InventoryItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	SKU          string `json:"sku"`
	CurrentStock int    `json:"currentStock"`
	ImageURL     string `json:"imageUrl"`
	Category     string `json:"category"`
}
