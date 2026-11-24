package comp

import "github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"

// ItemStack is some of one item, optionally tied to one person.
type ItemStack struct {
	Item    api.ContentID `json:"item"`
	Count   int           `json:"count"`
	BoundTo api.UID       `json:"bound_to,omitempty"`
}

// Inventory holds item stacks. They are sorted by (Item, BoundTo) and none is empty. Use the inventory
// package's Add, Take and Count rather than changing the slice.
type Inventory struct {
	Stacks []ItemStack `json:"stacks"`
}
