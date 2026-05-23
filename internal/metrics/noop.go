package metrics

type NoopOutboxMetrics struct{}

var _ OutboxMetrics = NoopOutboxMetrics{}

func (NoopOutboxMetrics) EventsClaimed(count int) {}
func (NoopOutboxMetrics) EventsPublished()        {}
func (NoopOutboxMetrics) EventsPublishFailed()    {}
func (NoopOutboxMetrics) EventsRetryScheduled()   {}
func (NoopOutboxMetrics) EventsDeadLettered()     {}
