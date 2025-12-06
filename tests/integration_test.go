package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jaswantK19/order-matching-engine/internal/api"
	"github.com/jaswantK19/order-matching-engine/internal/models"
)

func TestIntegrationFlow(t *testing.T) {
	// Setup Server
	mux := api.SetupServer()
	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := ts.Client()

	sellOrder := models.Order{
		Symbol:   "AAPL",
		Side:     models.SideSell,
		Type:     models.TypeLimit,
		Price:    15000,
		Quantity: 10,
	}
	sellResp := submitOrder(t, client, ts.URL, sellOrder)
	if sellResp["status"] != "ACCEPTED" && sellResp["status"] != "CREATED" {
	}
	if status, ok := sellResp["status"].(string); !ok || (status != "ACCEPTED" && status != "CREATED") {
	}

	buyOrder := models.Order{
		Symbol:   "AAPL",
		Side:     models.SideBuy,
		Type:     models.TypeLimit,
		Price:    15000,
		Quantity: 5,
	}
	buyResp := submitOrder(t, client, ts.URL, buyOrder)

	if buyResp["status"] != "FILLED" {
		t.Errorf("Expected Buy Order to be FILLED, got %v", buyResp["status"])
	}

	trades := buyResp["trades"].([]interface{})
	if len(trades) != 1 {
		t.Errorf("Expected 1 trade, got %d", len(trades))
	}

	resp, err := client.Get(ts.URL + "/api/v1/orderbook/AAPL?depth=10")
	if err != nil {
		t.Fatalf("Failed to get orderbook: %v", err)
	}
	defer resp.Body.Close()

	var obData map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&obData)

	asks := obData["asks"].([]interface{})
	if len(asks) != 1 {
		t.Errorf("Expected 1 ask level, got %d", len(asks))
	}
	askLevel := asks[0].(map[string]interface{})
	qty := askLevel["quantity"].(float64) 
	if qty != 5 {
		t.Errorf("Expected remaining ask quantity 5, got %v", qty)
	}
}

func submitOrder(t *testing.T, client *http.Client, baseURL string, order models.Order) map[string]interface{} {
	body, _ := json.Marshal(order)
	resp, err := client.Post(baseURL+"/api/v1/orders", "application/json", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to submit order: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		t.Errorf("Submit failed with status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	return result
}
