package orderbook

import (
	"sort"

	"github.com/jaswantK19/order-matching-engine/internal/models"
)

type PriceLevel struct {
	Price         int64         `json:"price"`
	Head          *models.Order `json:"-"`
	Tail          *models.Order `json:"-"`
	TotalQuantity int64         `json:"quantity"`
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

type OrderBookData struct {
	Symbol    string       `json:"symbol"`
	Timestamp int64        `json:"timestamp"`
	Bids      []PriceLevel `json:"bids"`
	Asks      []PriceLevel `json:"asks"`
}

type OrderBook struct {
	Symbol string        `json:"symbol"`
	Bids   []*PriceLevel `json:"bids"`
	Asks   []*PriceLevel `json:"asks"`

	Orders map[string]*models.Order `json:"-"`
}

func NewOrderBook(symbol string) *OrderBook {
	return &OrderBook{
		Symbol: symbol,
		Bids:   make([]*PriceLevel, 0, 100),
		Asks:   make([]*PriceLevel, 0, 100),
		Orders: make(map[string]*models.Order),
	}
}

func (ob *OrderBook) CancelOrder(orderID string) *models.Order {
	order, exists := ob.Orders[orderID]
	if !exists {
		return nil
	}
	order.Cancelled = true
	delete(ob.Orders, orderID)
	return order
}

func (ob *OrderBook) addOrderToBook(order *models.Order) {
	ob.Orders[order.ID] = order
	if order.Side == models.SideBuy {
		ob.insertBid(order)
	} else {
		ob.insertAsk(order)
	}
}

func (ob *OrderBook) insertBid(order *models.Order) {
	i := sort.Search(len(ob.Bids), func(i int) bool {
		return ob.Bids[i].Price <= order.Price
	})

	if i < len(ob.Bids) && ob.Bids[i].Price == order.Price {
		ob.Bids[i].Append(order)
	} else {
		ob.createBidLevel(i, order)
	}
}

func (ob *OrderBook) createBidLevel(index int, order *models.Order) {
	newLevel := &PriceLevel{Price: order.Price}
	newLevel.Append(order)
	ob.Bids = append(ob.Bids, nil)
	copy(ob.Bids[index+1:], ob.Bids[index:])
	ob.Bids[index] = newLevel
}

func (ob *OrderBook) insertAsk(order *models.Order) {
	i := sort.Search(len(ob.Asks), func(i int) bool {
		return ob.Asks[i].Price >= order.Price
	})

	if i < len(ob.Asks) && ob.Asks[i].Price == order.Price {
		ob.Asks[i].Append(order)
	} else {
		ob.createAskLevel(i, order)
	}
}

func (ob *OrderBook) createAskLevel(index int, order *models.Order) {
	newLevel := &PriceLevel{Price: order.Price}
	newLevel.Append(order)
	ob.Asks = append(ob.Asks, nil)
	copy(ob.Asks[index+1:], ob.Asks[index:])
	ob.Asks[index] = newLevel
}

func (ob *OrderBook) GetSnapshot(depth int) OrderBookData {
	snapshot := OrderBookData{
		Symbol: ob.Symbol,
		Bids:   make([]PriceLevel, 0),
		Asks:   make([]PriceLevel, 0),
	}

	bidsLen := len(ob.Bids)
	if depth > 0 && depth < bidsLen {
		bidsLen = depth
	}

	asksLen := len(ob.Asks)
	if depth > 0 && depth < asksLen {
		asksLen = depth
	}

	for i := 0; i < bidsLen; i++ {
		level := ob.Bids[i]
		if level.TotalQuantity > 0 {
			snapshot.Bids = append(snapshot.Bids, *level)
		}
	}
	for i := 0; i < asksLen; i++ {
		level := ob.Asks[i]
		if level.TotalQuantity > 0 {
			snapshot.Asks = append(snapshot.Asks, *level)
		}
	}
	return snapshot
}
