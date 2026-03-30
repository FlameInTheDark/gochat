package cache

type TimedOption interface {
	applyTimed(*TimedOptions)
}

type TimedOptions struct {
	Proactive bool
}

type timedOptionFunc func(*TimedOptions)

func (f timedOptionFunc) applyTimed(opts *TimedOptions) {
	if f != nil {
		f(opts)
	}
}

func DefaultTimedOptions() TimedOptions {
	return TimedOptions{Proactive: true}
}

func ResolveTimedOptions(opts ...TimedOption) TimedOptions {
	cfg := DefaultTimedOptions()
	for _, opt := range opts {
		if opt != nil {
			opt.applyTimed(&cfg)
		}
	}
	return cfg
}

func NoneProactive() TimedOption {
	return timedOptionFunc(func(opts *TimedOptions) {
		opts.Proactive = false
	})
}
