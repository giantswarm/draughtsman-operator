package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/giantswarm/microerror"
)

const (
	// deploymentStatusUrlFormat is the string format for the
	// GitHub API call for Deployment Statuses.
	// See: https://developer.github.com/v3/repos/deployments/#create-a-deployment-status
	deploymentStatusUrlFormat = "https://api.github.com/repos/%s/%s/deployments/%v/statuses"
)

// request makes a request, handling any metrics and logging.
func (e *Eventer) request(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", fmt.Sprintf("token %s", e.oauthToken))

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	// Update rate limit metrics.
	err = updateRateLimitMetrics(resp)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	return resp, err
}

// postDeploymentStatus posts a Deployment Status for the given Deployment.
func (e *Eventer) postDeploymentStatus(project string, id int, state deploymentStatusState) error {
	e.logger.Log("debug", "posting deployment status", "project", project, "id", id, "state", state)

	url := fmt.Sprintf(
		deploymentStatusUrlFormat,
		e.organisation,
		project,
		id,
	)

	status := deploymentStatus{
		State: state,
	}

	payload, err := json.Marshal(status)
	if err != nil {
		return microerror.Mask(err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return microerror.Mask(err)
	}

	startTime := time.Now()

	resp, err := e.request(req)
	if err != nil {
		return microerror.Mask(err)
	}
	defer resp.Body.Close()

	updateDeploymentStatusMetrics("POST", e.organisation, project, resp.StatusCode, startTime)

	if resp.StatusCode != http.StatusCreated {
		return microerror.Maskf(unexpectedStatusCode, fmt.Sprintf("received non-200 status code: %v", resp.StatusCode))
	}

	return nil
}
