package sentry

import gosentry "github.com/getsentry/sentry-go"

// ExtractExtra implements sentry.Integration interface.
// The logic of this structure is integrated with Sentry.
type ExtractExtra struct{}

// Name implements sentry.Integration interface.
func (ee ExtractExtra) Name() string {
	return "ExtractExtra"
}

// SetupOnce implements sentry.Integration interface.
func (ee ExtractExtra) SetupOnce(client *gosentry.Client) {
	client.AddEventProcessor(func(event *gosentry.Event, hint *gosentry.EventHint) *gosentry.Event {
		if hint == nil {
			return event
		}
		if ex, ok := hint.OriginalException.(CustomComplexError); ok {
			for key, val := range ex.MoreData {
				if event.Extra == nil {
					event.Extra = make(map[string]interface{})
				}
				event.Extra[key] = val
			}
		}

		return event
	})
}

// EventFormatter implements sentry.Integration interface.
// The logic of this structure is integrated with Sentry.
// This is needed to format errors properly.
type EventFormatter struct{}

// Name implements sentry.Integration interface.
func (ee EventFormatter) Name() string {
	return "EventFormatter"
}

// SetupOnce implements sentry.Integration interface.
func (ee EventFormatter) SetupOnce(client *gosentry.Client) {
	client.AddEventProcessor(func(event *gosentry.Event, hint *gosentry.EventHint) *gosentry.Event {
		for i := range event.Exception {
			event.Exception[i].Type = event.Message

		}
		return event
	})
}
