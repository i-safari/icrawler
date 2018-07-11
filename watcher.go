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
	followers, following bool
	media, stories       bool
	profile, highlights  bool
	newMedia             bool
	name                 string
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
	list := strings.Split(msg1, "\n")
floop:
	for _, line2 := range strings.Split(msg2, "\n") {
		for _, line1 := range list {
			if strings.Split(line1, " ")[0] == strings.Split(line2, " ")[0] {
				if !strings.EqualFold(line1, line2) {
					msg += "- " + line1 + "\n"
					msg += "+ " + line2 + "\n"
				}
				continue floop
			}
		}
		msg += "+ " + line2 + "\n"
	}

	if msg != "" {
		log.Println(msg)
	}
}

func (wc *watcherController) getMsg() string {
	msg := ""
	for _, user := range wc.list {
		msg += user.name
		if user.followers {
			msg += " !followers"
		}
		if user.following {
			msg += " !following"
		}
		if user.media {
			msg += " !media"
		}
		if user.stories {
			msg += " !stories"
		}
		if user.profile {
			msg += " !profile"
		}
		if user.highlights {
			msg += " !highlights"
		}
		if user.newMedia {
			msg += " only new media"
		}
		msg += "\n"
	}
	return msg
}

// topOpts converts list ([]string) to slice of options
func (wc *watcherController) toOpts(list []string) {
	wlist := wc.list[:0]
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
				o.followers = true
			case 'w':
				o.following = true
			case 'm':
				o.media = true
			case 's':
				o.media = true
			case 'h':
				o.highlights = true
			case 'p':
				o.profile = true
			case 'n':
				o.newMedia = true
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
					log.Printf("%stories have been deleted. Closing fsnotify.", file)
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
