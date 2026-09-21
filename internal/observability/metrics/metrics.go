// Package metrics defines the application's Prometheus instruments.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// HTTP instruments.
var (
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ticketflow_http_requests_total",
			Help: "Total HTTP requests by route, method and status.",
		},
		[]string{"route", "method", "status"},
	)
	HTTPDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ticketflow_http_duration_seconds",
			Help:    "HTTP request latency by route.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"route"},
	)
)

// Business instruments.
var (
	OrdersTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ticketflow_orders_total",
			Help: "Orders created by final status.",
		},
		[]string{"status"},
	)
	HoldsActive = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ticketflow_holds_active",
			Help: "Currently active holds.",
		},
	)
	PaymentDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "ticketflow_payment_duration_seconds",
			Help:    "Gateway authorize latency.",
			Buckets: []float64{.05, .1, .25, .5, 1, 2.5, 5},
		},
	)
)
