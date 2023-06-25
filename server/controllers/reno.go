package controllers

import (
	"log"
	"os"
	"server/helpers"
	"sync"
	"time"
)

// Congestion controller

type Reno struct {
	mu_ sync.Mutex // guarding struct
	running_tasks_ uint32
	limit_ uint32
	limit_upper_bound_ uint32
	threshold_ uint32
	tasks_before_overload_ uint32
	tasks_before_increase_ uint32
}

func CreateReno() *Reno  {
	reno := new(Reno)
	reno.limit_ = 2
	reno.limit_upper_bound_ = 2
	go reno.LogStats()
	return reno
}


func (c *Reno) slow_start_phase() bool {
	return c.threshold_ == 0 || c.limit_ < c.threshold_
}


func (c *Reno) TryTask() bool {
	c.mu_.Lock()
	defer c.mu_.Unlock()

	if (c.running_tasks_ >= c.limit_) {
		return false
	}
	c.running_tasks_++
	c.limit_upper_bound_ = helpers.Max(2 * c.running_tasks_, c.limit_upper_bound_)
	return true
}


func (c *Reno) done_task() {
	c.running_tasks_--

	if (c.tasks_before_increase_ != 0) {
		c.tasks_before_increase_--
	}

	if (c.tasks_before_overload_ != 0) {
		c.tasks_before_overload_--
	}
}


func (c *Reno) finish_overload() {
	c.mu_.Lock()
	defer c.mu_.Unlock()

	overload_was_detected := c.tasks_before_overload_ != 0

	c.done_task()
	if (!overload_was_detected) {
		c.threshold_ = c.limit_ / 2
		c.limit_ = c.threshold_
		c.limit_upper_bound_ = c.threshold_
		c.tasks_before_overload_ = c.running_tasks_
	}
}


func (c *Reno) finish_normal() {
	c.mu_.Lock()
	defer c.mu_.Unlock()

	if (c.slow_start_phase()) {
		c.done_task()
		c.limit_ = helpers.Min(c.limit_ + 1, c.limit_upper_bound_)
	} else {
		increase_was_req := c.tasks_before_increase_ == 0
		c.done_task()
		if (increase_was_req) {
			c.limit_ = helpers.Min(c.limit_ + 1, c.limit_upper_bound_)
			c.tasks_before_increase_ = c.running_tasks_
		}
	}
}


func (c *Reno) FinishTask(req_time time.Duration) {
	if (req_time <= 700 * time.Millisecond) {
		log.Printf("Finish normal")
		c.finish_normal()
	} else {
		log.Printf("Finish overload")
		c.finish_overload()
	}
}


func (c *Reno) LogStats() {
	stats, _ := os.OpenFile("stats.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	defer stats.Close()
	logger := log.New(stats, "", log.Ltime | log.Lmicroseconds)

	stats.WriteString("date_ts, running_tasks, limit, threshold\n")
	for {
		time.Sleep(200 * time.Millisecond)
		logger.Printf(", %d, %d, %d\n", c.running_tasks_, c.limit_, c.threshold_)
	}
}

// congestion controller ends