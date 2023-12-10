package model

import "time"

type EmployeeShifts struct {
	StartCash   float64     `firestore:"start_cash"`
	EndCash     *float64    `firestore:"end_cash"`
	StartTime   time.Time   `firestore:"start_time"`
	CashEntries []CashEntry `firestore:"cash_entries"`
}

type CashEntry struct {
	Description string  `firestore:"description"`
	Value       float64 `firestore:"value"`
	Expense     bool    `firestore:"expense"`
}
