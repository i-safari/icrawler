package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
)

type opts struct {
	f, w bool
	m, s bool
	p, h bool
	nm   bool
	name string
}

type watcherController struct {
	list    []opts
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

	listv := strings.Split(d, "\n")

	msg1 := wc.getMsg()
	wc.toOpts(listv)
	msg2 := wc.getMsg()

	msg := ""
	list = strings.Split(msg2, "\n")
floop:
	for _, line := range strings.Split(msg1, "\n") {
		for _, line2 := range list {
			if strings.Split(line, " ")[0] == strings.Split(line2, " ")[0] {
				if !strings.EqualFold(line, line2) {
					msg += "+ " line2 + "\n"
				}
				continue floop
			}
		}
		msg += "- " line + "\n"
	}

	log.Println(msg)
}

func (wc *watcherController) getMsg() {
	msg := ""
	wlist := wc.list
	for _, user := range wc.list {
		msg += user.name
		if user.f {
			msg += " !followers"
		}
		if user.w {
			msg += " !following"
		}
		if user.m {
			msg += " !media"
		}
		if user.s {
			msg += " !stories"
		}
		if user.p {
			msg += " !profile"
		}
		if user.h {
			msg += " !highlights"
		}
		if user.nm {
			msg += " only new media"
		}
		msg += "\n"
	}
	return msg
}

// topOpts converts list ([]string) to slice of options
func (wc *watcherController) toOpts(list []string) {
userLoop:
	for _, user := range list {
		if len(user) < 3 {
			continue
		}
		o := opts{}
		i := strings.IndexByte(user, ' ')
		if i == -1 {
			o.name = user
			wlist = append(wlist, o)
			continue userLoop
		}
		o.name, user = user[0:i], user[i+1:]
		for i = 0; i < len(user); i++ {
			switch user[i] {
			case ' ':
				continue
			case 'f':
				o.f = true
			case 'w':
				o.w = true
			case 'm':
				o.m = true
			case 's':
				o.m = true
			case 'h':
				o.h = true
			case 'p':
				o.p = true
			case 'n':
				o.nm = true
			}
		}
		wlist = append(wlist, o)
	}

	wc.locker.Lock()
	wc.list = wlist
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
