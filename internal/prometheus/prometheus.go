package prometheus

import "github.com/prometheus/client_golang/prometheus"

var (
	CacheHit     = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "cache_hit_total"}, []string{"domain"})
	CacheMiss    = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "cache_miss_total"}, []string{"domain"})
	ServeLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "serve_latency", Buckets: prometheus.DefBuckets}, []string{"domain"})
)
