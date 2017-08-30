package configurer

import (
	"github.com/giantswarm/draughtsman-operator/flag/service/configurer/configmap"
	"github.com/giantswarm/draughtsman-operator/flag/service/configurer/file"
	"github.com/giantswarm/draughtsman-operator/flag/service/configurer/secret"
)

type Configurer struct {
	ConfigMap configmap.ConfigMap
	File      file.File
	Secret    secret.Secret
	Types     string
}
