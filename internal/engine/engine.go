package engine

import (
	"github.com/jaswantK19/order-matching-engine/internal/models"
	"github.com/jaswantK19/order-matching-engine/internal/orderbook"
)


type CommandType int
const (
	CmdSubmit CommandType = iota
	CmdCancel
	CmdSnapshot
)


type Command struct {
	Type     CommandType
	Symbol   string
	Order    *models.Order 
	OrderID  string        
	RespChan chan interface{}
}

type Engine struct {
	orderbooks map[string]*orderbook.OrderBook
	inputChan  chan Command
}

func NewEngine() *Engine {
	return &Engine{
		orderbooks: make(map[string]*orderbook.OrderBook),
		inputChan:  make(chan Command, 10000),
	}
}

func (e *Engine) Start() {
	go func() {
		for cmd := range e.inputChan {
			e.process(cmd)
		}
	}()
}

func (e *Engine) process(cmd Command) {

	ob, ok := e.orderbooks[cmd.Symbol]
	if !ok {
		ob = orderbook.NewOrderBook(cmd.Symbol)
		e.orderbooks[cmd.Symbol] = ob
	}


	switch cmd.Type {
	case CmdSubmit:
		result := ob.ProcessOrder(cmd.Order)
		cmd.RespChan <- result

	case CmdCancel:
		cancelledOrder := ob.CancelOrder(cmd.OrderID)
		cmd.RespChan <- cancelledOrder

	case CmdSnapshot:
		snapshot := ob.GetSnapshot()
		cmd.RespChan <- snapshot
	}
}



func (e *Engine) SubmitOrder(order *models.Order) orderbook.MatchResult {
	respChan := make(chan interface{})
	e.inputChan <- Command{
		Type:     CmdSubmit,
		Symbol:   order.Symbol,
		Order:    order,
		RespChan: respChan,
	}
	return (<-respChan).(orderbook.MatchResult)
}

func (e *Engine) CancelOrder(symbol, orderID string) *models.Order {
	respChan := make(chan interface{})
	e.inputChan <- Command{
		Type:     CmdCancel,
		Symbol:   symbol,
		OrderID:  orderID,
		RespChan: respChan,
	}
	result := <-respChan
	if result == nil {
		return nil
	}
	return result.(*models.Order)
}

func (e *Engine) GetOrderBook(symbol string) orderbook.OrderBookData {
	respChan := make(chan interface{})
	e.inputChan <- Command{
		Type:     CmdSnapshot,
		Symbol:   symbol,
		RespChan: respChan,
	}
	return (<-respChan).(orderbook.OrderBookData)
}