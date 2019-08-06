package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	statushubGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "status",
		Name:      "probe_statushub",
		Help:      "Shows the component status as reported by StatusHub",
	},
		[]string{"target", "name", "group", "status"},
	)
)

// StatusHubExporter ...
type StatusHubExporter struct {
	URL string
}

// StatusHubData ...
type StatusHubData []struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Group  string `json:"group"`
	Status string `json:"status"`
}

// StatusHubExport ...
func StatusHubExport(registry *prometheus.Registry, target string) {
	start := time.Now()

	// registry := prometheus.NewRegistry()

	registry.MustRegister(statushubGauge)

	status, err := getStatus(target)
	if err != nil {
		return
		// log.Fatalf("OOPS!: %s\n", err)
	}
	for _, value := range status {
		// var statusval float64
		// fmt.Printf("+++ %s\n", value)
		// if value.Status == "up" {
		// 	statusval = 0
		// } else {
		// 	statusval = 1
		// }
		// statushub.WithLabelValues(string(e.URL), value.Name, value.Group, string(value.ID)).Set(statusval)
		statushubGauge.WithLabelValues(string(target), value.Name, value.Group, value.Status).Set(float64(value.ID))
	}

	elapsed := time.Since(start)
	// log.Printf("Query took %s", elapsed)
	probeDurationGauge.WithLabelValues(string(target)).Set(float64(elapsed.Seconds()))
}

func getStatus(target string) (StatusHubData, error) {
	// tr := &http.Transport{
	// 	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	// }
	// client := &http.Client{Transport: tr}
	myURL, _ := formatURL(target)
	httpClient := http.DefaultClient
	httpClient.Timeout = 6 * time.Second
	resp, err := httpClient.Get(myURL)
	if err != nil {
		log.Printf("could not retrieve statushub data: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	var status StatusHubData
	err = json.NewDecoder(resp.Body).Decode(&status)
	if err != nil {
		log.Printf("could not parse statushub json: %v", err)
		return nil, err
	}
	probeSuccessGauge.WithLabelValues(string(target)).Set(float64(resp.StatusCode))
	return status, nil
}

func formatURL(target string) (string, error) {
	// TODO: proper URL formatting
	return fmt.Sprintf("https://%s/?format=json", target), nil
}
