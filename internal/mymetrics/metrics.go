package mymetrics

import (
	"fmt"
	"httpServer/internal/http1"
	"io"
	"net/http"
	"time"
)

var (
	startTime = time.Now()
)

func Handler() http1.Handler {
	mux := http1.NewMux()
	mux.HandleFunc("/metrics", metricsHandler)
	return mux
}

func metricsHandler(w http1.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	writeMetrics(w)
}

func writeMetrics(w io.Writer) {
	//uptime
	uptimeSeconds := time.Since(startTime).Seconds()
	fmt.Fprintf(w, "# HELP process_uptime_seconds Total time the server has been running\n")
	fmt.Fprintf(w, "# TYPE process_uptime_seconds gauge\n")
	fmt.Fprintf(w, "process_uptime_seconds %.2f\n\n", uptimeSeconds)
}
