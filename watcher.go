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
	if len(listv[len(listv)-1]) == 0 {
		listv = listv[:len(listv)-1]
	}
	// (f)ollowers
	// follo(w)ing (m)edia (s)tory
	// (p)rofile changes
	list := make([]string, 0, len(listv))
	for i := range listv {
		list = append(list, strings.Split(listv[i], " ")[0])
	}

	if len(wc.list) != 0 {
		nt := false
		new := []string{}
		for i := range list {
			nw := true //new
		nloop:
			for n := range wc.list {
				if list[i] == wc.list[n].name {
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
					if wc.list[i].name == list[n] {
						dt = false
						break sloop
					}
				}
				if dt {
					old = append(old, wc.list[i].name)
				}
			}
			if len(old) != 0 {
				log.Printf("%v deleted to the list\n", old)
			}
		}
	}

	wc.toOpts(listv)

	msg := ""
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
}

// topOpts converts list ([]string) to slice of options
func (wc *watcherController) toOpts(list []string) {
	wlist := wc.list

userLoop:
	for _, user := range list {
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
