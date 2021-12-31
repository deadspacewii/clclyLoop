package event

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	DEFAULTINTERVAL = 10 * time.Second
)

var TIMETASKMAP = map[string]*TimeTask{}

type TimeTask struct {
	taskName string           // task name
	funcJob  Job              // running task
	running  bool             // if running
	stop     bool             // if stop
	result   chan interface{} // channel of result
	err      chan error       // channel of err
	success  int64            // count of success
	fail     int64            // count of fail
	op       eventOption      // option
	timer    *time.Timer      // timer
	mu       sync.Mutex       // concurrency safe
}


// NewTimeTask
func NewTimeTask(options eventOption) (EventLoop, error) {
	timeTask := TimeTask{
		running:  false,
		stop:     false,
		result:   make(chan interface{}),
		err:      make(chan error),
		mu:       sync.Mutex{},
	}
	if options.interval <= 0 {
		return nil, errors.New("task interval shouldn't under zero")
	}
	timeTask.op = options
	go func(time *TimeTask) {
		if time.stop {
			return
		}
		for {
			select {
			case <- time.err:
				time.fail ++
			case <- time.result:
				time.success ++
			}
			if time.stop {
				return
			}
		}
	}(&timeTask)
	return &timeTask, nil
}


// WithOptionInterval 插入定时器的间隔时间
func WithOptionInterval(i time.Duration) EventOption {
	if i == 0 {
		return EventOption{
			f: func(option *eventOption) {
				option.interval = DEFAULTINTERVAL
			},
		}
	}
	return EventOption{
		f: func(option *eventOption) {
			option.interval = i
		},
	}
}

// WithOptionHook 实现定时器hook函数注册
func WithOptionHook(f1 func([]interface{})) EventOption {
	if f1 == nil {
		return EventOption{
			f: func(option *eventOption) {
				return
			},
		}
	}
	return EventOption{
		f: func(option *eventOption) {
			option.hooks = append(option.hooks, f1)
		},
	}
}

// WithOptionLoop 确定hook函数只执行一次还是循环执行
func WithOptionLoop(l bool) EventOption {
	return EventOption{
		f: func(option *eventOption) {
			option.loop = l
		},
	}
}

// WithOptionValue 确定对比值
func WithOptionValue(i interface{}) EventOption {
	return EventOption{
		f: func(option *eventOption) {
			option.value = i
		},
	}
}

// Regist 注册函数
func Regist(options ...EventOption) eventOption {
	op := eventOption{
		hooks: make([]func([]interface{}), 0),
	}
	for _, option := range options {
		option.f(&op)
	}
	return op
}

// Add
func (t *TimeTask) Add(tName string, job func() (interface{}, error)) {
	if t.taskName == tName {
		return
	}
	t.taskName = tName
	if job == nil {
		return
	}
	t.add(tName, job)
}

func (t *TimeTask) add(tName string, job func() (interface{}, error)) {
	t.funcJob  = RunnableJob(job)
	TIMETASKMAP[tName] = t
}

// Update
func (t *TimeTask) Update(tName string, job func() (interface{}, error), options eventOption) {
	if _, ok := TIMETASKMAP[tName]; !ok {
		fmt.Println("task not exist")
		return
	}
	t.add(tName, job)
	t.op = options
}

// Start
func (t *TimeTask) Start() {
	if t.taskName == "" || t.funcJob == nil {
		fmt.Println("task not exist")
		return
	}
	for {
		if t.running {
			continue
		} else {
			break
		}
	}
	t.running = true
	go func() {
		defer func() {
			t.running = false
		}()
		timer := time.NewTimer(t.op.interval)
		t.timer = timer
		for {
			if t.stop {
				return
			}
			select {
			case <- t.timer.C:
				res, err1 := t.funcJob.Run()
				if err1 != nil {
					t.err <- err1
				}
				t.result <- res
				go t.runHook(err1, []interface{}{t, t.op.value, res})
				return
			}
		}
	}()
}

func (t *TimeTask) runHook(err error, result []interface{}) {
	if len(t.op.hooks) == 0 {
		return
	}
	if t.op.loop {
		t.cycleHook(err, result)
		return
	}
	t.nomalHook(err, result)
}

func (t *TimeTask) cycleHook(err error, result []interface{}) {
	if err != nil {
		return
	}
	f1 := t.op.hooks[0]
	if len(t.op.hooks) > 1 {
		t.op.hooks = t.op.hooks[1:]
	}
	t.op.hooks = append(t.op.hooks, f1)
	f1(result)
}

func (t *TimeTask) nomalHook(err error, result []interface{}) {
	if err != nil {
		return
	}
	f1 := t.op.hooks[0]
	if len(t.op.hooks) > 1 {
		t.op.hooks = t.op.hooks[1:]
	} else {
		t.op.hooks = t.op.hooks[:0]
	}
	f1(result)
	if len(t.op.hooks) == 0 {
		t.Stop()
	}
}

func (t *TimeTask) Stop() {
	defer t.mu.Unlock()
	t.mu.Lock()
	go func() {
		for {
			if t.running {
				continue
			} else {
				t.stop = true
				return
			}
		}
	}()
}

func (t *TimeTask) GetSuccess() int64 {
	return t.success
}

func (t *TimeTask) ModifyTime(interval time.Duration) {
	t.op.interval = interval
	t.timer.Reset(interval)
}