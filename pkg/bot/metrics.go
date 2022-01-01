package bot

import (
	"github.com/prometheus/client_golang/prometheus"
	log "k8s.io/klog/v2"
	"strconv"
	"strings"
)

const (
	metricNameEndpointHealth   = "endpoint_health"
	metricNameCertValidityDays = "cert_validity_days"

	namespaceLabel  = "namespace"
	urlLabel        = "url"
	statusCodeLable = "status_code"
)

// endpointHealthMetric defines a prometheus metric for each endpoint health.
var endpointHealthMetric = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: appName,
		Name:      metricNameEndpointHealth,
		Help:      "Value 0 means endpoint is down, value 1 means endpoint is up.",
	},
	[]string{namespaceLabel, statusCodeLable, urlLabel},
)

// certValidityDaysMetric defines a prometheus metric for each endpoint cert validity days.
var certValidityDaysMetric = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		Namespace: appName,
		Name:      metricNameCertValidityDays,
		Help:      "Number of days before cert expired.",
	},
	[]string{namespaceLabel, urlLabel},
)

// addHealthMetrics adds prometheus metrics for the endpoint health.
func addHealthMetrics(m Monitor) {
	upValue := float64(0)
	if m.Up {
		upValue = float64(1)
	}

	endpointHealthMetric.With(prometheus.Labels{
		namespaceLabel:  namespace,
		statusCodeLable: strconv.Itoa(m.StatusCode),
		urlLabel:        m.Url}).
		Set(upValue)

	if strings.HasPrefix(m.Url, "https") {
		certValidityDaysMetric.With(prometheus.Labels{
			namespaceLabel: namespace,
			urlLabel:       m.Url}).
			Set(float64(m.CertValidityDays))
	}
}

// registerHealthMetrics registers all endpoint health related prometheus metrics.
func registerHealthMetrics() {
	log.V(6).Info("Registering health related metrics...")
	prometheus.MustRegister(endpointHealthMetric, certValidityDaysMetric)
}

// resetHealthMetrics deletes all endpoint health related prometheus metrics.
func resetHealthMetrics() {
	log.V(6).Info("Resetting health related metrics...")
	endpointHealthMetric.Reset()
	certValidityDaysMetric.Reset()
}
