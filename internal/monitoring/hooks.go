package monitoring

import "log"

type Hooks interface {
	Observe(event string)
}

type NoopHooks struct{}

func (NoopHooks) Observe(string) {}

type LogHooks struct{}

func (LogHooks) Observe(event string) {
	log.Printf("monitoring_event=%s", event)
}
