package async

type (
	Subscription interface {
		Unsubscribe() error
	}
)
