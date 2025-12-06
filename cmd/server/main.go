package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jaswantK19/order-matching-engine/internal/engine"
	"github.com/jaswantK19/order-matching-engine/internal/models"
)

func main() {
	eg := engine.NewEngine()
	eg.Start()

	http.HandleFunc("/api/v1/orders", func(w http.ResponseWriter, r *http.Request){
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			var order models.Order
			if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			if order.Quantity <= 0 || (order.Type == models.TypeLimit && order.Price <= 0) {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid params"})
				return
			}

			order.OriginalQuantity = order.Quantity
			
			order.ID = fmt.Sprintf("%d", rand.Int63())
			order.Timestamp = time.Now().UnixNano()

			result := eg.SubmitOrder(&order)

			if result.Error != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": result.Error.Error()})
				return
			}

			resp := map[string]interface{}{
				"order_id": order.ID,
				"status": result.Status,
				"filled_quantity": result.FilledQuantity,
				"remaining_quantity": result.RemainingQuantity,
				"trades": result.Trades,
			}

			if result.Status == "FILLED" {
				w.WriteHeader(http.StatusOK)
			} else if result.Status == "PARTIAL_FILL" {
				w.WriteHeader(http.StatusAccepted) // 202
			} else {
				w.WriteHeader(http.StatusCreated) // 201
			}
			json.NewEncoder(w).Encode(resp)
	})

	http.HandleFunc("/api/v1/orders/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 4 {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}
		orderID := parts[4]

		if r.Method == http.MethodDelete {
			cancelled := eg.CancelOrder(orderID)
			if cancelled == nil {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": "Order not found"})
				return
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "CANCELLED", "order_id": orderID})
			return
		}
		
		if r.Method == http.MethodGet {
			order := eg.GetOrder(orderID)
			if order == nil {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": "Order not found"})
				return
			}
			
			status := "ACCEPTED"
			if order.Cancelled {
				status = "CANCELLED"
			} else if order.Quantity == 0 {
				status = "FILLED"
			} else if order.Quantity < order.OriginalQuantity {
				status = "PARTIAL_FILL"
			}

			resp := map[string]interface{}{
				"order_id":        order.ID,
				"symbol":          order.Symbol,
				"side":            order.Side,
				"type":            order.Type,
				"price":           order.Price,
				"quantity":        order.OriginalQuantity,       
				"filled_quantity": order.OriginalQuantity - order.Quantity,
				"status":          status,
				"timestamp":       order.Timestamp,
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)
			return
		}
	})

	http.HandleFunc("/api/v1/orderbook/", func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 4 {
			http.Error(w, "Invalid URL", http.StatusBadRequest)
			return
		}
		symbol := parts[4]
		
		depth := 10 
		depthStr := r.URL.Query().Get("depth")
		if depthStr != "" {
			if d, err := strconv.Atoi(depthStr); err == nil {
				depth = d
			}
		}

		snapshot := eg.GetOrderBook(symbol, depth)
		snapshot.Timestamp = time.Now().UnixMilli()
		
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(snapshot)
	})

	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		metrics := eg.GetMetrics()
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(metrics)
	})

    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"healthy"}`))
    })

	fmt.Println("Server running on :8080")
	http.ListenAndServe(":8080", nil)
}