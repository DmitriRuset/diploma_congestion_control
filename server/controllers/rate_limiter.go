package controllers

import (
	"sync"
	"time"
)

type RateLimiter struct {
	mux_ sync.Mutex
	running_tasks_ uint32
	limit_ uint32
}

func CreateRateLimiter(limit uint32) *RateLimiter {
	rl := new(RateLimiter)
	rl.limit_ = limit
	return rl
}

func (r *RateLimiter) TryTask() bool {
	r.mux_.Lock()
	defer r.mux_.Unlock()
	if (r.running_tasks_ >= r.limit_) {
		return false
	}
	r.running_tasks_++
	return true
}


func (r *RateLimiter) FinishTask(req_time time.Duration) {
	r.mux_.Lock()
	defer r.mux_.Unlock()
	r.running_tasks_--
}
