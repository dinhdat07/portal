package metrics

import (
	portalmetrics "portal-system/internal/metrics"

	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusOutboxMetrics struct {
	eventsClaimedTotal        prometheus.Counter
	eventsPublishedTotal      prometheus.Counter
	eventsPublishFailedTotal  prometheus.Counter
	eventsRetryScheduledTotal prometheus.Counter
	eventsDeadLetteredTotal   prometheus.Counter
}

var _ portalmetrics.OutboxMetrics = (*PrometheusOutboxMetrics)(nil)

func NewPrometheusOutboxMetrics(registerer prometheus.Registerer) *PrometheusOutboxMetrics {
	if registerer == nil {
		registerer = prometheus.DefaultRegisterer
	}

	m := &PrometheusOutboxMetrics{
		eventsClaimedTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "outbox_events_claimed_total",
			Help: "Total number of outbox events claimed by the worker",
		}),
		eventsPublishedTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "outbox_events_published_total",
			Help: "Total number of outbox events published to Kafka successfully",
		}),
		eventsPublishFailedTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "outbox_events_publish_failed_total",
			Help: "Total number of outbox events publish attempts that failed",
		}),
		eventsRetryScheduledTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "outbox_events_retry_scheduled_total",
			Help: "Total number of outbox events scheduled for retry after failures",
		}),
		eventsDeadLetteredTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "outbox_events_dead_lettered_total",
			Help: "Total number of outbox events marked as dead lettered",
		}),
	}

	registerer.MustRegister(
		m.eventsClaimedTotal,
		m.eventsPublishedTotal,
		m.eventsPublishFailedTotal,
		m.eventsRetryScheduledTotal,
		m.eventsDeadLetteredTotal,
	)

	return m
}

func (m *PrometheusOutboxMetrics) EventsClaimed(count int) {
	if count <= 0 {
		return
	}
	m.eventsClaimedTotal.Add(float64(count))
}

func (m *PrometheusOutboxMetrics) EventsPublished() {
	m.eventsPublishedTotal.Inc()
}

func (m *PrometheusOutboxMetrics) EventsPublishFailed() {
	m.eventsPublishFailedTotal.Inc()
}

func (m *PrometheusOutboxMetrics) EventsRetryScheduled() {
	m.eventsRetryScheduledTotal.Inc()
}

func (m *PrometheusOutboxMetrics) EventsDeadLettered() {
	m.eventsDeadLetteredTotal.Inc()
}
