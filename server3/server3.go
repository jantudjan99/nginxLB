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

func server3Handler(w http.ResponseWriter, r *http.Request) {

	start := time.Now()

	log.Println(start)

	duration := float64(time.Since(start).Seconds())

	httpRequestsTotal.WithLabelValues("server3", "http://127.0.0.1:8012").Inc()

	requestDuration.WithLabelValues("http://127.0.0.1:8012", r.Method, http.StatusText(http.StatusOK)).Observe(duration)

	fmt.Fprintln(w, "Welcome to server 3")
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
	httpRequestsTotal.WithLabelValues("memory", "http://127.0.0.1:8012").Inc()

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

	http.HandleFunc("/", server3Handler)
	http.HandleFunc("/memory", memoryUsageHandler)

	http.Handle("/metrics", promhttp.Handler())

	log.Println("Server 3 is running and change...")
	log.Fatal(http.ListenAndServe(":8012", nil))
}
