package constants

import (
	"time"
)

const (
	// App related.
	DefaultAppName   = "k8s-ingress-health-bot"
	DefaultPort      = 8088
	DefaultNamespace = "default"
	DefaultDelimiter = ","

	HttpPathHealth          = "/health"
	HttpPathNamespaceHealth = "/namespace-health"
	HttpPathMetrics         = "/metrics"

	// K8s realated.
	K8sInformerResyncPeriod = 10 * time.Second

	// Timeout and Retry related.
	DefaultHttpClientTimeOut = 15 * time.Second
	DefaultRetryInterval     = 3 * time.Second
	DefaultRetryAttempts     = 3

	// Prometheus metrics related.
	DefaultPrometheusMetricsPopulateInterval = 60 * time.Second
	DefaultPrometheusMetricsResetPeriod      = 10 * time.Minute

	// Ingress related.
	IngressHealthCheckUrlsAnnotationKey = "ingress.endpoint.healthcheck.urls"
)
