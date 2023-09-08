package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"handler", "server_url"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests",
			Buckets: []float64{0.0001, 0.1, 0.5, 1, 2, 5},
		},
		[]string{"URL", "method", "status"},
	)

	memoryAlloc = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "memory_alloc_bytes",
			Help: "Current memory allocation in bytes",
		},
	)
	processResidentMemory = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "process_resident_memory_bytes_server",
			Help: "Resident memory size of the process in bytes",
		},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(memoryAlloc)
	prometheus.MustRegister(requestDuration)
	prometheus.MustRegister(processResidentMemory)
}

func server1Handler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	log.Println(start)

	duration := float64(time.Since(start).Seconds())

	log.Println(float64(duration))

	httpRequestsTotal.WithLabelValues("server1", "http://127.0.0.1:8010").Inc()

	requestDuration.WithLabelValues("http://127.0.0.1:8010", r.Method, http.StatusText(http.StatusOK)).Observe(float64(duration))

	fmt.Fprintln(w, "Welcome to server 1")
}

func updateProcessResidentMemory() {
	for {
		stats := runtime.MemStats{}
		runtime.ReadMemStats(&stats)

		processResidentMemory.Set(float64(stats.Sys))

		time.Sleep(5 * time.Second)
	}
}

func memoryUsageHandler(w http.ResponseWriter, r *http.Request) {
	httpRequestsTotal.WithLabelValues("memory", "http://127.0.0.1:8010").Inc()

	stats := runtime.MemStats{}
	runtime.ReadMemStats(&stats)
	memoryAlloc.Set(float64(stats.Alloc))
	fmt.Fprintf(w, "Alloc: %d bytes\n", stats.Alloc)
	fmt.Fprintf(w, "TotalAlloc: %d bytes\n", stats.TotalAlloc)
	fmt.Fprintf(w, "Sys: %d bytes\n", stats.Sys)
	fmt.Fprintf(w, "Mallocs: %d\n", stats.Mallocs)
	fmt.Fprintf(w, "Frees: %d\n", stats.Frees)
}

func main() {
	http.HandleFunc("/", server1Handler)
	go updateProcessResidentMemory()

	http.HandleFunc("/memory", memoryUsageHandler)

	http.Handle("/metrics", promhttp.Handler())

	log.Println("Server 1 is running...")
	log.Fatal(http.ListenAndServe(":8010", nil))
}
