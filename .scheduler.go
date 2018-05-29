package main

import (
	"sort"
	"sync"
	"time"
)

type task struct {
	f     func()
	after time.Time
	timer *time.Timer
}

type scheduler struct {
	locker sync.Mutex
	tasks  []task
	close  bool
}

func newSchedule() *scheduler {
	sch := &scheduler{
		close: false,
	}
	go sch.do()
}

func (sch *scheduler) do() {
	sort.Slice(sch.tasks, func(i, j int) {
		return sch.tasks[i].after < sch.tasks[j].after
	})
	for i := range sch.tasks {
		sch.tasks[i].timer = time.NewTimer(sch.tasks[i].after)
	}
}

func (sch *scheduler) Close() {
	for i := range sch.tasks {
		sch.tasks[i].Stop()
		close(sch.tasks[i].C)
	}
	sch.locker.Lock()
	sch.tasks = nil
	sch.locker.Unlock()
}
