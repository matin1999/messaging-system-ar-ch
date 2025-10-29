package metrics

import (
	"net/http"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	FfmpegClientCounter           *prometheus.CounterVec
	StorageWriteSizeHistogram     *prometheus.SummaryVec
	MongoErrorCounter             *prometheus.CounterVec
	KafkaErrorCounter             *prometheus.CounterVec
	CephUploadCounter             *prometheus.CounterVec
	MinioUploadCounter            *prometheus.CounterVec
	PlaylistStatusCodeCount       *prometheus.CounterVec
	IndexStatusCodeCount          *prometheus.CounterVec
	PlaylistResponseTimeHistogram *prometheus.HistogramVec
	IndexResponseTimeHistogram    *prometheus.HistogramVec
}

func InitMetrics() *Metrics {

	m := &Metrics{
		CephUploadCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "katana",
				Name:      "ceph_uploaded",
				Help:      "Number of files uploaded by katana in minio",
			}, []string{"status", "channel"},
		),
		MinioUploadCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "katana",
				Name:      "minio_uploaded",
				Help:      "Number of files uploaded by katana in minio",
			}, []string{"status", "channel"},
		),
		FfmpegClientCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "katana",
				Name:      "ffmpeg_client",
				Help:      "Count of ffmpeg client",
			}, []string{"channel"},
		),
		StorageWriteSizeHistogram: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace: "katana",
				Name:      "storage_file_size_bytes",
				Help:      "Histogram of file sizes uploaded in ceph ",
			}, []string{"storage_type", "channel"},
		),
		MongoErrorCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "katana",
				Name:      "mongo_error",
				Help:      "Number of mongo database error",
			}, []string{"status", "channel"},
		),
		KafkaErrorCounter: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "katana",
				Name:      "kafka_error",
				Help:      "Number of kafka error",
			}, []string{"status", "channel"},
		),
		PlaylistResponseTimeHistogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "katana",
				Name:      "katana_http_response_time_playlist",
				Help:      "Histogram of response times for playlist HTTP requests.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"status_code", "channel"},
		),
		PlaylistStatusCodeCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "katana",
				Name:      "katana_status_codes_playlist",
				Help:      "Number of playlist status code count",
			}, []string{"status_code", "channel"},
		),
		IndexResponseTimeHistogram: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: "katana",
				Name:      "katana_http_response_time_index",
				Help:      "Histogram of response times for index HTTP requests.",
				Buckets:   prometheus.DefBuckets,
			},
			[]string{"status_code", "channel"},
		),
		IndexStatusCodeCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: "katana",
				Name:      "katana_status_codes_index",
				Help:      "Number of index status code count",
			}, []string{"status_code", "channel"},
		),
	}
	prometheus.MustRegister(m.CephUploadCounter)
	prometheus.MustRegister(m.MinioUploadCounter)
	prometheus.MustRegister(m.FfmpegClientCounter)
	prometheus.MustRegister(m.StorageWriteSizeHistogram)
	prometheus.MustRegister(m.MongoErrorCounter)
	prometheus.MustRegister(m.KafkaErrorCounter)
	prometheus.MustRegister(m.PlaylistResponseTimeHistogram)
	prometheus.MustRegister(m.PlaylistStatusCodeCount)
	prometheus.MustRegister(m.IndexResponseTimeHistogram)
	prometheus.MustRegister(m.IndexStatusCodeCount)

	return m
}

func StartMetricsServer(addr string) {
	http.Handle("/metrics", promhttp.Handler())
	if err := http.ListenAndServe("0.0.0.0:"+addr, nil); err != nil {
		panic(err)
	}
}
