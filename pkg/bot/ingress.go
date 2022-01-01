package bot

import (
	"errors"
	"fmt"
	"github.com/aarontams/k8s-ingress-health-bot/pkg/constants"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	log "k8s.io/klog/v2"
	"strings"
)

func getIngressesHealth() ([]Monitor, error) {
	var ingresses []*networkingv1.Ingress
	var err error

	err = retry(retryAttempts, retryInterval, func() error {
		ingresses, err = ingressNamespaceLister.List(labels.SelectorFromSet(map[string]string{}))
		if err != nil {
			log.Warningf("unable to list ingresses - %v", err)
			return err
		}
		if len(ingresses) == 0 {
			return errors.New(fmt.Sprintf("there is no ingress in namespace '%s'.", namespace))
		}
		return nil
	})
	if err != nil {
		return []Monitor{}, err
	}

	urls := []string{}
	for _, ingress := range ingresses {
		log.V(6).Infof("Ingress Name: %s", ingress.Name)

		if _, exists := excludedIngressNamesMap[ingress.Name]; exists {
			log.V(6).Infof("Skip ingress '%s' because it is in the excluded list.", ingress.Name)
			continue
		}

		urls = append(urls, getIngressHealthCheckUrls(ingress.ObjectMeta, ingressHealthCheckUrlsAnnotationKey)...)
	}

	// Starts background processes to check endpoints health.
	monitorCh := make(chan *Monitor, len(urls))
	for _, url := range urls {
		log.V(6).Infof("checking %s ...", url)
		go checkUrlHealth(url, monitorCh)
	}

	// Gather monitors from health check results.
	monitors := make([]Monitor, len(urls))
	for i := 0; i < len(urls); i++ {
		monitor := <-monitorCh
		log.V(6).Infof(monitor.toString())
		monitors[i] = *monitor
	}

	return monitors, nil
}

func getIngressHealthCheckUrls(obj metav1.ObjectMeta, annotationKey string) (endpointHealthCheckUrls []string) {
	annotationValue, exists := obj.Annotations[annotationKey]
	if !exists {
		log.V(6).Infof("Skip. Annotations '%s' is not defined in ingress '%s'.", annotationKey, obj.Name)
		return
	}

	return strings.Split(strings.ReplaceAll(annotationValue, " ", ""), constants.DefaultDelimiter)
}
