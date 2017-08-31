package operator

import (
	"fmt"
	"sync"

	"github.com/giantswarm/draughtsmantpr"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/giantswarm/operatorkit/framework"
	"github.com/giantswarm/operatorkit/tpr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// Config represents the configuration used to create a new service.
type Config struct {
	// Dependencies.
	K8sClient         kubernetes.Interface
	Logger            micrologger.Logger
	OperatorFramework *framework.Framework
	Resources         []framework.Resource
}

// DefaultConfig provides a default configuration to create a new service by
// best effort.
func DefaultConfig() Config {
	return Config{
		// Dependencies.
		K8sClient:         nil,
		Logger:            nil,
		OperatorFramework: nil,
		Resources:         nil,
	}
}

// Service implements the operator service.
type Service struct {
	// Dependencies.
	logger            micrologger.Logger
	operatorFramework *framework.Framework
	resources         []framework.Resource

	// Internals.
	bootOnce       sync.Once
	draughtsmanTPR *tpr.TPR
	mutex          sync.Mutex
}

// New creates a new configured service.
func New(config Config) (*Service, error) {
	// Dependencies.
	if config.K8sClient == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.K8sClient must not be empty")
	}
	if config.Logger == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.Logger must not be empty")
	}
	if config.OperatorFramework == nil {
		return nil, microerror.Maskf(invalidConfigError, "config.OperatorFramework must not be empty")
	}
	if len(config.Resources) == 0 {
		return nil, microerror.Maskf(invalidConfigError, "config.Resources must not be empty")
	}

	var err error
	var draughtsmanTPR *tpr.TPR
	{
		tprConfig := tpr.DefaultConfig()

		tprConfig.K8sClient = config.K8sClient
		tprConfig.Logger = config.Logger

		tprConfig.Description = draughtsmantpr.Description
		tprConfig.Name = draughtsmantpr.Name
		tprConfig.Version = draughtsmantpr.VersionV1

		draughtsmanTPR, err = tpr.New(tprConfig)
		if err != nil {
			return nil, microerror.Mask(err)
		}
	}

	newService := &Service{
		// Dependencies.
		logger:            config.Logger,
		operatorFramework: config.OperatorFramework,
		resources:         config.Resources,

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
		err := s.draughtsmanTPR.CreateAndWait()
		if tpr.IsAlreadyExists(err) {
			s.logger.Log("debug", "third party resource already exists")
		} else if err != nil {
			s.logger.Log("error", fmt.Sprintf("%#v", err))
			return
		}

		s.logger.Log("debug", "starting list/watch")

		newResourceEventHandler := &cache.ResourceEventHandlerFuncs{
			AddFunc:    s.addFunc,
			DeleteFunc: s.deleteFunc,
			UpdateFunc: s.updateFunc,
		}
		newZeroObjectFactory := &tpr.ZeroObjectFactoryFuncs{
			NewObjectFunc:     func() runtime.Object { return &draughtsmantpr.CustomObject{} },
			NewObjectListFunc: func() runtime.Object { return &draughtsmantpr.List{} },
		}

		s.draughtsmanTPR.NewInformer(newResourceEventHandler, newZeroObjectFactory).Run(nil)
	})
}

func (s *Service) addFunc(obj interface{}) {
	// We lock the addFunc/deleteFunc/updateFunc to make sure only one
	// addFunc/deleteFunc/updateFunc is executed at a time.
	// addFunc/deleteFunc/updateFunc is not thread safe. This is important because
	// the source of truth for the draughtsman-operator are Kubernetes resources.
	// In case we would run the operator logic in parallel, we would run into race
	// conditions.
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.logger.Log("debug", "executing the operator's addFunc")

	err := s.operatorFramework.ProcessCreate(obj, s.resources)
	if err != nil {
		s.logger.Log("error", fmt.Sprintf("%#v", err), "event", "create")
	}
}

func (s *Service) deleteFunc(obj interface{}) {
	// We lock the addFunc/deleteFunc/updateFunc to make sure only one
	// addFunc/deleteFunc/updateFunc is executed at a time.
	// addFunc/deleteFunc/updateFunc is not thread safe. This is important because
	// the source of truth for the draughtsman-operator are Kubernetes resources.
	// In case we would run the operator logic in parallel, we would run into race
	// conditions.
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.logger.Log("debug", "executing the operator's deleteFunc")

	err := s.operatorFramework.ProcessDelete(obj, s.resources)
	if err != nil {
		s.logger.Log("error", fmt.Sprintf("%#v", err), "event", "delete")
	}
}

func (s *Service) updateFunc(oldObj, newObj interface{}) {
	// We lock the addFunc/deleteFunc/updateFunc to make sure only one
	// addFunc/deleteFunc/updateFunc is executed at a time.
	// addFunc/deleteFunc/updateFunc is not thread safe. This is important because
	// the source of truth for the draughtsman-operator are Kubernetes resources.
	// In case we would run the operator logic in parallel, we would run into race
	// conditions.
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.logger.Log("debug", "executing the operator's updateFunc")

	err := s.operatorFramework.ProcessUpdate(newObj, s.resources)
	if err != nil {
		s.logger.Log("error", fmt.Sprintf("%#v", err), "event", "update")
	}
}
