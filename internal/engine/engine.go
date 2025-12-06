package engine

import (
	"github.com/jaswantK19/order-matching-engine/internal/models"
	"github.com/jaswantK19/order-matching-engine/internal/orderbook"
)

type OrderRequest struct {
	Order        *models.Order
	ResponseChan chan orderbook.MatchResult
}

type Engine struct {
	orderbooks map[string]*orderbook.OrderBook
	inputChan  chan OrderRequest
}

func NewEngine() *Engine {
	return &Engine{
		orderbooks: make(map[string]*orderbook.OrderBook),
		inputChan:  make(chan OrderRequest, 10000),
	}
}

func (e *Engine) Start() {
	go func() {
		for req := range e.inputChan {
			e.process(req)
		}
	}()
}

func (e *Engine) process(req OrderRequest) {
	ob, ok := e.orderbooks[req.Order.Symbol]
	if !ok {
		ob = orderbook.NewOrderBook(req.Order.Symbol)
		e.orderbooks[req.Order.Symbol] = ob
	}

	result := ob.ProcessOrder(req.Order)
	req.ResponseChan <- result
}