package model

import "time"

type ActiveOrderGroup struct {
	DeletedAt *time.Time `firestore:"deleted_at"`
	UUID      string     `firstore:"uuid"`
	Orders    map[string]Order    `firestore:"orders"`
}

type Order struct {
	UUID      string `firestore:"uuid"`
	CreatedBy string `firestore:"created_by"`

	Item      OrderItem       			`firestore:"item"`
	Modifiers map[string]OrderModifier	`firestore:"modifiers"`
	Discounts map[string]OrderDiscount	`firestore:"discounts"`

	Quantity         int          `firestore:"quantity"`
	RefundedQuantity int          `firestore:"refunded_quantity"`
	Note             string       `firestore:"note"`
	CancelReason     string       `firestore:"cancel_reason,omitempty"`
	Waiter           *OrderWaiter `firestore:"waiter"`

	CreatedAt time.Time  `firestore:"created_at"`
	UpdatedAt *time.Time `firestore:"updated_at"`
	DeletedAt *time.Time `firestore:"deleted_at"`
}

type OrderItem struct {
	UUID         string        `firestore:"uuid"`
	CategoryUUID string        `firestore:"category_uuid"`
	Name         string        `firestore:"name"`
	CategoryName string        `firestore:"category_name"`
	Label        string        `firestore:"label"`
	Description  string        `firestore:"description"`
	ImagePath    *string       `firestore:"image_path"`
	Price        float64       `firestore:"price"`
	Variant      *OrderVariant `firestore:"variant"`
}

type OrderDiscount struct {
	UUID    string  `firestore:"uuid"`
	Name    string  `firestore:"name"`
	Fixed   float64 `firestore:"fixed"`
	Percent float32 `firestore:"percent"`
}

type OrderVariant struct {
	UUID        string  `firestore:"uuid"`
	Label       string  `firestore:"label"`
	ImagePath   string  `firestore:"image_path,omitempty"`
	Description string  `firestore:"description"`
	Price       float64 `firestore:"price"`
}

type OrderModifier struct {
	UUID             string  `firestore:"uuid"`
	Name             string  `firestore:"name"`
	Quantity         int     `firestore:"quantity"`
	RefundedQuantity int     `firestore:"refunded_quantity"`
	Price            float64 `firestore:"price"`
}

// TODO: Check this data structure exists
type OrderWaiter struct {
	UUID string `firestore:"uuid"`
	Name string `firestore:"name"`
}
