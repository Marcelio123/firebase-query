package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"example.com/query-firebase/math"
	"example.com/query-firebase/model"
	"example.com/query-firebase/whatsapp"
	_ "github.com/mattn/go-sqlite3"
	"google.golang.org/api/option"

	"cloud.google.com/go/firestore"
)

func handler(w http.ResponseWriter, r *http.Request) {
	mainCtx := context.Background()
	client, err := firestore.NewClient(mainCtx, "lucy-cashier-dev", option.WithCredentialsFile("service-account.json"))
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

    wac, err := whatsapp.Connect()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer wac.Disconnect()

	// Get today's date
	now := time.Now()
	year, month, day := now.Date()
	today := fmt.Sprintf("%02d-%02d-%d", day, month, year)

	iter := client.Collection("branches").Documents(mainCtx)

	// iterate branches
	for {
		msg := ""
		doc, err := iter.Next()
		if err != nil {
			fmt.Println(err)
			break
		}

		var branch model.DataBranch
		if err := doc.DataTo(&branch); err != nil {
			fmt.Println("failed parsing branch")
			panic(err)
		}

		msg += fmt.Sprintf("Untuk cabang: %s\n", branch.Name)
		msg += fmt.Sprintf("Penjualan %s pada tanggal %s telah di jumlahkan di bawah.\n\n", branch.Name, today)
		msg += fmt.Sprintf("Ini dia ringkasannya:\n")

		transaction_summaries_itter := client.Collection("transaction_summaries").
		Where("date", "==", today).
		Where("branch_uuid", "==", branch.UUID).
		Documents(mainCtx)
		branch_payment := make(map[string]float64)
		var total_sales float64
		
		// iterate transaction summaries
		for {
			snap, err := transaction_summaries_itter.Next()
			if err != nil {
				log.Print(err)
				break
			}
			
			var transaction_summaries model.TransactionSummaries
			if err := snap.DataTo(&transaction_summaries); err != nil {
				log.Print("failed parsing transaction summaries")
				panic(err)
			}
			
			for _, payment_group := range transaction_summaries.PaymentTypes {
				for key, payment_name := range payment_group {
					total := payment_name.Total
					branch_payment[key] += total
				}
			}
			
			for key, value := range branch_payment {
				msg += fmt.Sprintf("%s:\n %s\n\n", key, math.FormatCurrency(value))
			}
			total_sales = transaction_summaries.TotalSales
			msg += fmt.Sprintf("Total diskon: %s\n", math.FormatCurrency(transaction_summaries.TotalDiscount))
			msg += fmt.Sprintf("Total transaksi: %s\n", math.FormatCurrency(transaction_summaries.TotalSales))
			
		}
		// Laporan kas
		msg += fmt.Sprintf("\nLaporan Kas\n")
		var total_expanse float64
		employee_shifts_iter := client.Collection("employee_shifts").
			Where("branch_uuid", "==", branch.UUID).
			Where("date", "==", today).
			Documents(mainCtx)
		
		for {
			snap, err := employee_shifts_iter.Next()
			if err != nil {
				fmt.Println(err)
				break
			}
	
			var employee_shifts model.EmployeeShifts
			if err := snap.DataTo(&employee_shifts); err != nil {
				fmt.Println("failed parsing employee shifts")
				panic(err)
			}
			msg += fmt.Sprintf("Kas awal: %s\n", math.FormatCurrency(employee_shifts.StartCash))
			for _, value := range employee_shifts.CashEntries {
				msg += fmt.Sprintf("- %s %s\n", value.Description, math.FormatCurrency(value.Value))
				total_expanse += value.Value
			}
			if employee_shifts.EndCash != nil {
				// EndCash exists in Firestore, and its value is not nil
				// Perform computation
				total_expanse = employee_shifts.StartCash + *employee_shifts.EndCash
				msg += fmt.Sprintf("\nTotal Kas: %s\n", math.FormatCurrency(total_expanse))
			} else {
				// EndCash is absent in Firestore (or its value is nil)
				msg += fmt.Sprintf("EndCash is absent or nil, cannot compute total cash\n")
			}
		}

		var tertampung float64
		activeOrderGroupIter := client.Collection("active_order_groups").
			Where("branch_uuid", "==", branch.UUID).
			Where("deleted_at", "==", nil).
			Documents(mainCtx)

		msg += fmt.Sprintf("\nPesanan tertampung: \n")
		for {
			snap, err := activeOrderGroupIter.Next()
			if err != nil {
				log.Print(err)
				break
			}

			var orderGroup model.ActiveOrderGroup
			if err := snap.DataTo(&orderGroup); err != nil {
				log.Print("failed parsing order group")
				panic(err)
			}

			for _, order := range orderGroup.Orders {
				if orderWasCancelled := order.DeletedAt != nil; orderWasCancelled {
					continue
				}
				total_quantity := order.Quantity - order.RefundedQuantity
				price := 0.0
				// price += item price * (item quantity - item refunded)
				price += order.Item.Price * float64(total_quantity)

				// iterate over modifiers map
				for _, modifier := range order.Modifiers {
					// price += modifier price * (modifier quantity - modifier refunded)
					price += modifier.Price * float64(modifier.Quantity)
				}

				// discount
				total_discount := 0.0
				for _, discount := range order.Discounts {
					// price -= fixed * quantity
					if discount.Fixed != 0 {
						total_discount += float64(discount.Fixed) * float64(total_quantity)
					}
					// price *= (100 - percent) /100
					if discount.Percent != 0 {
						total_discount += price * float64(discount.Percent) / 100
					}
				}
				price -= total_discount

				msg += fmt.Sprintf("-%s: %s\n", orderGroup.UUID, math.FormatCurrency(price))
				tertampung += price
			}
		}
		msg += fmt.Sprintf("\nTotal Pesanan Tertampung %s\n\n", math.FormatCurrency(tertampung))
		grand_total := total_sales - total_expanse
		msg += fmt.Sprintf("Grand total: %s\n", math.FormatCurrency(grand_total))
		msg += fmt.Sprintf("Terima Kasih telah menggunakan Lucy <3\n")

		resp, err := whatsapp.SendMessage(mainCtx, wac, fmt.Sprintf("%s%s", branch.Whatsapp.CountryNumber, branch.Whatsapp.Number), msg)
		if err != nil {
			log.Println(resp.ID, err)
		} else {
			log.Printf("Message sent successfully to branch %s.\n", branch.Name)
		}
	}

}

func main() {
	log.Print("starting server...")

	http.HandleFunc("/", handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
