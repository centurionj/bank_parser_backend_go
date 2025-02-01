package utils

import (
	"math/rand"
	"time"
)

func RandomDuration(min, max int) time.Duration {
	return time.Duration(rand.Intn(max-min+1)+min) * time.Second
}
