package main

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	netstat "github.com/shirou/gopsutil/v3/net"
)

var (
	// 系统指标
	cpuUsageGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "test_server_cpu_usage_percent",
			Help: "CPU usage percentage",
		},
		[]string{"cpu"},
	)

	memoryUsageGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "test_server_memory_usage_bytes",
			Help: "Memory usage in bytes",
		},
	)

	memoryTotalGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "test_server_memory_total_bytes",
			Help: "Total memory in bytes",
		},
	)

	// 网络指标
	networkBytesSentGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "test_server_network_bytes_sent",
			Help: "Network bytes sent",
		},
		[]string{"interface"},
	)

	networkBytesRecvGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "test_server_network_bytes_recv",
			Help: "Network bytes received",
		},
		[]string{"interface"},
	)

	networkPacketsSentGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "test_server_network_packets_sent",
			Help: "Network packets sent",
		},
		[]string{"interface"},
	)

	networkPacketsRecvGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "test_server_network_packets_recv",
			Help: "Network packets received",
		},
		[]string{"interface"},
	)

	networkErrorsInGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "test_server_network_errors_in",
			Help: "Network errors in",
		},
		[]string{"interface"},
	)

	networkErrorsOutGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "test_server_network_errors_out",
			Help: "Network errors out",
		},
		[]string{"interface"},
	)

	// HTTP 请求指标
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_server_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "test_server_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Go runtime 指标
	goRoutinesGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "test_server_goroutines",
			Help: "Number of goroutines",
		},
	)
)

func init() {
	// 注册所有指标
	prometheus.MustRegister(cpuUsageGauge)
	prometheus.MustRegister(memoryUsageGauge)
	prometheus.MustRegister(memoryTotalGauge)
	prometheus.MustRegister(networkBytesSentGauge)
	prometheus.MustRegister(networkBytesRecvGauge)
	prometheus.MustRegister(networkPacketsSentGauge)
	prometheus.MustRegister(networkPacketsRecvGauge)
	prometheus.MustRegister(networkErrorsInGauge)
	prometheus.MustRegister(networkErrorsOutGauge)
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(goRoutinesGauge)
}

func collectSystemMetrics() {
	for {
		// CPU 使用率
		percentages, err := cpu.Percent(time.Second, true)
		if err == nil {
			for i, percent := range percentages {
				cpuUsageGauge.WithLabelValues(fmt.Sprintf("cpu%d", i)).Set(percent)
			}
		}

		// 内存使用情况
		memInfo, err := mem.VirtualMemory()
		if err == nil {
			memoryUsageGauge.Set(float64(memInfo.Used))
			memoryTotalGauge.Set(float64(memInfo.Total))
		}

		// 网络统计
		netStats, err := netstat.IOCounters(true)
		if err == nil {
			for _, stat := range netStats {
				networkBytesSentGauge.WithLabelValues(stat.Name).Set(float64(stat.BytesSent))
				networkBytesRecvGauge.WithLabelValues(stat.Name).Set(float64(stat.BytesRecv))
				networkPacketsSentGauge.WithLabelValues(stat.Name).Set(float64(stat.PacketsSent))
				networkPacketsRecvGauge.WithLabelValues(stat.Name).Set(float64(stat.PacketsRecv))
				networkErrorsInGauge.WithLabelValues(stat.Name).Set(float64(stat.Errin))
				networkErrorsOutGauge.WithLabelValues(stat.Name).Set(float64(stat.Errout))
			}
		}

		// Go runtime
		goRoutinesGauge.Set(float64(runtime.NumGoroutine()))

		time.Sleep(5 * time.Second)
	}
}

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(ww, r)
		duration := time.Since(start).Seconds()

		httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, fmt.Sprintf("%d", ww.statusCode)).Inc()
		httpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK\n")
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Test Server is running\n")
	fmt.Fprintf(w, "Metrics available at /metrics\n")
	fmt.Fprintf(w, "Health check at /health\n")
}

func main() {
	// 启动指标收集协程
	go collectSystemMetrics()

	// 设置路由
	mux := http.NewServeMux()
	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/health", healthHandler)
	mux.Handle("/metrics", promhttp.Handler())

	// 应用 metrics middleware
	handler := metricsMiddleware(mux)

	port := "8080"
	log.Printf("Starting test server on port %s", port)
	log.Printf("Metrics endpoint: http://localhost:%s/metrics", port)
	log.Printf("Health endpoint: http://localhost:%s/health", port)

	if err := http.ListenAndServe(":"+port, handler); err != nil {
		log.Fatal(err)
	}
}

