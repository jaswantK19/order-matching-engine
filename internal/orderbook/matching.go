package orderbook

import (
	"fmt"
	"time"

	"github.com/jaswantK19/order-matching-engine/internal/models"
)

type MatchResult struct {
	Trades            []models.Trade
	FilledQuantity    int64
	RemainingQuantity int64
	Status            string
	Error             error
}

func (ob *OrderBook) ProcessOrder(order *models.Order) MatchResult {
	initial_quantity := order.Quantity
	result := MatchResult{
		Trades: []models.Trade{},
	}

	var opponentLevels []*PriceLevel
	if order.Side == models.SideBuy {
		opponentLevels = ob.Asks
	} else {
		opponentLevels = ob.Bids
	}

	if order.Type == models.TypeMarket && len(opponentLevels) == 0 {
		result.Error = fmt.Errorf("insufficient liquidity")
		return result
	}

	for order.Quantity > 0 && len(opponentLevels) > 0 {
		bestLevel := opponentLevels[0]

		if order.Type == models.TypeLimit {
			if order.Side == models.SideBuy && order.Price < bestLevel.Price {
				break
			}
			if order.Side == models.SideSell && order.Price > bestLevel.Price {
				break
			}
		}

		makerOrder := bestLevel.Head

		quantityToTrade := order.Quantity
		if makerOrder.Quantity < quantityToTrade {
			quantityToTrade = makerOrder.Quantity
		}

		trade := models.Trade{
			TradeID: fmt.Sprintf("t-%d", time.Now().UnixNano()),
			Price: bestLevel.Price,
			Quantity: quantityToTrade,
			Timestamp: time.Now().UnixNano(),
			MakerID: makerOrder.ID,
			TakerID: order.ID,
		}
		result.Trades = append(result.Trades, trade)

		order.Quantity -= quantityToTrade
		makerOrder.Quantity -= quantityToTrade
		bestLevel.TotalQuantity -= quantityToTrade

		if makerOrder.Quantity == 0 {
			bestLevel.Remove()
		}

		if bestLevel.Head == nil {
			opponentLevels = opponentLevels[1:]

			if order.Side == models.SideBuy {
				ob.Asks = opponentLevels
			} else {
				ob.Bids = opponentLevels
			}
		}
	}

	result.FilledQuantity = initial_quantity - order.Quantity
	result.RemainingQuantity = order.Quantity

	if order.Quantity > 0 {
		if order.Type == models.TypeLimit {
			ob.addOrderToBook(order)
			result.Status = "ACCEPTED"
		}else{
			result.Status = "PARTIAL_FILL"
		}
	} else {
		result.Status = "FILLED"
	}

	if len(result.Trades) > 0 && result.RemainingQuantity > 0 {
		result.Status = "PARTIAL_FILL"
	}

	return result
}
