package metrics

type OutboxMetrics interface {
	EventsClaimed(count int)
	EventsPublished()
	EventsPublishFailed()
	EventsRetryScheduled()
	EventsDeadLettered()
}
