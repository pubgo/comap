package comap

import (
	"go.uber.org/atomic"
	"runtime"
	"sync"
)

func New(n ...uint32) *Map {
	var n1 = uint32(0)
	var n2 = uint32(100)
	if len(n) > 0 {
		n1 = n[0]
	}

	if len(n) > 1 {
		n2 = n[1]
	}

	m := &Map{data: make(map[interface{}]interface{}, n1), n: n1, dup: make(chan interface{}, n2)}
	runtime.SetFinalizer(m, func(p *Map) {
		p.closed.Store(true)
	})
	go m.run()
	return m
}

type setCommand struct {
	key   interface{}
	value interface{}
}

type delCommand struct {
	key interface{}
}

type Map struct {
	rw     sync.RWMutex
	n      uint32
	dup    chan interface{}
	writer atomic.Uint32
	closed atomic.Bool
	data   map[interface{}]interface{}
}

func (m *Map) run() {
	for {
		if m.writer.Load() > 0 {
			m.rw.Lock()
			for i := len(m.dup); i > 0; i-- {
				cmd := <-m.dup
				switch cmd := cmd.(type) {
				case delCommand:
					delete(m.data, cmd.key)
				case setCommand:
					m.data[cmd.key] = cmd.value
				}
			}
			m.rw.Unlock()
		}

		if m.closed.Load() {
			return
		}
		runtime.Gosched()
	}
}

func (m *Map) Set(k, v interface{}) {
	m.writer.Inc()
	m.dup <- setCommand{key: k, value: v}
}

func (m *Map) Get(k interface{}) interface{} {
	m.rw.RLock()
	defer m.rw.RUnlock()
	v, ok := m.data[k]
	if ok {
		return v
	}
	return nil
}

func (m *Map) Delete(k interface{}) {
	m.writer.Inc()
	m.dup <- delCommand{key: k}
}

func (m *Map) DeleteRand(n int) {
	m.rw.RLock()
	defer m.rw.RUnlock()
	for k := range m.data {
		if n == 0 {
			break
		}
		m.dup <- delCommand{key: k}
		n--
	}
}

func (m *Map) Clear() {
	m.rw.Lock()
	defer m.rw.Unlock()
	m.data = make(map[interface{}]interface{}, m.n)
	m.dup = nil
	*m = Map{}
}

func (m *Map) Rand() (k, v interface{}) {
	m.rw.RLock()
	defer m.rw.RUnlock()
	for k, v = range m.data {
		return
	}
	return nil, nil
}

func (m *Map) RandN(n int) map[interface{}]interface{} {
	var data = make(map[interface{}]interface{}, n)
	m.rw.RLock()
	for k, v := range m.data {
		if n == 0 {
			break
		}
		data[k] = v
		n--
	}
	m.rw.RUnlock()
	return data
}
