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

// Source is a stock of one item that refills each phase, such as a well. It is taken from like a
// container's inventory, but its stock is kept here, apart from any Inventory the entity has.
type Source struct {
	Item  api.ContentID `json:"item"`
	Stock int           `json:"stock"`
	Max   int           `json:"max"`
	// Refill is the template ("template." ID) whose source.refill formula sets the rate.
	Refill api.ContentID `json:"refill"`
}
