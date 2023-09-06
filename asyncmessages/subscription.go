package asyncmessages

type (
	Subscription interface {
		Unsubscribe() error
	}
)
