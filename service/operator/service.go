package operator

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cenk/backoff"
	"github.com/giantswarm/draughtsmantpr"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/tpr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
)

const (
	ResyncPeriod = 10 * time.Second
)

// Config represents the configuration used to create a new service.
type Config struct {
	// Dependencies.
	BackOff           backoff.BackOff
	K8sClient         kubernetes.Interface
	Logger            micrologger.Logger
	OperatorFramework *framework.Framework
}

// DefaultConfig provides a default configuration to create a new service by
// best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		BackOff:           nil,
		K8sClient:         nil,
		Logger:            nil,
		OperatorFramework: nil,
	}
}

// Service implements the operator service.
type Service struct {
	// Dependencies.
	backOff           backoff.BackOff
	logger            micrologger.Logger
	operatorFramework *framework.Framework

	// Internals.
	bootOnce       sync.Once
	draughtsmanTPR *tpr.TPR
	mutex          sync.Mutex
}

// New creates a new configured service.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.BackOff == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.BackOff must not be empty")
	}
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.OperatorFramework == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.OperatorFramework must not be empty")
	}

	var err error
	var draughtsmanTPR *tpr.TPR
	{
		tprConfig := tpr.DefaultConfig()

		tprConfig.K8sClient = config.K8sClient
		tprConfig.Logger = config.Logger

		tprConfig.Description = draughtsmantpr.Description
		tprConfig.Name = draughtsmantpr.Name
		tprConfig.ResyncPeriod = ResyncPeriod
		tprConfig.Version = draughtsmantpr.VersionV1

		draughtsmanTPR, err = tpr.New(tprConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	newService := &Service{
		// Dependencies.
		backOff:           config.BackOff,
		logger:            config.Logger,
		operatorFramework: config.OperatorFramework,

		// Internals
		bootOnce:       sync.Once{},
		draughtsmanTPR: draughtsmanTPR,
		mutex:          sync.Mutex{},
	}

	return newService, nil
}

// Boot starts the service and implements the watch for the cluster TPR.
func (s *Service) Boot() {
	s.bootOnce.Do(func() {
		o := func() error {
			err := s.bootWithError()
			if err != nil {
				return microerror.Mask(err)
			}

			return nil
		}

		n := func(err error, d time.Duration) {
			s.logger.Log("warning", fmt.Sprintf("retrying operator boot due to error: %#v", microerror.Mask(err)))
		}

		err := backoff.RetryNotify(o, s.backOff, n)
		if err != nil {
			s.logger.Log("error", fmt.Sprintf("stop operator boot retries due to too many errors: %#v", microerror.Mask(err)))
			os.Exit(1)
		}
	})
}

func (s *Service) bootWithError() error {
	err := s.draughtsmanTPR.CreateAndWait()
	if tpr.IsAlreadyExists(err) {
		s.logger.Log("debug", "third party resource already exists")
	} else if err != nil {
		return microerror.Mask(err)
	}

	s.logger.Log("debug", "starting list/watch")

	newResourceEventHandler := s.operatorFramework.NewCacheResourceEventHandler()

	newZeroObjectFactory := &tpr.ZeroObjectFactoryFuncs{
		NewObjectFunc:     func() runtime.Object { return &draughtsmantpr.CustomObject{} },
		NewObjectListFunc: func() runtime.Object { return &draughtsmantpr.List{} },
	}

	s.draughtsmanTPR.NewInformer(newResourceEventHandler, newZeroObjectFactory).Run(nil)

	return nil
}
