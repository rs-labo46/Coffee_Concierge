package usecase

import "time"

// 現在時刻を返すClock実装。
type RealClock struct{}

// RealClockを作る。
func NewRealClock() *RealClock {
	return &RealClock{}
}

// 現在時刻を返す。
func (c *RealClock) Now() time.Time {
	return time.Now()
}
