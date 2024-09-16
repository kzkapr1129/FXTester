package common

import (
	"math"
	"time"
)

type Candle struct {
	Time  time.Time
	High  float64
	Open  float64
	Close float64
	Low   float64
}

func (c *Candle) BoxMax() float64 {
	return math.Max(c.Open, c.Close)
}

func (c *Candle) BoxMin() float64 {
	return math.Min(c.Open, c.Close)
}

func (c *Candle) Contains(t *Candle) bool {
	return c.BoxMin() <= t.BoxMin() && t.BoxMax() <= c.BoxMax()
}

func (c *Candle) IsUpdatedBoxMaxBy(t *Candle) bool {
	return c.BoxMax() < t.BoxMax()
}

func (c *Candle) IsUpdatedBoxMinBy(t *Candle) bool {
	return t.BoxMin() < c.BoxMin()
}

func (c *Candle) IsPositive() bool {
	return c.Open < c.Close
}

func (c *Candle) IsNegative() bool {
	return c.Close < c.Open
}
