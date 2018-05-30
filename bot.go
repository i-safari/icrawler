package main

import (
	"fmt"
	"log"
	"path"
	"strconv"

	"github.com/ahmdrz/goinsta"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func state(wc *watcherController, c *nConn) {
	db, err := gorm.Open("sqlite3", *dbFile)
	if err != nil {
		log.Println("error opening database: %s\n", err)
		return
	}
	db.LogMode(false)
	defer db.Close()

	user := &User{}
	feed := &Feed{}
	story := &Stories{}
	db.CreateTable(user, feed, story)

	for _, name := range wc.list {
		err := db.Where("username = ?", name).Find(user).Error
		if err != nil { // user does not exist in database
			if err != gorm.ErrRecordNotFound {
				log.Printf("error getting record: %s", err)
			}
			log.Println(name, "not found in db")
			guser, err := insta.Profiles.ByName(name)
			if err != nil {
				log.Printf("error getting profile of %s", name)
				continue
			}

			// downloading user highlights
			hlgts, err := guser.Highlights()
			if err != nil {
				log.Printf("error downloading %s highlights: %s", guser.Username, err)
			}
			// using user gender to store highlights
			copyGuserToUser(guser, user)
			user.Highlights = len(hlgts)

			// saving user
			err = db.Create(user).Error
			if err != nil {
				log.Printf("error saving user in database: %s", err)
				continue
			}

			c.logger.Printf("Downloading highlights of %s (%d)\n", guser.Username, guser.ID)
			for _, h := range hlgts {
				output := path.Join(*outDir, strconv.FormatInt(guser.ID, 10), "highlights", h.Title)
			iloop:
				for _, item := range h.Items {
					imgs, vds, err := item.Download(output, "")
					if err != nil {
						log.Println(err)
						continue iloop
					}
					story.Title = h.Title
					story.Path, story.Url = imgs, goinsta.GetBest(item.Images.Versions)
					if vds != "" {
						// if item is a video is not an image. (xd)
						story.Path, story.Url = vds, goinsta.GetBest(item.Videos)
					}
					story.Highlight = true
					copyItemToStory(&item, story)

					err = db.Save(story).Error
					if err != nil {
						log.Println(err)
						continue iloop
					}
					c.logger.Printf("Downloaded in %s\n", story.Path)
				}
			}

			media := guser.Feed(nil)

			c.logger.Printf("Downloading feed media of %s (%d)\n", guser.Username, guser.ID)
			for media.Next() {
				// saving feed media in *outDir/{userid}/feed
				output := path.Join(*outDir, strconv.FormatInt(guser.ID, 10), "feed")
				for _, item := range media.Items {
					imgs, vds, err := item.Download(output, "")
					if err != nil {
						log.Println(err)
						continue
					}
					feed.Path, feed.Url = imgs, goinsta.GetBest(item.Images.Versions)
					if vds != "" {
						feed.Path, feed.Url = vds, goinsta.GetBest(item.Videos)
					}

					copyItemToFeed(&item, feed)
					err = db.Save(feed).Error
					if err != nil {
						log.Println(err)
					}
					c.logger.Printf("Downloaded in %s\n", feed.Path)
				}
			}
		}

		// getting new user strucure
		nguser, err := insta.Profiles.ByID(user.Id)
		if err != nil {
			log.Printf("error getting profile of %s (%d)", name, nguser.ID)
			return
		}

		// checking all values.
		// up is used to update database values
		up := false
		if nguser.Username != user.Username {
			up = true
			log.Printf("%s changed username to %s", user.Username, nguser.Username)
		}
		if nguser.FullName != user.FullName {
			up = true
			log.Printf(
				"%s changed fullname from '%s' to '%s'",
				user.Username, user.FullName, nguser.FullName,
			)
		}
		if nguser.Biography != user.Biography {
			up = true
			log.Printf(
				"%s changed biography from '%s' to '%s'",
				user.Username, user.Biography, nguser.Biography,
			)
		}
		if nguser.PublicEmail != user.Email {
			up = true
			log.Printf(
				"%s changed email from '%s' to '%s'",
				user.Username, user.Email, nguser.PublicEmail,
			)
		}
		if nguser.PublicPhoneNumber != user.Phone {
			if nguser.PublicPhoneNumber == "" {
				log.Printf("%s deleted his/her phone number", nguser.Username)
			} else {
				up = true // do not update
				log.Printf(
					"%s changed phone number from '%s' to '%s'",
					user.Username, user.Phone, nguser.PublicPhoneNumber,
				)
			}
		}
		if n := nguser.FollowerCount - user.Followers; n != 0 {
			up = true
			if n > 0 {
				log.Printf("%s has %d new follows", user.Username, n)
			} else {
				n = (n ^ -1) + 1
				log.Printf("%s has %d new unfollows", user.Username, n)
			}
		}
		if n := nguser.FollowingCount - user.Following; n != 0 {
			up = true
			if n > 0 {
				log.Printf("%s started following %d users", user.Username, n)
			} else {
				n = (n ^ -1) + 1
				log.Printf("%s stopped following %d users", user.Username, n)
			}
		}
		// TODO: check deleted values
		if n := nguser.MediaCount - user.Media; n != 0 {
			up = true
			if n > 0 {
				log.Printf("%s has %d new medias", user.Username, n)
			} else {
				n = (n ^ -1) + 1
				log.Printf("%s has deleted %d medias", user.Username, n)
			}
			gfeed := nguser.Feed(nil)

			i := 0
			for gfeed.Next() {
				for _, item := range gfeed.Items {
					i++
					copyItemToFeed(&item, feed)
					if !db.NewRecord(feed) {
						continue
					}
					v := false

					output := path.Join(*outDir, strconv.FormatInt(nguser.ID, 10), "feed")
					imgs, vds, err := item.Download(output, "")
					if err != nil {
						log.Println(err)
						continue
					}
					feed.Path, feed.Url = imgs, goinsta.GetBest(item.Images.Versions)
					if vds != "" {
						v, feed.Path, feed.Url = true, vds, goinsta.GetBest(item.Videos)
					}

					err = db.Save(feed).Error
					if err != nil {
						log.Println(err)
					}

					if v {
						c.SendVideo(fmt.Sprintf("New media of %s", nguser.Username), feed.Path)
					} else {
						c.SendPhoto(fmt.Sprintf("New media of %s", nguser.Username), feed.Path)
					}
					c.logger.Printf("Downloaded in %s\n", feed.Path)
				}
				if n < i {
					break
				}
			}
		}

		stories := nguser.Stories()
		for stories.Next() {
		itemLoop:
			for _, item := range stories.Items {
				copyItemToStory(&item, story)
				if !db.NewRecord(item) {
					continue
				}
				v := false
				output := path.Join(*outDir, strconv.FormatInt(nguser.ID, 10), "stories")

				imgs, vds, err := item.Download(output, "")
				if err != nil {
					log.Println(err)
					continue itemLoop
					continue
				}
				story.Path, story.Url = imgs, goinsta.GetBest(item.Images.Versions)
				if vds != "" {
					v, story.Path, story.Url = true, vds, goinsta.GetBest(item.Videos)
				}
				story.Highlight = false

				err = db.Save(story).Error
				if err != nil {
					log.Println(err)
					continue itemLoop
				}

				if v {
					c.SendVideo(fmt.Sprintf("New story of %s", nguser.Username), story.Path)
				} else {
					c.SendPhoto(fmt.Sprintf("New story of %s", nguser.Username), story.Path)
				}
				c.logger.Printf("Downloaded in %s\n", story.Path)
			}
		}

		if nguser.IsPrivate != user.IsPrivate {
			err = nguser.FriendShip()
			if err != nil {
				log.Println(err)
				goto end
			}
			if nguser.IsPrivate {
				msg := "%s is private now. "
				if !nguser.Friendship.Following {
					msg += "And you doesn't follow this user"
				}
				log.Printf(msg, nguser.Username)
			}
			up = true
		}
	end:
		if up {
			copyGuserToUser(nguser, user)
			err = db.Save(user).Error
			if err != nil {
				log.Println(err)
			}
		}
	}
}
