# k8s-ingress-health-bot
A K8s Ingress Health Bot is a lightweight application to check the health of qualified ingress endpoints for a given kubernetes namespace.

## What Ingress Endpoints Will Be Monitored?
The Bot will monitor all the ingress endpoint URLs specified in `ingress.endpoint.healthcheck.urls` annotation. 
For example:
```yaml
metadata:
  annotations:
    ingress.endpoint.healthcheck.urls: https://bot1.aaron-apps.com/health,https://bot2.aaron-apps.com/health
```

## How It Works?
The Bot will start a web server in the background.

Users can use one of the http paths to retrieve the health of qualified ingress endpoints in real time or
through the latest populated prometheus metrics.

### /health
Return status 200 if the Bot is running.

### /metrics
Expose the prometheus metrics for all qualified ingress endpoints to be monitored.
Periodically (default every minute), the bot will check the health of each qualify ingress endpoint within a given k8s namespace and
populate the health metrics for each endpoint.

**Prometheus Metrics**
##### endpoint_health - value 0 means endpoint is down, value 1 means endpoint is up.
```text
  endpoint_health{namespace="example",status_code="200",url="https://bot1.aaron-apps.com/health"} 1
```
##### cert_validity_days - number of days before cert expired.
```text
  cert_validity_days{namespace="example",url="https://bot1.aaron-apps.com/-/healthy"} 88
```

### /namespace-health
Return the real time health of the all qualified ingress endpoints within a given k8s namespace.

##### 200 OK
All endpoints are up.

##### 404 StatusNotFound
There is no ingress endpoint to monitor.

##### 500 StatusInternalServerError
Internal error, unable to retrieve ingress endpoints.

##### 503 ServiceUnavailable**
One or more endpoints are down.

## Build
```text
go mod vendor
go install ./cmd/...
```

## Usage (Flags)
```text
usage: k8s-ingress-health-bot [<flags>]

Flags:

  -appName string
        Application name of this bot. (default "k8s-ingress-health-bot")

  -excludedIngressNames string
        A comma delimited ingress list that the bot won't monitor.

  -httpClientTimeOut duration
        Maximum time that the HTTP client will wait when pinging an endpoint. (default 15s)

  -ingressHealthCheckUrlsAnnotationKey string
        Ingress annotation key that the bot will look for. (default "ingress.endpoint.healthcheck.urls")

  -kubeconfig string
        Path to a kubeconfig. Only required if the application is running outside of K8s cluster.

  -namespace string
        K8s namespace that this bot will watch. (default "default")

  -port int
        Web server Port. (default 8088)

  -prometheusMetricsPopulateInterval duration
        Time between populate Prometheus metrics for a monitored endpoints. (default 1m0s)

  -prometheusMetricsResetPeriod duration
        Controls how often Prometheus health metrics cache will be cleared. (default 10m0s)

  -retryAttempts int
        Number of attempts before declaring the endpoint is not accessible. (default 3)

  -retryInterval duration
        Time to wait before pinging an endpoint in case of error. (default 3s)
```
