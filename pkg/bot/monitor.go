package bot

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	log "k8s.io/klog/v2"
	"net/http"
	"strings"
	"time"
)

type Monitor struct {
	Url              string
	Up               bool
	StatusCode       int
	CertValidityDays int
	ResponseTime     time.Duration
}

func (m Monitor) toString() string {
	status := "Down"
	if m.Up {
		status = "Up"
	}
	return fmt.Sprintf("[%s] %s {Status:%d CertValidityDays:%d RespondTime: %v}", status, m.Url, m.StatusCode,
		m.CertValidityDays, m.ResponseTime.Round(time.Millisecond))
}

func checkUrlHealth(url string, monitorCh chan<- *Monitor) {
	monitor := Monitor{Url: url, CertValidityDays: -1, StatusCode: -1}

	err := retry(retryAttempts, retryInterval, func() error {
		// Construct a http client with HTTP keep-alives disabled.
		// The client will only use the connection to the server for a single HTTP request.
		transport := http.Transport{
			DisableKeepAlives: true,
		}
		httpClient := http.Client{
			Transport: &transport,
			Timeout:   httpClientTimeOut,
		}
		client := resty.NewWithClient(&httpClient)
		client.SetHeader("User-Agent", appName)
		client.SetRedirectPolicy(resty.NoRedirectPolicy())

		resp, responseErr := client.R().Execute(http.MethodGet, url)
		if resp != nil {
			monitor.ResponseTime = resp.Time()
			monitor.StatusCode = resp.StatusCode()
			switch resp.StatusCode() {
			case http.StatusOK:
				monitor.Up = true
				log.V(8).Infof("Up. %+v", monitor)
			default:
				monitor.Up = false
				log.Warningf("Down. %s {Status:%d} - %v", monitor.Url, resp.StatusCode(), responseErr)
			}

			// Get the certificate expiration info if TLS is defined.
			// If there is more than one CA cert, only the first one will be considered.
			if strings.HasPrefix(url, "https") && resp.RawResponse != nil && resp.RawResponse.TLS != nil {
				for _, cert := range resp.RawResponse.TLS.PeerCertificates {
					if !cert.IsCA {
						monitor.CertValidityDays = int(cert.NotAfter.Sub(time.Now()).Hours() / 24)
						break
					}
				}
			}
		}
		return responseErr
	})

	if err != nil {
		log.Errorf(monitor.toString())
	} else {
		log.V(8).Infof(monitor.toString())
	}
	monitorCh <- &monitor
}
