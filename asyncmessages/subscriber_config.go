package asyncmessages

import "time"

const (
	AckTypeAuto AckType = iota
	AckTypeManual
)

var (
	defaultAckWait      = 30 * time.Second
	defaultMaxRedeliver = 3
)

type (
	AckType int

	SubscriberConfig struct {
		msgFilter    []string
		groupName    string
		ackType      AckType
		ackWait      time.Duration
		maxRedeliver int
	}

	SubscriberOption interface {
		configureSubscriberConfig(*SubscriberConfig)
	}

	MessageFilter []string

	GroupName string

	AckWait time.Duration

	MaxRedeliver int
)

func NewSubscriberConfig(options []SubscriberOption) SubscriberConfig {
	cfg := SubscriberConfig{
		msgFilter:    []string{},
		groupName:    "",
		ackType:      AckTypeAuto,
		ackWait:      defaultAckWait,
		maxRedeliver: defaultMaxRedeliver,
	}
	for _, opt := range options {
		opt.configureSubscriberConfig(&cfg)
	}
	return cfg
}

func (c SubscriberConfig) MessageFilters() []string { return c.msgFilter }
func (c SubscriberConfig) GroupName() string        { return c.groupName }
func (c SubscriberConfig) AckType() AckType         { return c.ackType }
func (c SubscriberConfig) AckWait() time.Duration   { return c.ackWait }
func (c SubscriberConfig) MaxRedeliver() int        { return c.maxRedeliver }

func (s MessageFilter) configureSubscriberConfig(cfg *SubscriberConfig) { cfg.msgFilter = []string(s) }

func (n GroupName) configureSubscriberConfig(cfg *SubscriberConfig) { cfg.groupName = string(n) }

func (t AckType) configureSubscriberConfig(cfg *SubscriberConfig) { cfg.ackType = t }

func (w AckWait) configureSubscriberConfig(cfg *SubscriberConfig) { cfg.ackWait = time.Duration(w) }

func (i MaxRedeliver) configureSubscriberConfig(cfg *SubscriberConfig) { cfg.maxRedeliver = int(i) }
