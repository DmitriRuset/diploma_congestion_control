package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"server/controllers"
	"sync"
	"time"
)

type first_handler struct {
	mu sync.Mutex // guards execution
	controller  *controllers.Cubic
	log_codes   *log.Logger
	log_timings *log.Logger
}

func step(h *first_handler) {
	h.mu.Lock()
	defer h.mu.Unlock()

	time.Sleep(30 * time.Millisecond)
}

func (h *first_handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !h.controller.TryTask() {
		http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
		h.log_codes.Println("; 429")
		return
	}
	start := time.Now()
	step(h)
	log.Printf("First step completed")
	step(h)
	log.Printf("Second step completed")
	step(h)
	log.Printf("Third step completed")
	end := time.Now()
	h.log_timings.Printf("; %v\n", end.Sub(start).Seconds())

	h.controller.FinishTask(end.Sub(start))
	if end.Sub(start) <= time.Second {
		h.log_codes.Println("; 200")
	} else {
		h.log_codes.Println("; 499")
	}
	fmt.Fprintf(w, "timer elapsed: %v\n", end.Sub(start))
}

func simple_run() {
	f, _ := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer f.Close()

	log.SetOutput(f)

	codes, _ := os.OpenFile("codes.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer codes.Close()
	codes.WriteString("date_ts; code\n")

	timings, _ := os.OpenFile("timings.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	defer timings.Close()
	timings.WriteString("date_ts; timings\n")

	logger_codes := log.New(codes, "", log.Ltime)
	logger_timings := log.New(timings, "", log.Ltime)
	controller := controllers.CreateCubic(0.3, 0.4)

	first_handler := new(first_handler)
	first_handler.controller = controller
	first_handler.log_codes = logger_codes
	first_handler.log_timings = logger_timings
	http.Handle("/connect_first", first_handler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	simple_run()
}
