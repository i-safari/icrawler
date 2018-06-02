package main

import (
	"html/template"
	"log"
	"os"

	"github.com/valyala/bytebufferpool"
	tg "gopkg.in/telegram-bot-api.v4"
)

const tgTemplateText = `<b>{{.Title}}</b>

{{.Body}}`

type Data struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

// notification bot
type nConn struct {
	bot    *tg.BotAPI
	tmpl   *template.Template
	logger *log.Logger
	file   *os.File
}

func (c *nConn) Close() {
	c.file.Close()
}

func (c *nConn) Write(b []byte) (int, error) {
	text := b2s(b)

	if text[len(b)-1] == '\n' {
		c.logger.Print(text)
	} else {
		c.logger.Println(text)
	}
	if c.bot == nil {
		return -1, nil
	}

	data := Data{
		Title: "Instagram Notification",
		Body:  text,
	}

	bf := bytebufferpool.Get()
	defer bytebufferpool.Put(bf)
	c.tmpl.Execute(bf, data)

	_, err := c.bot.Send(
		tg.MessageConfig{
			BaseChat: tg.BaseChat{
				ChatID:           *tgID,
				ReplyToMessageID: 0,
			},
			Text: bf.String(),
			DisableWebPagePreview: false,
			ParseMode:             tg.ModeHTML,
		},
	)
	if err != nil {
		c.logger.Println(err)
	}
	return bf.Len(), err
}

func notDial(key string) *nConn {
	bot, err := tg.NewBotAPI(key)
	if err != nil {
		log.Println(err)
		return &nConn{nil, nil, nil, nil}
	}
	tmpl, _ := template.New("").Parse(tgTemplateText)
	file, err := os.OpenFile(
		*logfile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644,
	)
	if err != nil {
		log.Fatalln(err)
	}
	logger := log.New(file, "", log.LstdFlags)
	log.SetFlags(0)
	return &nConn{bot, tmpl, logger, file}
}

type data struct {
	Title string `json:"title"`
	Body  string `json:"body"`
}

func (c *nConn) SendPhoto(caption, file string) {
	data := Data{
		Title: "Instagram Notification",
		Body:  caption,
	}

	bf := bytebufferpool.Get()
	defer bytebufferpool.Put(bf)
	c.tmpl.Execute(bf, data)

	if bf.Len() > 200 {
		bf.B = append(bf.B[:197], "..."...)
	}
	_, err := c.bot.Send(
		tg.PhotoConfig{
			BaseFile: tg.BaseFile{
				BaseChat:    tg.BaseChat{ChatID: *tgID},
				File:        file,
				UseExisting: false,
			},
			Caption:   bf.String(),
			ParseMode: tg.ModeHTML,
		},
	)
	if err != nil {

	}
}

func (c *nConn) SendVideo(caption, file string) {
	data := Data{
		Title: "Instagram Notification",
		Body:  caption,
	}

	bf := bytebufferpool.Get()
	defer bytebufferpool.Put(bf)
	c.tmpl.Execute(bf, data)

	if bf.Len() > 200 {
		bf.B = append(bf.B[:197], "..."...)
	}
	_, err := c.bot.Send(
		tg.VideoConfig{
			BaseFile: tg.BaseFile{
				BaseChat:    tg.BaseChat{ChatID: *tgID},
				File:        file,
				UseExisting: false,
			},
			Caption:   bf.String(),
			ParseMode: tg.ModeHTML,
		},
	)
	if err != nil {
		log.Println(err)
	}
}
