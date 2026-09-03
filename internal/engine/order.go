package engine

import (
	"errors"
	"sync"
	"time"
)

// allows only these 4 states as string and not an arbitrary string like "umbrella"
type OrderStatus string

const (
	StatusOpen      OrderStatus = "OPEN"
	StatusFilled    OrderStatus = "FILLED"
	StatusPartial   OrderStatus = "PARTIAL"
	StatusCancelled OrderStatus = "CANCELLED"
)

// blueprint of an order
type Order struct {
	mu sync.RWMutex

	ID             string
	UserID         string
	Side           string
	Asset          string
	Quantity       int64
	Price          int64
	FilledQuantity int64
	Status         OrderStatus
	CreatedAt      time.Time
}

// func for changing the quantity if buying then selling means filling  and vice versa
func (o *Order) Fill(tradeQty int64) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.Status == StatusFilled || o.Status == StatusCancelled {
		return errors.New("cannot fill a closed order!❌")
	}

	if o.FilledQuantity+tradeQty > o.Quantity {
		return errors.New("Original quantity limit exceeded!")
	}

	if tradeQty <= 0 {
		return errors.New("Trade Quantity must be POSITIVE")
	}

	o.FilledQuantity += tradeQty

	if o.FilledQuantity == o.Quantity {
		o.Status = StatusFilled
	} else {
		o.Status = StatusPartial
	}

	return nil
}

func (o *Order) Cancel() error {

	o.mu.RLock()
	defer o.mu.RUnlock()

	if o.Status == StatusFilled {
		return errors.New("Order already filled. Cannot cancel!")
	}
	if o.Status == StatusCancelled {
		return errors.New("Order already cancelled!")
	}

	return nil
}
func (o *Order) GetStatus() OrderStatus {

	o.mu.RLock()
	defer o.mu.RUnlock()

	return o.Status
}
func (o *Order) GetRemainingQuantity() int64 {

	o.mu.RLock()
	defer o.mu.RUnlock()

	return o.Quantity - o.FilledQuantity
}
func (o *Order) GetFilledQuantity() int64 {

	o.mu.RLock()
	defer o.mu.RUnlock()

	return o.FilledQuantity
}
