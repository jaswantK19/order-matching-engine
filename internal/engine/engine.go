package engine

import (
	"sync"
	"sync/atomic"

	"time"

	"github.com/jaswantK19/order-matching-engine/internal/models"
	"github.com/jaswantK19/order-matching-engine/internal/orderbook"
)

type CommandType int

const (
	CmdSubmit CommandType = iota
	CmdCancel
	CmdSnapshot
	CmdGetOrder
)

type Command struct {
	Type     CommandType
	Symbol   string
	Order    *models.Order
	OrderID  string
	Depth    int
	RespChan chan interface{}
}

type Metrics struct {
	OrdersReceived  uint64  `json:"orders_received"`
	OrdersMatched   uint64  `json:"orders_matched"`
	OrdersCancelled uint64  `json:"orders_cancelled"`
	TradesExecuted  uint64  `json:"trades_executed"`
	OrdersInBook    int64   `json:"orders_in_book"`
	SystemStartTime int64   `json:"-"`
	Throughput      float64 `json:"throughput_orders_per_sec"`
}

type Engine struct {
	orderbooks map[string]*orderbook.OrderBook
	inputChan  chan Command
	orderMap   sync.Map
	metrics    Metrics
}

func NewEngine() *Engine {
	return &Engine{
		orderbooks: make(map[string]*orderbook.OrderBook),
		inputChan:  make(chan Command, 100000),
		metrics:    Metrics{SystemStartTime: time.Now().Unix()},
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
		if cmd.Type == CmdSubmit {
			ob = orderbook.NewOrderBook(cmd.Symbol)
			e.orderbooks[cmd.Symbol] = ob
		} else {
			cmd.RespChan <- nil
			return
		}
	}

	switch cmd.Type {
	case CmdSubmit:
		e.orderMap.Store(cmd.Order.ID, cmd.Symbol)
		atomic.AddUint64(&e.metrics.OrdersReceived, 1)
		atomic.AddInt64(&e.metrics.OrdersInBook, 1)

		result := ob.ProcessOrder(cmd.Order)

		if len(result.Trades) > 0 {
			atomic.AddUint64(&e.metrics.OrdersMatched, 1)
			atomic.AddUint64(&e.metrics.TradesExecuted, uint64(len(result.Trades)))
		}

		if result.OrdersFilled > 0 {
			atomic.AddInt64(&e.metrics.OrdersInBook, -1*result.OrdersFilled)
		}

		if result.RemainingQuantity == 0 {
			atomic.AddInt64(&e.metrics.OrdersInBook, -1)
		}
		cmd.RespChan <- result

	case CmdCancel:
		cancelledOrder := ob.CancelOrder(cmd.OrderID)
		if cancelledOrder != nil {
			e.orderMap.Delete(cmd.OrderID)
			atomic.AddUint64(&e.metrics.OrdersCancelled, 1)
			atomic.AddInt64(&e.metrics.OrdersInBook, -1)
		}
		cmd.RespChan <- cancelledOrder

	case CmdSnapshot:
		snapshot := ob.GetSnapshot(cmd.Depth)
		cmd.RespChan <- snapshot

	case CmdGetOrder:
		order, found := ob.Orders[cmd.OrderID]
		if !found {
			cmd.RespChan <- nil
		} else {
			orderCopy := *order
			cmd.RespChan <- &orderCopy
		}
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

func (e *Engine) CancelOrder(orderID string) *models.Order {
	val, ok := e.orderMap.Load(orderID)
	if !ok {
		return nil
	}
	symbol := val.(string)

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

func (e *Engine) GetOrder(orderID string) *models.Order {
	val, ok := e.orderMap.Load(orderID)
	if !ok {
		return nil
	}
	symbol := val.(string)

	respChan := make(chan interface{})
	e.inputChan <- Command{
		Type:     CmdGetOrder,
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

func (e *Engine) GetOrderBook(symbol string, depth int) orderbook.OrderBookData {
	respChan := make(chan interface{})
	e.inputChan <- Command{
		Type:     CmdSnapshot,
		Symbol:   symbol,
		Depth:    depth,
		RespChan: respChan,
	}
	return (<-respChan).(orderbook.OrderBookData)
}

func (e *Engine) GetMetrics() Metrics {
	m := Metrics{
		OrdersReceived:  atomic.LoadUint64(&e.metrics.OrdersReceived),
		OrdersMatched:   atomic.LoadUint64(&e.metrics.OrdersMatched),
		OrdersCancelled: atomic.LoadUint64(&e.metrics.OrdersCancelled),
		TradesExecuted:  atomic.LoadUint64(&e.metrics.TradesExecuted),
		OrdersInBook:    atomic.LoadInt64(&e.metrics.OrdersInBook),
	}

	now := time.Now().Unix()
	duration := now - e.metrics.SystemStartTime
	if duration > 0 {
		m.Throughput = float64(m.OrdersMatched) / float64(duration)
	}
	return m
}
