package bot

import (
	"errors"
	"github.com/aarontams/k8s-ingress-health-bot/pkg/constants"
	"github.com/aarontams/k8s-ingress-health-bot/pkg/signals"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	networkinglistersv1 "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	log "k8s.io/klog/v2"
	"strings"
	"time"
)

// Variables which will be used by different bot components.
var (
	appName                             string
	namespace                           string
	httpClientTimeOut                   time.Duration
	retryInterval                       time.Duration
	retryAttempts                       int
	prometheusMetricsPopulateInterval   time.Duration
	prometheusMetricsResetPeriod        time.Duration
	ingressNamespaceLister              networkinglistersv1.IngressNamespaceLister
	excludedIngressNamesMap             map[string]string
	ingressHealthCheckUrlsAnnotationKey string
)

type Controller struct {
	kubeClient      kubernetes.Interface
	port            int
	ingressesSynced cache.InformerSynced
	stopCh          <-chan struct{}
}

//  Use by retry logic.
type stop struct {
	error
}

// NewController returns a new controller.
func NewController(appNameValue string, namespaceValue string, kubeconfig string, port int, httpClientTimeOutValue time.Duration,
	retryIntervalValue time.Duration, retryAttemptsValue int, prometheusMetricsPopulateIntervalValue time.Duration,
	prometheusMetricsResetPeriodValue time.Duration, excludedIngressNames string, ingressHealthCheckUrlsAnnotationKeyValue string) *Controller {
	// Create a new k8s Clientset for the given k8s config.
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		log.Fatalf("unable to creates ClientConfig with the given config: %s", err.Error())
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("unable to create Clientset: %s", err.Error())
	}

	// Init global variables which will be used by different bot components.
	appName = appNameValue
	namespace = namespaceValue
	httpClientTimeOut = httpClientTimeOutValue
	retryInterval = retryIntervalValue
	retryAttempts = retryAttemptsValue
	prometheusMetricsPopulateInterval = prometheusMetricsPopulateIntervalValue
	prometheusMetricsResetPeriod = prometheusMetricsResetPeriodValue
	ingressHealthCheckUrlsAnnotationKey = ingressHealthCheckUrlsAnnotationKeyValue
	for _, ingressName := range strings.Split(excludedIngressNames, ",") {
		excludedIngressNamesMap = make(map[string]string)
		excludedIngressNamesMap[ingressName] = ingressName
	}

	// Construct shared ingress informer.
	var kubeInformerFactory = informers.NewSharedInformerFactoryWithOptions(kubeClient, constants.K8sInformerResyncPeriod,
		informers.WithNamespace(namespace), informers.WithTweakListOptions(nil))
	ingressInformer := kubeInformerFactory.Networking().V1().Ingresses()
	ingressNamespaceLister = ingressInformer.Lister().Ingresses(namespace)

	// Construct a new controller.
	controller := &Controller{
		port:            port,
		kubeClient:      kubeClient,
		ingressesSynced: ingressInformer.Informer().HasSynced,
	}

	// Set up signals for controller graceful shutdown.
	log.Info("Setting up signals for controller graceful shutdown...")
	controller.stopCh = signals.SetupSignalHandler()

	// Start Ingress Informer Factory in the background.
	log.Info("Starting K8s informer factory in the background...")
	go kubeInformerFactory.Start(controller.stopCh)

	return controller
}

// Run will run indefinitely to monitor the health of the ingresses in the given namespace.
func (c *Controller) Run() error {
	log.Infof("Starting controller...")

	// Start web server.
	go startWebServer(c.port)

	// Wait for K8s ingress informer caches to populate.
	log.Info("Waiting for K8s ingress informer caches to populate...")
	if ok := cache.WaitForCacheSync(c.stopCh, c.ingressesSynced); !ok {
		return errors.New("failed to wait for K8s ingress informer caches to populate")
	}

	// Run controller forever.
	log.Infof("Controller is ready.")
	startTime := time.Now()
	for {
		// To ensure there is no leftover Prometheus health metrics, clear the cache periodically.
		if time.Since(startTime) > prometheusMetricsResetPeriod {
			log.Infof("Clear the Prometheus health metrics cache.")
			resetHealthMetrics()
			startTime = time.Now()
		}

		log.Infof("Checking ingress endpoints...")
		monitors, err := getIngressesHealth()
		if err != nil {
			log.Errorf("unable to get ingresses health - %v", err)
			log.Infof("Sleep for %v and try again...", 2*constants.DefaultRetryInterval)
			time.Sleep(2 * constants.DefaultRetryInterval)
			continue
		}

		// Republish all health related metrics.
		resetHealthMetrics()
		for _, monitor := range monitors {
			log.Info(monitor.toString())
			addHealthMetrics(monitor)
		}

		log.Infof("Check completed!")
		time.Sleep(prometheusMetricsPopulateInterval)
	}
}

// retry executes the provided function repeatedly, retrying until the function returns with no error or exceeds the given attempts.
func retry(attempts int, interval time.Duration, fn func() error) error {
	if err := fn(); err != nil {
		if s, ok := err.(stop); ok {
			// Return the original error for later checking
			return s.error
		}
		if attempts--; attempts > 0 {
			time.Sleep(interval)
			return retry(attempts, interval, fn)
		}
		return err
	}
	return nil
}
