package main

import (
	"flag"
	"github.com/aarontams/k8s-ingress-health-bot/pkg/bot"
	"github.com/aarontams/k8s-ingress-health-bot/pkg/constants"
	log "k8s.io/klog/v2"
	"time"
)

var (
	// Global variables.
	appName                             string
	namespace                           string
	httpClientTimeOut                   time.Duration
	retryInterval                       time.Duration
	retryAttempts                       int
	prometheusMetricsPopulateInterval   time.Duration
	prometheusMetricsResetPeriod        time.Duration
	excludedIngressNames                string
	ingressHealthCheckUrlsAnnotationKey string

	// Controller related variables.
	kubeconfig string
	port       int
)

func main() {
	log.InitFlags(nil)
	defer log.Flush()
	flag.Parse()
	printConfig()

	log.Infof("Creating new Bot to watch ingresses in namespace '%s'...", namespace)

	// Create and run Bot.
	botController := bot.NewController(appName, namespace, kubeconfig, port, httpClientTimeOut, retryInterval, retryAttempts,
		prometheusMetricsPopulateInterval, prometheusMetricsResetPeriod, excludedIngressNames, ingressHealthCheckUrlsAnnotationKey)
	botController.Run()
	log.Fatalln("Bot stopped unexpectedly!")
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if the application is running outside of K8s cluster.")
	flag.StringVar(&appName, "appName", constants.DefaultAppName, "Application name of this bot.")
	flag.StringVar(&namespace, "namespace", constants.DefaultNamespace, "K8s namespace that this bot will watch.")
	flag.IntVar(&port, "port", constants.DefaultPort, "Web server Port.")
	flag.DurationVar(&httpClientTimeOut, "httpClientTimeOut", constants.DefaultHttpClientTimeOut, "Maximum time that the HTTP client will wait when pinging an endpoint.")
	flag.DurationVar(&retryInterval, "retryInterval", constants.DefaultRetryInterval, "Time to wait before pinging an endpoint in case of error.")
	flag.IntVar(&retryAttempts, "retryAttempts", constants.DefaultRetryAttempts, "Number of attempts before declaring the endpoint is not accessible.")
	flag.DurationVar(&prometheusMetricsPopulateInterval, "prometheusMetricsPopulateInterval", constants.DefaultPrometheusMetricsPopulateInterval, "Time between populate Prometheus metrics for a monitored endpoints.")
	flag.DurationVar(&prometheusMetricsResetPeriod, "prometheusMetricsResetPeriod", constants.DefaultPrometheusMetricsResetPeriod, "Controls how often Prometheus health metrics cache will be cleared.")
	flag.StringVar(&excludedIngressNames, "excludedIngressNames", "", "A comma delimited ingress list that the bot won't monitor.")
	flag.StringVar(&ingressHealthCheckUrlsAnnotationKey, "ingressHealthCheckUrlsAnnotationKey", constants.IngressHealthCheckUrlsAnnotationKey, "Ingress annotation key that the bot will look for.")

	flag.Usage = func() {
		flag.PrintDefaults()
	}
}

func printConfig() {
	log.Infof("%s will be run with following setting: ", appName)
	log.Infof("  Namespace: %s", namespace)
	log.Infof("  Port: %d", port)
	log.Infof("  HttpClientTimeOut: %v", httpClientTimeOut)
	log.Infof("  RetryIntervals: %v", retryInterval)
	log.Infof("  RetryAttempts: %d", retryAttempts)
	log.Infof("  PrometheusMetricsPopulateInterval: %v", prometheusMetricsPopulateInterval)
	log.Infof("  PrometheusMetricsResetPeriod: %v", prometheusMetricsResetPeriod)
	log.Infof("  ExcludedIngressNames: %s", excludedIngressNames)
	log.Infof("  Expected ingress annotation key: %s", ingressHealthCheckUrlsAnnotationKey)
}
