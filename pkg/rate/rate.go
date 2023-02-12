package rate

import (
	"time"
	"sync"
)

type Limiter struct {
	cost int
	store int
	bucketSize int
	duration time.Duration
	throttleLimit int

	age time.Time
	mutex sync.RWMutex
}

func NewLimiter(cost int, bucketSize int, dur time.Duration, throttle int) *Limiter {
	return &Limiter{
		cost: cost,
		store: 0,
		bucketSize: bucketSize,
		throttleLimit: throttle,
		duration: dur,

		age: time.Now(),
		mutex: sync.RWMutex{},
	}
}

func (limter *Limiter) IsAllowed() bool {
	limter.mutex.Lock()
	defer limter.mutex.Unlock()

	if limter.store >= limter.bucketSize {
		return false
	}

	limter.store = limter.store + limter.cost
	return true
}

func (limter *Limiter) CheckTime() {
	if time.Since(limter.age) > limter.duration * time.Minute {
		go limter.Reset()
	}
}

func (limter *Limiter) IsThrottled() bool {
	limter.mutex.RLock()
	defer limter.mutex.RUnlock()

	if limter.store < limter.bucketSize && limter.store >= limter.throttleLimit {
		return true
	}
	return false
}

func (limter *Limiter) Reset() {
	limter.mutex.Lock()
	defer limter.mutex.Unlock()

	limter.store = 0
	limter.age = time.Now()
}

func (limter *Limiter) GetStore() int {
	return limter.store
}

func (limter *Limiter) GetBucketSize() int {
	return limter.bucketSize
}
