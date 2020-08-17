package db

import (
	"time"
)

// Point represents value for given time for a sport
type Point struct {
	Time  time.Time `gorm:"primary_key"`
	Value float32
}
