package main

import (
	"io/ioutil"
	"log"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type watcherController struct {
	list    []string
	file    string
	running bool
	locker  sync.Mutex
}

func (wc *watcherController) dump() {
	data, err := ioutil.ReadFile(wc.file)
	if err != nil {
		log.Println("error opening targets file", err)
		return
	}
	d := string(data)

	list := strings.Split(d, "\n")
	if len(list[len(list)-1]) == 0 {
		list = list[:len(list)-1]
	}

	if len(wc.list) != 0 {
		nt := false
		for i := range list {
			nw := true //new
		nloop:
			for n := range wc.list {
				if list[i] == wc.list[n] {
					nw = false
					break nloop
				}
			}
			if nw {
				nt = true
				log.Printf("%s added to the list", list[i])
			}
		}
		if !nt { // reverse search
			for i := range wc.list {
				dt := true //deleted
			sloop:
				for n := range list {
					if wc.list[i] == list[n] {
						dt = false
						break sloop
					}
				}
				if dt {
					log.Printf("%s deleted to the list", wc.list[i])
				}
			}
		}
	}

	wc.locker.Lock()
	wc.list = append(wc.list[:0], list...)
	wc.locker.Unlock()
}

func (wc *watcherController) do(file string) {
	if wc.running {
		return
	}
	wc.running = true
	wc.file = file
	wc.dump()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		defer func() {
			watcher.Close()
			wc.running = false
		}()
		for event := range watcher.Events {
			if event.Op&fsnotify.Write == fsnotify.Write {
				wc.dump()
			}
			if event.Op&fsnotify.Remove == fsnotify.Remove {
				err = watcher.Add(file)
				if err != nil {
					log.Println(err)
				}
				wc.dump()
			}
		}
	}()

	err = watcher.Add(file)
	if err != nil {
		log.Fatal(err)
	}
}
