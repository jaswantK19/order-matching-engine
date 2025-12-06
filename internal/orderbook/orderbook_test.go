package orderbook

import (
	"fmt"
	"testing"
	"time"

	"github.com/jaswantK19/order-matching-engine/internal/models"
)

// Helper to create a dummy order
func newOrder(side models.Side, price int64, qty int64) *models.Order {
	return &models.Order{
		ID:        fmt.Sprintf("ord-%d", time.Now().UnixNano()),
		Symbol:    "AAPL",
		Side:      side,
		Type:      models.TypeLimit,
		Price:     price,
		Quantity:  qty,
		Timestamp: time.Now().UnixMilli(),
	}
}

func TestSimpleMatch(t *testing.T) {
	ob := NewOrderBook("AAPL")

	// 1. Add a Sell Order (Maker) @ 150.00
	sell := newOrder(models.SideSell, 15000, 100)
	ob.ProcessOrder(sell)

	// 2. Add a Buy Order (Taker) @ 150.00
	buy := newOrder(models.SideBuy, 15000, 100)
	result := ob.ProcessOrder(buy)

	// Assertions
	if result.Status != "FILLED" {
		t.Errorf("Expected FILLED, got %s", result.Status)
	}
	if len(result.Trades) != 1 {
		t.Errorf("Expected 1 trade, got %d", len(result.Trades))
	}
	if result.Trades[0].Quantity != 100 {
		t.Errorf("Expected trade qty 100, got %d", result.Trades[0].Quantity)
	}
}

// PDF Example: Walking the Book
func TestWalkingTheBook(t *testing.T) {
	ob := NewOrderBook("AAPL")

	// Setup: Sell 10 @ 100, Sell 10 @ 101
	ob.ProcessOrder(newOrder(models.SideSell, 10000, 10))
	ob.ProcessOrder(newOrder(models.SideSell, 10100, 10))

	// Buy 15 @ 102. Should eat all of 100 and half of 101.
	buy := newOrder(models.SideBuy, 10200, 15)
	result := ob.ProcessOrder(buy)

	if len(result.Trades) != 2 {
		t.Errorf("Expected 2 trades, got %d", len(result.Trades))
	}
	if result.FilledQuantity != 15 {
		t.Errorf("Expected 15 filled, got %d", result.FilledQuantity)
	}
}

// PDF Example: FIFO (Time Priority)
func TestTimePriority(t *testing.T) {
	ob := NewOrderBook("AAPL")

	// Seller A (Early)
	s1 := newOrder(models.SideSell, 15000, 100)
	s1.ID = "seller-A"
	ob.ProcessOrder(s1)

	// Seller B (Late)
	s2 := newOrder(models.SideSell, 15000, 100)
	s2.ID = "seller-B"
	ob.ProcessOrder(s2)

	// Buyer takes 100. MUST match Seller A.
	buy := newOrder(models.SideBuy, 15000, 100)
	result := ob.ProcessOrder(buy)

	if result.Trades[0].MakerID != "seller-A" {
		t.Errorf("FIFO Violation: Matched %s instead of seller-A", result.Trades[0].MakerID)
	}
}