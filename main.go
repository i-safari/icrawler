package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"unsafe"

	"github.com/ahmdrz/goinsta"
	"github.com/ahmdrz/goinsta/utils"
	"github.com/marcsantiago/gocron"
)

var (
	insta *goinsta.Instagram

	tgBot   = flag.String("n", "", "Telegram bot api id")
	tgID    = flag.Int64("g", 0, "Telegram chat id")
	targets = flag.String("t", "./targets", "Targets file")
	logfile = flag.String("l", "./icrawler.log", "Log file")
	outDir  = flag.String("o", "./files", "Output directory or storing directory")
	dbFile  = flag.String("d", "./instagram.db", "Instagram database")
	uptime  = flag.Uint64("u", 5, "Update time in minutes")
)

func init() {
	flag.Parse()
	if *tgID == 0 {
		panic("tgid must be specified")
	}
	if *tgBot == "" {
		panic("tgbot must be specified")
	}
	os.MkdirAll(*outDir, 0777)
}

func main() {
	// login
	insta = utils.New()

	conn := notDial(*tgBot)
	defer conn.Close()
	log.SetOutput(conn)

	wc := watcherController{}
	wc.do(*targets)

	gocron.Every(*uptime).Minutes().Do(state, &wc, conn)
	cron := gocron.Start()

	log.Printf("Bot started\nWatching %v", wc.list)
	go state(&wc, conn)

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)
	<-ch
	select {
	case cron <- struct{}{}: // stops cron jobs
	}
	log.Print("Stopping bot")
}

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
