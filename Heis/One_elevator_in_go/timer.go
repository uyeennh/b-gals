package main

import "time"

// Timer wraps time.Timer with safe stop/reset behaviour
// Ensure reset after stop, time.timer is a type in the go-library, where we point timer to the type.
// time.Timer from he Go's library
type Timer struct {
	timer *time.Timer
}

// Afactory that builds you a ready-to-use box with a clock inside
// the clock is stopped, needs to start and reset it
func NewTimer() *Timer {
	tt := time.NewTimer(time.Hour)
	if !tt.Stop() {
		select {
		case <-tt.C:
		default:
		}
	}
	return &Timer{timer: tt}
}

// the beeper
func (tm *Timer) C() <-chan time.Time {
	return tm.timer.C
}

// cancel button, press it and the countdown stops immmediatly.
func (tm *Timer) Stop() {
	if tm.timer == nil {
		return
	}
	if !tm.timer.Stop() {
		select {
		case <-tm.timer.C:
		default:
		}
	}
}

// a set and go button, you tell it how long to count down, and it starts immediately
func (tm *Timer) Start(d time.Duration) {
	if tm.timer == nil {
		t := time.NewTimer(d)
		tm.timer = t
		return
	}
	tm.Stop()
	tm.timer.Reset(d)
}

func (tm *Timer) Reset(d time.Duration) {
	tm.Start(d)
}
