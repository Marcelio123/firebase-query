package model

type TransactionSummaries struct {
	Date         string             					`firestore:"date"`
	PaymentTypes map[string]map[string]PaymentSummary	`firestore:"payment_types"`
	TotalSales   float64								`firestore:"total_sales"`
}

type PaymentSummary struct {
	Count		float64		`firestore:"count"`
	Total		float64		`firestore:"total"`
}
