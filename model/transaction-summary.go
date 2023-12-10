package model

type TransactionSummaries struct {
	Date         string                                   `firestore:"date"`
	PaymentTypes map[string]map[string]map[string]float64 `firestore:"payment_types"`
	TotalSales   float64                                  `firestore:"total_sales"`
}
