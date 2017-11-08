package versionbundle

import (
	"fmt"

	"github.com/coreos/go-semver/semver"
	"github.com/giantswarm/microerror"
)

type ReleaseConfig struct {
	Bundles []Bundle
}

func DefaultReleaseConfig() ReleaseConfig {
	return ReleaseConfig{
		Bundles: nil,
	}
}

type Release struct {
	bundles     []Bundle
	changelogs  []Changelog
	components  []Component
	deprecated  bool
	description string
	timestamp   string
	version     string
}

func NewRelease(config ReleaseConfig) (Release, error) {
	if len(config.Bundles) == 0 {
		return Release{}, microerror.Maskf(invalidConfigError, "config.Bundles must not be empty")
	}

	var changelogs []Changelog
	var components []Component
	var deprecated bool
	var timestamp string

	version, err := aggregateReleaseVersion(config.Bundles)
	if err != nil {
		return Release{}, microerror.Maskf(invalidConfigError, err.Error())
	}

	r := Release{
		bundles:    config.Bundles,
		changelogs: changelogs,
		components: components,
		deprecated: deprecated,
		timestamp:  timestamp,
		version:    version,
	}

	return r, nil
}

func (r Release) Bundles() []Bundle {
	return CopyBundles(r.bundles)
}

func (r Release) Changelogs() []Changelog {
	return r.changelogs
}

func (r Release) Components() []Component {
	return r.components
}

func (r Release) Deprecated() bool {
	return r.deprecated
}

func (r Release) Description() string {
	return r.description
}

func (r Release) Timestamp() string {
	return r.timestamp
}

func (r Release) Version() string {
	return r.version
}

func aggregateReleaseVersion(bundles []Bundle) (string, error) {
	var major int64
	var minor int64
	var patch int64

	for _, b := range bundles {
		v, err := semver.NewVersion(b.Version)
		if err != nil {
			return "", microerror.Mask(err)
		}

		major += v.Major
		minor += v.Minor
		patch += v.Patch
	}

	version := fmt.Sprintf("%d.%d.%d", major, minor, patch)

	return version, nil
}
