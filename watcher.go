package main

import (
	"io/ioutil"
	"log"
	"os"
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
	d := b2s(data)

	list := strings.Split(d, "\n")
	if len(list[len(list)-1]) == 0 {
		list = list[:len(list)-1]
	}

	if len(wc.list) != 0 {
		nt := false
		new := []string{}
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
				if !nt {
					nt = true
				}
				new = append(new, list[i])
			}
		}
		if len(new) != 0 {
			log.Printf("%v added to the list\n", new)
		}
		if !nt { // reverse search
			old := []string{}
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
					old = append(old, wc.list[i])
				}
			}
			if len(old) != 0 {
				log.Printf("%v deleted to the list\n", old)
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
				if _, err := os.Stat(file); err != nil {
					log.Printf("%s have been deleted. Closing fsnotify.", file)
					return
				}
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
