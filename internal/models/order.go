package models

type OrderType string

const (
	TypeLimit  OrderType = "LIMIT"
	TypeMarket OrderType = "MARKET"
)

type Side string

const (
	SideBuy  Side = "BUY"
	SideSell Side = "SELL"
)

type Order struct {
	ID        string    `json:"id"`
	Symbol    string    `json:"symbol"`
	Side      Side      `json:"side"`
	Type      OrderType `json:"type"`
	Price     int64     `json:"price"`
	
	Quantity  int64     `json:"quantity"` 
	OriginalQuantity int64 `json:"-"` 
	
	Timestamp int64     `json:"timestamp"`

	Cancelled bool   `json:"-"`
	Next      *Order `json:"-"`
}

type Trade struct {
	TradeID   string `json:"trade_id"`
	Price     int64  `json:"price"`
	Quantity  int64  `json:"quantity"`
	Timestamp int64  `json:"timestamp"`
	MakerID   string `json:"maker_id"`
	TakerID   string `json:"taker_id"`
}