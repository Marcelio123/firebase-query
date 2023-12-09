package main

import (
    "context"
    "fmt"
    "log"
	"net/http"
    "time"
    "strconv"
	"strings"
    "sync"
    "os"

    _ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal"
	"go.mau.fi/whatsmeow"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"

    "cloud.google.com/go/firestore"
    "google.golang.org/api/option"
)

var (
    processedRequests = make(map[string]bool)
    mu               sync.Mutex
)

type DataBranch struct {
    UUID string                     `firestore:"uuid"`
    Name string                     `firestore:"name"`
    Whatsapp map[string]interface{} `firestore:"whatsapp"`
}

type TransactionSummaries struct {
    Date string `firestore:"date"`
    PaymentTypes map[string]map[string]map[string]float64 `firestore:"payment_types"`
    TotalSales float64 `firestore:"total_sales"`
}

type EmployeeShifts struct {
    StartCash   float64                 `firestore:"start_cash"`
    EndCash     *float64                `firestore:"end_cash"`
    StartTime   time.Time               `firestore:"start_time"`
    CashEntries []interface{}           `firestore:"cash_entries"`
}

type activeOrderGroup struct {
    DeletedAt *time.Time `firestore:"deleted_at"`
    UUID string    `firstore:"uuid"`
    Orders map[string]map[string]interface{} `firestore:"orders"`
}

func WAConnect() (*whatsmeow.Client, error) {
	container, err := sqlstore.New("sqlite3", "file:wapp.db?_foreign_keys=on", waLog.Noop)
	if err != nil {
		return nil, err
	}
	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		panic(err)
	}
	client := whatsmeow.NewClient(deviceStore, waLog.Noop)
	if client.Store.ID == nil {
		// No ID stored, new login
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			return nil, err
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}
	} else {
		err := client.Connect()
		if err != nil {
			return nil, err
		}
	}
	return client, nil
}

func sendMessage(ctx context.Context, wac *whatsmeow.Client, user string, msg string) (error){
    _, err := wac.SendMessage(ctx, types.JID{
		User:   "6282269305789",
		Server: types.DefaultUserServer,
	}, &waProto.Message{
			Conversation: proto.String(msg),
	})
	// Check if an error occurred
	if err != nil {
		return err
	} 
    return nil
}

func interfaceToFloat64(value interface{}) float64 {
    // Checking the type and handling it  accordingly
    switch v := value.(type) {
    case int:
        return float64(value.(int))
    case int64:
        return float64(value.(int64))
    case float32:
        return float64(value.(float32))
    case float64:
        return value.(float64)
    default:
        // Handle other types or unknown types
        fmt.Println("Unknown Type or Value:", v)
        return 0
    }
}

func formatCurrency(value float64) string {
	intPart := int(value)
	decimalPart := int((value - float64(intPart)) * 100)
	intString := addCommasToInteger(intPart)
	return fmt.Sprintf("Rp. %s,%02d", intString, decimalPart)
}

func addCommasToInteger(value int) string {
	strValue := strconv.Itoa(value)
	var parts []string
	for i := len(strValue); i > 0; i -= 3 {
		start := i - 3
		if start < 0 {
			start = 0
		}
		parts = append([]string{strValue[start:i]}, parts...)
	}
	return strings.Join(parts, ",")
}

