package orderbook

import "github.com/jaswantK19/order-matching-engine/internal/models"

type PriceLevel struct {
	Price         int64
	Head          *models.Order
	Tail          *models.Order
	TotalQuantity int64
}

func (pl *PriceLevel) Append(o *models.Order) {
	o.Next = nil
	if pl.Head == nil {
		pl.Head = o
		pl.Tail = o
	} else {
		pl.Tail.Next = o
		pl.Tail = o
	}
	pl.TotalQuantity += o.Quantity
}

func (pl *PriceLevel) Remove() *models.Order {
	if pl.Head == nil {
		return nil
	}
	o := pl.Head
	pl.Head = o.Next
	if pl.Head == nil {
		pl.Tail = nil
	}
	pl.TotalQuantity -= o.Quantity
	return o
}

type OrderBook struct {
	Symbol string
	Bids   map[int64]*PriceLevel
	Asks   map[int64]*PriceLevel
}

func NewOrderBook(symbol string) *OrderBook {
	return &OrderBook{
		Symbol: symbol,
		Bids:   make(map[int64]*PriceLevel),
		Asks:   make(map[int64]*PriceLevel),
	}
}
