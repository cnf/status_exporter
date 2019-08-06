package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	listenAddress = flag.String("web.listen", ":9118", "Address on which to expose metrics and web interface.")

	probeSuccessGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "status",
		Name:      "probe_http_code",
		Help:      "Displays the HTTP response code from the target",
	},
		[]string{"url"},
	)
	probeDurationGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "status",
		Name:      "probe_duration_seconds",
		Help:      "Returns how long the probe took to complete in seconds",
	},
		[]string{"url"},
	)
)

func main() {
	flag.Parse()

	// prometheus.MustRegister(NewExporter(statusURL, "statushub"))
	log.Printf("Starting Server: %s", *listenAddress)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
            <head><title>Status Page Exporter</title></head>
            <body>
            <h1>Status Page Exporter</h1>
            <p><a href="/probe">Run a probe</a></p>
            <p><a href="/metrics">Metrics</a></p>
            </body>
            </html>`))
	})
	http.HandleFunc("/probe", probeHandler)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}

func probeHandler(w http.ResponseWriter, r *http.Request) {

	params := r.URL.Query()
	log.Printf("->>> %s\n", params)
	target := params.Get("target")
	if target == "" {
		http.Error(w, "Target parameter is missing", 400)
		return
	}
	statustype := params.Get("type")
	if statustype == "" {
		http.Error(w, "The type to lookup", 400)
		return
	}

	registry := prometheus.NewRegistry()
	registry.MustRegister(probeSuccessGauge)
	registry.MustRegister(probeDurationGauge)
	if statustype == "statushub" {
		StatusHubExport(registry, target)
	}

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}
