package controllers

import (
	"log"
	"math"
	"os"
	"server/helpers"
	"sync"
	"time"
)

type Cubic struct {
	mu_ sync.Mutex // guarding struct
	running_tasks_ uint32
	limit_upper_bound_ uint32
	limit_before_reduction_ uint32
	last_reduction_time_ time.Time
	tasks_before_overload_ uint32
	beta float64
	C float64
}

func CreateCubic(beta float64, C float64) *Cubic  {
	cubic := new(Cubic)
	cubic.limit_upper_bound_ = 2
	cubic.last_reduction_time_ = time.Now()
	cubic.limit_before_reduction_ = 2
	cubic.beta = beta
	cubic.C = C
	go cubic.LogStats()
	return cubic
}

func (c *Cubic) get_limit() uint32 {
	time_since_last_reduction := (time.Since(c.last_reduction_time_)).Seconds()
	K := math.Pow((float64(c.limit_before_reduction_) * (1 - c.beta) / c.C), 1./3)
	float_limit := c.C * math.Pow((time_since_last_reduction - K), 3) + float64(c.limit_before_reduction_)
	limit := helpers.Min(uint32(math.Round(float_limit)), c.limit_upper_bound_)
	return limit
}


func (c *Cubic) TryTask() bool {
	c.mu_.Lock()
	defer c.mu_.Unlock()

	if (c.running_tasks_ >= c.get_limit()) {
		return false
	}
	c.running_tasks_++
	c.limit_upper_bound_ = helpers.Max(2 * c.running_tasks_, c.limit_upper_bound_)
	return true
}

func (c *Cubic) done_task() {
	c.running_tasks_--

	if (c.tasks_before_overload_ != 0) {
		c.tasks_before_overload_--
	}
}

func (c *Cubic) finish_overload() {
	c.mu_.Lock()
	defer c.mu_.Unlock()

	overload_was_detected := c.tasks_before_overload_ != 0

	c.done_task()
	if (!overload_was_detected) {
		c.tasks_before_overload_ = c.running_tasks_
		c.limit_before_reduction_ = c.get_limit()
		c.last_reduction_time_ = time.Now()
		c.limit_upper_bound_ = c.limit_before_reduction_
	}
}


func (c *Cubic) finish_normal() {
	c.mu_.Lock()
	defer c.mu_.Unlock()

	c.done_task()
}


func (c *Cubic) FinishTask(req_time time.Duration) {
	if (req_time <= 700 * time.Millisecond) {
		log.Printf("Finish normal")
		c.finish_normal()
	} else {
		log.Printf("Finish overload")
		c.finish_overload()
	}
}


func (c *Cubic) LogStats() {
	stats, _ := os.OpenFile("stats.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	defer stats.Close()
	logger := log.New(stats, "", log.Ltime | log.Lmicroseconds)

	stats.WriteString("date_ts, running_tasks, limit, limit_upper_bound, limit_before_reduction\n")
	for {
		time.Sleep(100 * time.Millisecond)
		logger.Printf(", %d, %d, %d, %d\n", c.running_tasks_, c.get_limit(), c.limit_upper_bound_, c.limit_before_reduction_)
	}
}