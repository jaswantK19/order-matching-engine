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
	Bids   []*PriceLevel
	Asks   []*PriceLevel

	Orders map[string]*models.Order
}

func NewOrderBook(symbol string) *OrderBook {
	return &OrderBook{
		Symbol: symbol,
		Bids:   make([]*PriceLevel, 0, 100), // Pre-allocating memory helps performance
		Asks:   make([]*PriceLevel, 0, 100),
		Orders: make(map[string]*models.Order), //lookup map
	}
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
	for i, level := range ob.Bids {
		if level.Price == order.Price {
			level.Append(order)
			return
		}
		if level.Price < order.Price {
			ob.createBidLevel(i, order)
			return
		}
	}
}

func (ob *OrderBook) createBidLevel(index int, order *models.Order){
	newLevel := &PriceLevel{Price: order.Price}
	newLevel.Append(order)
	ob.Bids = append(ob.Bids, nil)
	copy(ob.Bids[index+1:], ob.Bids[index:])
	ob.Bids[index] = newLevel
}

func (ob *OrderBook) insertAsk(order *models.Order){
	for i, level := range ob.Asks {
		if level.Price == order.Price {
			level.Append(order)
			return
		}
		if level.Price > order.Price {
			ob.createAskLevel(i, order)
			return
		}
	}
}

func (ob *OrderBook) createAskLevel(index int, order *models.Order) {
	newLevel := &PriceLevel{Price: order.Price}
	newLevel.Append(order)
	ob.Asks = append(ob.Asks, nil)
	copy(ob.Asks[index+1:], ob.Asks[index:])
	ob.Asks[index] = newLevel
}


