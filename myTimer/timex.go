package timex

import (
	"fmt"
	"sync"
	//"sync/atomic"
	"time"
)

type ContextHandler func(...*interface{}) bool
type onDone func(uint64)
type timerRoutine interface {
	execute() bool
	lock()
	unlock()
	stop(endCbk onDone)
	start(endCbk onDone)
}

const (
	OneShot         = iota
	Periodic        = iota
	PeriodicLimited = iota
)

type context struct {
	id       uint64
	m        sync.Mutex
	handler  ContextHandler
	contexts []*interface{}
	d        time.Duration
	endCbk   onDone
}

type oneShotTimer struct {
	c context
	t *time.Timer
}

type periodicTimer struct {
	c    context
	t    *time.Ticker
	n    uint64
	i    uint64
	done chan bool
}

func (t *oneShotTimer) execute() bool {
	return t.c.execute()
}

func (t *periodicTimer) execute() bool {
	if t.n > 0 {
		t.i++
	}
	return t.c.execute() && ((t.i) == t.n)
}

func (c *context) unlock() {
	c.m.Unlock()
}

func (c *context) lock() {
	c.m.Lock()
}

func (t *oneShotTimer) stop() {
	t.t.Stop()
}

func (t *periodicTimer) stop() {
	t.t.Stop()
}

func (t *oneShotTimer) lock() {
	t.c.lock()
}

func (t *periodicTimer) lock() {
	t.c.lock()
}

func (t *oneShotTimer) unlock() {
	t.c.unlock()
}

func (t *periodicTimer) unlock() {
	t.c.unlock()
}

func (t *oneShotTimer) run() {
	<-t.t.C
	t.lock()
	t.execute()
	t.unlock()
	t.c.endCbk(t.c.id)
}

func (t *oneShotTimer) start(endCbk *onDone) {
	t.t = time.NewTimer(t.c.d)
	t.c.endCbk = *endCbk
	go t.run()
}

func (t *periodicTimer) run() {
	for {
		select {
		case <-t.done:
			return
		case ts := <-t.t.C:
			fmt.Printf("Ticker %d ticket at %s", t.c.id, ts)
			t.lock()
			fin := t.execute()
			t.unlock()
			if fin {
				go t.c.endCbk(t.c.id)
			}
		}
	}
}

func (t *periodicTimer) start(endCbk *onDone) {
	t.t = time.NewTicker(t.c.d)
	t.c.endCbk = *endCbk
	t.done = make(chan bool)
	go t.run()
}

func (m *Manager) loadTimer(i uint64) *timerRoutine {
	timer, err := m.timers.Load(i)
	if !err {
		switch timer.(type) {
		case *oneShotTimer:
			return timer.(*timerRoutine)
		default:
			panic("Timer type is not correct!!!")
		}
	} else {
		panic("Loading already deleted timer!!!")
	}
	return nil
}

func (m *Manager) loadAndDeleteTimer(i uint64) *timerRoutine {
	timer, err := m.timers.LoadAndDelete(i)
	if !err {
		switch timer.(type) {
		case *oneShotTimer:
			return timer.(*timerRoutine)
		default:
			panic("Timer type is not correct!!!")
		}
	} else {
		panic("Loading already deleted timer!!!")
	}
	return nil
}

func (m *Manager) execute(i uint64) {
	timer := m.loadTimer(i)
	(*timer).execute()
}

func (c *context) execute() bool {
	return c.handler(c.contexts...)
}

type Manager struct {
	timers    sync.Map
	idCounter uint64
}

func (m *Manager) create(timerType int, handler ContextHandler, contexts ...*interface{}) uint64 {
	var x uint64 //todo get this dynamically
	switch timerType {
	case OneShot:
		m.timers.Store(x, &oneShotTimer{c: context{handler: handler, contexts: contexts}})
	case Periodic, PeriodicLimited:
		m.timers.Store(x, &periodicTimer{c: context{handler: handler, contexts: contexts}})
	default:
		panic("Unimplemented timer time!")
	}

	return x
}

func (m *Manager) Start(n uint64) {
	timer := m.loadTimer(n)
	(*timer).lock()
	defer (*timer).unlock()
	(*timer).execute()
	(*timer).start(m.End)
}

func (m *Manager) End(n uint64) {
	timer := m.loadAndDeleteTimer(n)
	(*timer).lock()
	defer (*timer).unlock()
}

func (m *Manager) InsertOneTimeTimer(duration time.Duration, handler ContextHandler, contexts ...*interface{}) uint64 {
	handler(contexts...)
	return 0
}
