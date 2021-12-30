package event

import "time"

type EventLoop interface {
	Add(string, func() (interface{}, error))
	Update(string, func() (interface{}, error), eventOption)
	Start()
	Stop()
	GetSuccess() int64
	ModifyTime(t time.Duration)
}

type Job interface {
	Run() (interface{}, error)
}

type RunnableJob func() (interface{}, error)

func (r RunnableJob) Run() (interface{}, error) { return r() }

type EventOption struct {
	f func(*eventOption)
}

type eventOption struct {
	interval time.Duration
	hooks    []func([]interface{})
	loop     bool
	value    interface{}
}