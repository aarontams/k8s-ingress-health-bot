package bot

import (
	"fmt"
	"github.com/aarontams/k8s-ingress-health-bot/pkg/constants"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io"
	log "k8s.io/klog/v2"
	"net/http"
)

// startWebServer starts the HTTP server
//   /health path
//     Return status 200 if the web server is up ad running.
//   /namespace-health
//     Return status 200 if all ingress endpoints are healthy.
//     Return status 503 if some monitors are pingable.
//     Return status 404 is nothing to monitor.
//   /metrics path
//     Expose Prometheus type metrics for the health of ingress endpoints.
func startWebServer(port int) {
	// Register custom Prometheus metrics.
	registerHealthMetrics()

	// Start web server.
	log.Infof("Starting webserver on port: %d ...", port)
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc(constants.HttpPathHealth, healthResponse)
	router.HandleFunc(constants.HttpPathNamespaceHealth, namespaceHealthResponse)
	router.Handle(constants.HttpPathMetrics, promhttp.Handler())
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), router))
}

// healthResponse returns a http handler for the bot health request.
func healthResponse(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Healthy")
}

// namespaceHealthResponse returns a http handler for the namespace health request.
func namespaceHealthResponse(w http.ResponseWriter, r *http.Request) {
	monitors, err := getIngressesHealth()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf("Bot '/health' endpoint is unhealthy - %v", err)
		io.WriteString(w, fmt.Sprintf("Unhealthy\n%v\n", err))
		return
	}
	if len(monitors) == 0 {
		log.Infof("Bot '/health' endpoint is considered as unhealthy. There is nothing to monitor in namespace '%s'.", namespace)
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, fmt.Sprintf("Unhealthy\nThere is nothing to monitor in namespace '%s'.\n", namespace))
		return
	}

	monitorsInfo := ""
	hasDownMonitors := false
	for _, monitor := range monitors {
		monitorsInfo += monitor.toString() + "\n"
		if !monitor.Up {
			hasDownMonitors = true
		}
	}

	if hasDownMonitors {
		log.Errorf("Bot '/health' endpoint is unhealthy.")
		w.WriteHeader(http.StatusServiceUnavailable)
		io.WriteString(w, fmt.Sprintf("Unhealthy\nOne or more endpoints are down.\n%s", monitorsInfo))
	} else {
		log.Infof("Bot '/health' endpoint is healthy.")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, fmt.Sprintf("Healthy\nAll enpoints are up.\n%s", monitorsInfo))
	}
}
