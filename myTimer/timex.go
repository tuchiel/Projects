package timex

import (
	"Projects/myLoger"
	"sync"
	"sync/atomic"
	"time"
)

type ContextHandler func(interface{}) bool
type onDone func(uint64)
type timerRoutine interface {
	execute() bool
	lock()
	unlock()
	stop()
	start(endCbk onDone)
	printInfo()
}

var logger = logg.DefaultModuleLogger

type context struct {
	id      uint64
	m       sync.Mutex
	handler ContextHandler
	context interface{}
	d       time.Duration
	endCbk  onDone
	created time.Time
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

func (c *context) printInfo(typeString string) {
	logger.Debug("%s timer(%d) : started %s, duration %d miliseconds.\n", typeString, c.id, c.created, c.d/time.Millisecond)
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

func (t *oneShotTimer) lock() {
	t.c.lock()
}

func (t *periodicTimer) stop() {
	t.t.Stop()
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
	t.c.created = time.Now()
	ts := <-t.t.C
	logger.Debug("One shot %d fired at %s\n", t.c.id, ts)
	t.lock()
	t.execute()
	t.unlock()
	t.c.endCbk(t.c.id)
}

func (t *oneShotTimer) start(endCbk onDone) {
	t.t = time.NewTimer(t.c.d)
	t.c.endCbk = endCbk
	go t.run()
}

func (t *periodicTimer) run() {
	t.c.created = time.Now()
	for {
		select {
		case <-t.done:
			return
		case ts := <-t.t.C:
			logger.Debug("Ticker %d ticked at %s\n", t.c.id, ts)
			t.lock()
			fin := t.execute()
			t.unlock()
			if fin {
				t.c.endCbk(t.c.id)
				t.stop()
				return
			}
		}
	}
}

func (t *oneShotTimer) printInfo() {
	t.c.printInfo("One shot        ")
}

func (t *periodicTimer) printInfo() {
	if t.n > 0 {
		t.c.printInfo("Limited periodic")
	} else {
		t.c.printInfo("Periodic        ")

	}
}

func (t *periodicTimer) start(endCbk onDone) {
	t.t = time.NewTicker(t.c.d)
	t.c.endCbk = endCbk
	t.done = make(chan bool)
	go t.run()
}

func (m *Manager) startTimer(t timerRoutine) {
	t.lock()
	t.start(m.End)
	t.unlock()
}

func (m *Manager) stopTimer(t timerRoutine) {
	t.lock()
	t.stop()
	t.unlock()
}

func (m *Manager) loadAndStartTimer(i uint64) {
	timer, found := m.timers.Load(i)
	if found {
		switch timer.(type) {
		case *oneShotTimer:
			m.startTimer(timer.(*oneShotTimer))
		case *periodicTimer:
			m.startTimer(timer.(*periodicTimer))
		default:
			panic("Timer type is not correct!!!")
		}
	} else {
		panic("Starting already deleted timer!!!")
	}
}

func (c *context) execute() bool {
	return c.handler(c.context)
}

type Manager struct {
	timers    sync.Map
	idCounter uint64
}

func (m *Manager) Start(n uint64) {
	m.loadAndStartTimer(n)
}

func (m *Manager) Stop(n uint64) {
	m.End(n)
}

func (m *Manager) End(n uint64) {
	timer, found := m.timers.LoadAndDelete(n)
	if found {
		switch timer.(type) {
		case *oneShotTimer:
			m.stopTimer(timer.(*oneShotTimer))
		case *periodicTimer:
			m.stopTimer(timer.(*periodicTimer))
		default:
			panic("Timer type is not correct!!!")
		}
	}
}

func printTimerInfo(t timerRoutine) {
	t.lock()
	t.printInfo()
	t.unlock()
}

func (m *Manager) getIdx() uint64 {
	return atomic.AddUint64(&m.idCounter, 1)
}

func (m *Manager) PrintTimers() {
	m.timers.Range(func(k interface{}, timer interface{}) bool {
		switch timer.(type) {
		case *oneShotTimer:
			printTimerInfo(timer.(*oneShotTimer))
		case *periodicTimer:
			printTimerInfo(timer.(*periodicTimer))
		default:
			panic("Timer type is not correct!!!")
		}
		return true
	})
}

func (m *Manager) CreateOneTimeTimer(duration time.Duration, handler ContextHandler, ctx interface{}) uint64 {
	id := m.getIdx()
	m.timers.Store(id, &oneShotTimer{c: context{id: id, d: duration, handler: handler, context: ctx}})
	return id
}

func (m *Manager) CreatePeriodicTimer(duration time.Duration, handler ContextHandler, ctx interface{}) uint64 {
	id := m.getIdx()
	m.timers.Store(id, &periodicTimer{i: 1, n: 0, c: context{id: id, d: duration, handler: handler, context: ctx}})
	return id
}

func (m *Manager) CreateLimitedPeriodicTimer(numberOfRepeats uint64, duration time.Duration, handler ContextHandler, ctx interface{}) uint64 {
	id := m.getIdx()
	m.timers.Store(id, &periodicTimer{i: 0, n: numberOfRepeats, c: context{id: id, d: duration, handler: handler, context: ctx}})
	return id
}
