package metrics

import (
	"net/http"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	SmsProviderErrors             *prometheus.CounterVec
	SmsProviderResponseTimeHistogram    *prometheus.HistogramVec
}

func InitMetrics() *Metrics {

	m := &Metrics{
		SmsProviderErrors: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "katana",
				Name:      "ceph_uploaded",
				Help:      "Number of files uploaded by katana in minio",
			}, []string{"status", "channel"},
		),
		SmsProviderResponseTimeHistogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "katana",
				Name:      "katana_http_response_time_index",
				Help:      "Histogram of response times for index HTTP requests.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"status_code", "channel"},
		),

	}
	prometheus.MustRegister(m.SmsProviderErrors)
	prometheus.MustRegister(m.SmsProviderResponseTimeHistogram)

	return m
}

func StartMetricsServer(addr string) {
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe("0.0.0.0:"+addr, nil); err != nil {
		panic(err)
	}
}
