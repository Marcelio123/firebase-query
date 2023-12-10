package model

import "time"

type ActiveOrderGroup struct {
	DeletedAt *time.Time `firestore:"deleted_at"`
	UUID      string     `firstore:"uuid"`
	Orders    []Order    `firestore:"orders"`
}

type Order struct {
	UUID      string `firestore:"uuid"`
	CreatedBy string `firestore:"created_by"`

	Item      OrderItem       `firestore:"item"`
	Modifiers []OrderModifier `firestore:"modifiers"`
	Discounts []OrderDiscount `firestore:"discounts"`

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
	UUID    string  `bson:"uuid" json:"uuid"`
	Name    string  `bson:"name" json:"name"`
	Fixed   float64 `bson:"fixed" json:"fixed"`
	Percent float32 `bson:"percent" json:"percent"`
}

type OrderVariant struct {
	UUID        string  `bson:"uuid" json:"uuid"`
	Label       string  `bson:"label" json:"label"`
	ImagePath   string  `bson:"image_path,omitempty" json:"image_path,omitempty"`
	Description string  `bson:"description" json:"description"`
	Price       float64 `bson:"price" json:"price"`
}

type OrderModifier struct {
	UUID             string  `bson:"uuid" json:"uuid"`
	Name             string  `bson:"name" json:"name"`
	Quantity         int     `bson:"quantity" json:"quantity"`
	RefundedQuantity int     `bson:"refunded_quantity" json:"refunded_quantity"`
	Price            float64 `bson:"price" json:"price"`
}

// TODO: Check this data structure exists
type OrderWaiter struct {
	UUID string `bson:"uuid" json:"uuid"`
	Name string `bson:"name" json:"name"`
}