func handler(w http.ResponseWriter, r *http.Request) {
    requestID := r.Header.Get("X-Request-ID") // Assuming a custom header X-Request-ID is sent by the client

    mu.Lock()
    defer mu.Unlock()

    mainCtx := context.Background()
    client, err := firestore.NewClient(mainCtx, "lucy-cashier-dev", option.WithCredentialsFile("service-account.json"))
    if err != nil {
        fmt.Println(err)
        panic(err)
    }
    wac, err := WAConnect()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer wac.Disconnect()

    // Get today's date
	// now := time.Now()
	// year, month, day := now.Date()
    // today := fmt.Sprintf("%d-%02d-%02d", year, month, day)
    today := "08-12-2023"
    
    iter := client.Collection("branches").Documents(mainCtx)
    
    // iterate branches
    for {
        msg := ""
        doc, err := iter.Next()
        if err != nil {
            fmt.Println(err)
            break
        }

        var branch DataBranch
        if err := doc.DataTo(&branch); err != nil {
            fmt.Println("failed parsing branch")
            panic(err)
        }

        // if branch.Whatsapp == nil {
        //     fmt.Println("Error cannot find whatsapp account")
        //     continue
        // }

        msg += fmt.Sprintf("Untuk cabang: %s\n", branch.Name)
        msg += fmt.Sprintf("Penjualan %s pada tanggal %s telah di jumlahkan di bawah.\n", branch.Name, today)
        msg += fmt.Sprintf("Ini dia ringkasannya:\n")

        transaction_summaries_itter := client.Collection("transaction_summaries").
            Where("date", "==", today).
            Where("branch_uuid", "==", branch.UUID).
            Documents(mainCtx)
        branch_payment := make(map[string]float64)
        // iterate transaction summaries
        for {
            snap, err := transaction_summaries_itter.Next()
            if err != nil {
                log.Print(err)
                break
            }

            var transaction_summaries TransactionSummaries
            if err := snap.DataTo(&transaction_summaries); err != nil {
                log.Print("failed parsing transaction summaries")
                panic(err)
            }

            for _, payment := range transaction_summaries.PaymentTypes {
                for key, value := range payment {
                    total := value["total"]
                    branch_payment[key] += total
                }
            }

            for key, value := range branch_payment {
                msg += fmt.Sprintf("%s: %s\n", key, formatCurrency(value))
            }
            msg += fmt.Sprintf("Total penjualan: %s\n", formatCurrency(transaction_summaries.TotalSales))
        }

        //Laporan kas
        // msg += fmt.Sprintf("Laporan Kas\n")
        // employee_shifts_iter := client.Collection("employee_shifts").
        //     Where("branch_uuid", "==", branch.UUID).
        //     Where("date", "==", today).
        //     Documents(mainCtx)
        // for {
        //     snap, err := employee_shifts_iter.Next()
        //     if err != nil {
        //         fmt.Println(err)
        //         break
        //     }

        //     var employee_shifts EmployeeShifts
        //     if err := snap.DataTo(&employee_shifts); err != nil {
        //         fmt.Println("failed parsing employee shifts")
        //         panic(err)
        //     }
        //     msg += fmt.Sprintf("Kas awal: %s", formatCurrency(employee_shifts.StartCash))
        //     var total_expanse float64
        //     for _, value := range employee_shifts.CashEntries {
        //         cash_entry := value.(map[string]interface{})
        //         msg += fmt.Sprintf("- %s %s", cash_entry["description"], formatCurrency(cash_entry["value"].(float64)))
        //         total_expanse += cash_entry["value"].(float64)
        //     }
        //     if employee_shifts.EndCash != nil {
        //         // EndCash exists in Firestore, and its value is not nil
        //         // Perform computation
        //         total_expanse = employee_shifts.StartCash + *employee_shifts.EndCash
        //         msg += fmt.Sprintf("Total Kas: %s", formatCurrency(total_expanse))
        //     } else {
        //         // EndCash is absent in Firestore (or its value is nil)
        //         msg += fmt.Sprintf("EndCash is absent or nil, cannot compute total cash\n")
        //     }

        // }

        var tertampung float64
        activeOrderGroupIter := client.Collection("active_order_groups").
            Where("branch_uuid", "==", branch.UUID).
            Where("deleted_at", "==", nil).
            Documents(mainCtx)
        
        msg += fmt.Sprintf("Pesanan tertampung: \n")
        for {
            snap, err := activeOrderGroupIter.Next()
            if err != nil {
                log.Print(err)
                break
            }

            var orderGroup activeOrderGroup
            if err := snap.DataTo(&orderGroup); err != nil {
                log.Print("failed parsing order group")
                panic(err)
            }

            for _, value := range orderGroup.Orders {
                if _, ok := value["cancel_reason"]; ok {
                    continue
                }
                // to calculate the price of the item considering the quantity, modifier, and discount
                var price float64
                item := value["item"].(map[string]interface{})
                modifiers := value["modifiers"].(map[string]interface{})
                discounts := value["discounts"].(map[string]interface{})
                total_quantity := value["quantity"].(int64) - value["refunded_quantity"].(int64)
                // price += item price * (item quantity - item refunded)
                price += interfaceToFloat64(item["price"]) * interfaceToFloat64(total_quantity)
                // iterate over modifiers map
                for _, modifier := range modifiers {
                    // price += modifier price * (modifier quantity - modifier refunded)
                    modifier_map := modifier.(map[string]interface{})
                    price += interfaceToFloat64(modifier_map["price"]) * interfaceToFloat64(modifier_map["quantity"].(int64) - modifier_map["refunded_quantity"].(int64))
                }
                // discount
                for _, discount := range discounts {
                    // price -= fixed * quantity
                    discount_map := discount.(map[string]interface{})
                    if fixed, ok := discount_map["fixed"]; ok {
                        price -= interfaceToFloat64(fixed) * interfaceToFloat64(total_quantity)
                    }
                    // price *= (100 - percent) /100
                    if percent, ok := discount_map["percent"]; ok {
                        price *= (100.0 - interfaceToFloat64(percent))/100.0
                    }
                }
                msg += fmt.Sprintf("%s %s\n", orderGroup.UUID, formatCurrency(price))
                tertampung += price
            }
        }
        msg += fmt.Sprintf("Total Pesanan Tertampung\n %s",formatCurrency(tertampung))
        
        //err = sendMessage(mainCtx, wac, fmt.Sprintf("%s%s", branch.Whatsapp["country_code"], branch.Whatsapp["number"]), msg)
        err = sendMessage(mainCtx, wac, fmt.Sprintf("%s%s", "62", "82269305789"), msg)
        if err != nil {
            log.Print(err)
        } else {
            log.Printf("Message sent:\n%s", msg)
        }
    }
    
    processedRequests[requestID] = true
	
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