package webhook

import (
	"time"

	"github.com/radiofrance/gitlab-ci-pipelines-exporter/pkg/gitlab_events"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// btof allows to convert quickly bool into float value
	btof = map[bool]float64{false: 0., true: 1.}

	// timestamp returns the timestamp that should be used by the timestamp* collector.
	// Putting it in a variable allows us to simulate it during the tests.
	timestamp = func() float64 { return float64(time.Now().Unix()) }
)

// handlePipelineEvent handles Gitlab pipeline events.
func (w Webhook) handlePipelineEvent(event gitlab_events.PipelineEvent) error {
	project, ref, kind := event.ProjectName(), event.Ref(), event.RefKind()
	labels := prometheus.Labels{"project": project, "ref": ref, "kind": kind.String()}

	w.collectors.ID.With(labels).Set(float64(event.ObjectAttributes.ID))
	w.collectors.Timestamp.With(labels).Set(timestamp())

	if event.ObjectAttributes.QueuedDuration.IsSome() {
		w.collectors.QueuedDurationSeconds.With(labels).Set(event.ObjectAttributes.QueuedDuration.Unwrap())
	}
	if event.ObjectAttributes.Duration.IsSome() {
		w.collectors.DurationSeconds.With(labels).Set(event.ObjectAttributes.Duration.Unwrap())
	}

	if event.ObjectAttributes.Status == gitlab_events.Running {
		w.collectors.RunCount.With(labels).Inc()
	}

	for _, status := range gitlab_events.Statuses[1:] {
		labels["status"] = status
		w.collectors.Status.With(labels).Set(btof[event.ObjectAttributes.Status.String() == status])
	}
	return nil
}

// handleJobEvent handles Gitlab job events.
func (w Webhook) handleJobEvent(event gitlab_events.JobEvent) error {
	project, ref, kind := event.ProjectName(), event.Ref(), event.RefKind()
	stage, job := event.Stage(), event.JobName()
	labels := prometheus.Labels{
		"project": project, "ref": ref, "kind": kind.String(),
		"stage": stage, "job_name": job,
	}

	w.collectors.JobID.With(labels).Set(float64(event.BuildID))
	w.collectors.JobTimestamp.With(labels).Set(timestamp())

	if event.BuildQueuedDuration.IsSome() {
		w.collectors.JobQueuedDurationSeconds.With(labels).Set(event.BuildQueuedDuration.Unwrap())
	}
	if event.BuildDuration.IsSome() {
		w.collectors.JobDurationSeconds.With(labels).Set(event.BuildDuration.Unwrap())
	}

	if event.BuildStatus == gitlab_events.Running {
		w.collectors.JobRunCount.With(labels).Inc()
	}

	for _, status := range gitlab_events.Statuses[1:] {
		labels["status"] = status
		w.collectors.JobStatus.With(labels).Set(btof[event.BuildStatus.String() == status])
	}
	return nil
}
