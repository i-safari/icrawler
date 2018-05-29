package main

import (
	"fmt"
	"log"
	"path"
	"strconv"
)

func state(wc *watcherController, c *nConn) {
	// openning database
	db := dbOpen(*dbFile)
	defer db.Close()

	for _, name := range wc.list {
		user, err := db.Get(name)
		if err != nil { // user does not exist in database
			user, err = insta.Profiles.ByName(name)
			if err != nil {
				log.Printf("error getting profile of %s", name)
				continue
			}

			// downloading user highlights
			hlgts, err := user.Highlights()
			if err != nil {
				log.Println(err)
			}
			// using user gender to store highlights
			user.Gender = len(hlgts)

			// saving user
			err = db.Put(user)
			if err != nil {
				log.Println("error saving user in database:", err)
				continue
			}

			c.logger.Printf("Downloading highlights of %s (%d)\n", user.Username, user.ID)
			for _, h := range hlgts {
				output := path.Join(*outDir, strconv.FormatInt(user.ID, 10), "highlights", h.Title)
			iloop:
				for _, item := range h.Items {
					imgs, vds, err := item.Download(output, "")
					if err != nil {
						log.Println(err)
						continue iloop
					}
					to := imgs
					if vds != "" {
						// if item is a video is not an image. (xd)
						to = vds
					}
					err = db.PutStory(user, &item, true, to)
					if err != nil {
						log.Println(err)
						continue iloop
					}
					err = db.SetTitle(&item, h.Title)
					if err != nil {
						log.Println(err)
						continue iloop
					}
					c.logger.Printf("Downloaded in %s\n", to)
				}
			}

			media := user.Feed(nil)

			c.logger.Printf("Downloading feed media of %s (%d)\n", user.Username, user.ID)
			for media.Next() {
				// saving feed media in *outDir/{userid}/feed
				output := path.Join(*outDir, strconv.FormatInt(user.ID, 10), "feed")
				for _, item := range media.Items {
					imgs, vds, err := item.Download(output, "")
					if err != nil {
						log.Println(err)
						continue
					}
					to := imgs
					if vds != "" {
						to = vds
					}
					err = db.PutMedia(user, &item, to)
					if err != nil {
						log.Println(err)
					}
					c.logger.Printf("Downloaded in %s\n", to)
				}
			}
			continue
		}

		// getting new user strucure
		nuser, err := insta.Profiles.ByID(user.ID)
		if err != nil {
			log.Printf("error getting profile of %s (%d)", name, nuser.ID)
			return
		}

		// checking all values.
		// up is used to update database values
		up := false
		if nuser.Username != user.Username {
			up = true
			log.Printf("%s changed username to %s", user.Username, nuser.Username)
		}
		if nuser.FullName != user.FullName {
			up = true
			log.Printf(
				"%s changed fullname from '%s' to '%s'",
				user.Username, user.FullName, nuser.FullName,
			)
		}
		if nuser.Biography != user.Biography {
			up = true
			log.Printf(
				"%s changed biography from '%s' to '%s'",
				user.Username, user.Biography, nuser.Biography,
			)
		}
		if nuser.PublicEmail != user.PublicEmail {
			up = true
			log.Printf(
				"%s changed email from '%s' to '%s'",
				user.Username, user.PublicEmail, nuser.PublicEmail,
			)
		}
		if nuser.PublicPhoneNumber != user.PublicPhoneNumber {
			if nuser.PublicPhoneNumber == "" {
				log.Printf("%s deleted his/her phone number", user.Username)
			} else {
				up = true // do not update
				log.Printf(
					"%s changed phone number from '%s' to '%s'",
					user.Username, user.PublicPhoneNumber, nuser.PublicPhoneNumber,
				)
			}
		}
		if n := nuser.FollowerCount - user.FollowerCount; n != 0 {
			up = true
			if n > 0 {
				log.Printf("%s has %d new follows", user.Username, n)
			} else {
				n = (n ^ -1) + 1
				log.Printf("%s has %d new unfollows", user.Username, n)
			}
		}
		if n := nuser.FollowingCount - user.FollowingCount; n != 0 {
			up = true
			if n > 0 {
				log.Printf("%s started following %d users", user.Username, n)
			} else {
				n = (n ^ -1) + 1
				log.Printf("%s stopped following %d users", user.Username, n)
			}
		}
		// TODO: check deleted values
		if n := nuser.MediaCount - user.MediaCount; n != 0 {
			up = true
			if n > 0 {
				log.Printf("%s has %d new medias", user.Username, n)
			} else {
				n = (n ^ -1) + 1
				log.Printf("%s has deleted %d medias", user.Username, n)
			}
			feed := nuser.Feed(nil)

			i := 0
			for feed.Next() {
				for _, item := range feed.Items {
					i++
					if db.ExistsMedia(&item) {
						continue
					}
					output := path.Join(*outDir, strconv.FormatInt(user.ID, 10), "feed")
					imgs, vds, err := item.Download(output, "")
					if err != nil {
						log.Println(err)
						continue
					}
					v, to := false, imgs
					if vds != "" {
						to, v = vds, true
					}
					err = db.PutMedia(user, &item, to)
					if err != nil {
						log.Println(err)
					}

					if v {
						c.SendVideo(fmt.Sprintf("New media of %s", nuser.Username), to)
					} else {
						c.SendPhoto(fmt.Sprintf("New media of %s", nuser.Username), to)
					}
					c.logger.Printf("Downloaded in %s\n", to)
				}
				if n < i {
					break
				}
			}
		}

		stories := user.Stories()
		for stories.Next() {
			for _, item := range stories.Items {
				if db.ExistsStory(&item) {
					continue
				}
				output := path.Join(*outDir, strconv.FormatInt(user.ID, 10), "stories")
				imgs, vds, err := item.Download(output, "")
				if err != nil {
					log.Println(err)
					continue
				}
				v, to := false, imgs
				if vds != "" {
					v, to = true, vds
				}
				err = db.PutStory(user, &item, false, to)
				if err != nil {
					log.Println(err)
				}

				if v {
					c.SendVideo(fmt.Sprintf("New story of %s", nuser.Username), to)
				} else {
					c.SendPhoto(fmt.Sprintf("New story of %s", nuser.Username), to)
				}
				c.logger.Printf("Downloaded in %s\n", to)
			}
		}

		if nuser.IsPrivate != user.IsPrivate {
			err = nuser.FriendShip()
			if err != nil {
				log.Println(err)
				goto end
			}
			if nuser.IsPrivate {
				msg := "%s is private now. "
				if !nuser.Friendship.Following {
					msg += "And you doesn't follow this user"
				}
				log.Printf(msg, nuser.Username)
			}
			up = true
		}
	end:
		if up {
			err = db.Update(nuser)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
