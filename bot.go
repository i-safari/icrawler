package main

import (
	"errors"
	"fmt"
	"log"
	"path"
	"strconv"

	"github.com/ahmdrz/goinsta"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var (
	errIsCarousel = errors.New("is carousel media")
)

func init() {
	db, err := gorm.Open("sqlite3", *dbFile)
	if err != nil {
		log.Println("error opening database: %s\n", err)
		return
	}
	db.LogMode(false)
	db.CreateTable(&User{}, &Feed{}, &Stories{})
	db.Close()
}

func state(wc *watcherController, c *nConn) {
	db, err := gorm.Open("sqlite3", *dbFile)
	if err != nil {
		log.Println("error opening database: %s\n", err)
		return
	}
	db.LogMode(false)
	defer db.Close()

	for _, target := range wc.list {
		user := &User{Username: target.name}
		err := db.Where(user).Find(user).Error
		if err != nil { // user does not exist in database
			if err != gorm.ErrRecordNotFound {
				log.Printf("error getting record: %s", err)
			}
			guser, err := insta.Profiles.ByName(target.name)
			if err != nil {
				log.Printf("error getting profile of %s", target.name)
				continue
			}

			// saving user
			copyGuserToUser(guser, user)
			err = db.Create(user).Error
			if err != nil {
				log.Printf("error saving %s in database: %s", target.name, err)
				continue
			}

			if !target.h {
				// downloading user highlights
				hlgts, err := guser.Highlights()
				if err != nil {
					log.Printf("error downloading %s highlights: %s", guser.Username, err)
				}
				user.Highlights = len(hlgts)

				c.logger.Printf("Downloading highlights of %s (%d)\n", guser.Username, guser.ID)
				for _, h := range hlgts {
					for _, item := range h.Items {
						downloadAndStoreStory(db, &item, c, h.Title, guser, nil)
					}
				}

				db.Save(user)
			}

			if !target.m {
				media := guser.Feed(nil)

				c.logger.Printf("Downloading feed media of %s (%d)\n", guser.Username, guser.ID)
				for media.Next() {
					// saving feed media in *outDir/{userid}/feed
					for _, item := range media.Items {
						_, err := downloadAndStoreFeed(db, &item, c, guser, nil)
						if err != nil {
							if err != errIsCarousel {
								log.Printf("error downloading feed media of %s: %s\n", guser.Username, err)
								continue
							}
						}
						user.MediaCount++
					}
				}
				db.Save(user)
			}
			if target.nm {
				user.MediaCount = user.Media
				db.Save(user)
			}
		}

		// getting new user strucure
		nguser, err := insta.Profiles.ByID(user.Id)
		if err != nil {
			log.Printf("error getting profile of %s (%d): %s\n", target.name, user.Id, err)
			continue
		}

		// checking all values.
		// up is used to update database values
		up := false
		if !target.p { // see profile
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
			if nguser.IsPrivate != user.IsPrivate {
				err = nguser.FriendShip()
				if err != nil {
					log.Println(err)
					goto end
				}
				msg := ""
				if nguser.IsPrivate {
					msg = "%s is private now. "
				} else {
					msg = "%s is public now. "
				}
				if !nguser.Friendship.Following {
					msg += "And you doesn't follow this user"
				}
				log.Printf(msg, nguser.Username)
				up = true
			}
		}
		if !target.f { // followers
			if n := nguser.FollowerCount - user.Followers; n != 0 {
				up = true
				if n > 0 {
					log.Printf("%s has %d new followers", user.Username, n)
				} else {
					n = (n ^ -1) + 1
					log.Printf("%s has %d unfollowers", user.Username, n)
				}
			}
		}
		if !target.w { // following
			if n := nguser.FollowingCount - user.Following; n != 0 {
				up = true
				if n > 0 {
					log.Printf("%s started following %d users", user.Username, n)
				} else {
					n = (n ^ -1) + 1
					log.Printf("%s stopped following %d users", user.Username, n)
				}
			}
		}

		if !target.s { // stories
			stories := nguser.Stories()
			for stories.Next() {
			itemLoop:
				for _, item := range stories.Items {
					story := &Stories{}
					copyItemToStory(&item, story)
					if db.Where(story).Find(story).Error == nil { // exists
						continue itemLoop
					}

					v, err := downloadAndStoreStory(db, &item, c, "", nguser, story)
					if err != nil {
						log.Printf("error downloading story: %s", err)
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
		}

		// TODO: check deleted values
		if !target.m { // media
			n := user.MediaCount
			if target.nm || user.IsPrivate {
				n = user.Media
			}
			// user.MediaCount is the number of feed items in database
			if n = nguser.MediaCount - n; n != 0 {
				up = true
				if n > 0 {
					log.Printf("%s has %d new medias", user.Username, n)
				} else {
					n = (n ^ -1) + 1
					log.Printf("%s deleted %d medias", user.Username, n)
					user.MediaCount -= n
					goto highLoop
				}

				gfeed := nguser.Feed(nil)
				for gfeed.Next() {
				gitemLoop:
					for _, item := range gfeed.Items {
						feed := &Feed{}
						copyItemToFeed(&item, feed)
						if db.Where(feed).Find(feed).Error == nil { // exists
							continue gitemLoop
						}

						v, err := downloadAndStoreFeed(db, &item, c, nguser, feed)
						if err != nil {
							if err != errIsCarousel {
								log.Printf("error downloading feed: %s\n", err)
								continue gitemLoop
							}
						}
						user.MediaCount++

						if v {
							c.SendVideo(fmt.Sprintf("New media of %s\n %s", nguser.Username, item.Caption.Text), feed.Path)
						} else {
							c.SendPhoto(fmt.Sprintf("New media of %s\n %s", nguser.Username, item.Caption.Text), feed.Path)
						}
						c.logger.Printf("Downloaded in %s\n", feed.Path)
					}
					if nguser.MediaCount <= user.MediaCount {
						break
					}
				}
			}
		}
	highLoop:
		if !target.h { // highlights
			// downloading user highlights
			hlgts, err := nguser.Highlights()
			if err != nil {
				log.Printf("error downloading %s highlights: %s", nguser.Username, err)
				goto end
			}
			if n := len(hlgts) - user.Highlights; n != 0 {
				user.Highlights += n
				if n > 0 {
					log.Printf("%s has %d new highlights\n", nguser.Username, n)
				} else {
					n = (n ^ -1) + 1
					log.Printf("%s deleted %d highlights\n", nguser.Username, n)
				}
				c.logger.Printf("Downloading highlights of %s (%d)\n", nguser.Username, nguser.ID)
				for _, h := range hlgts {
					for _, item := range h.Items {
						downloadAndStoreStory(db, &item, c, h.Title, nguser, nil)
					}
				}
			}
		}

	end:
		if up {
			copyGuserToUser(nguser, user)
			err = db.Save(user).Error
			if err != nil {
				log.Println("error updating database", err)
			}
		}
	}
}

func downloadAndStoreFeed(db *gorm.DB, item *goinsta.Item, c *nConn, guser *goinsta.User, feed *Feed) (bool, error) {
	v, output := false, path.Join(*outDir, strconv.FormatInt(guser.ID, 10), "feed")
	imgs, vds, err := item.Download(output, "")
	if err != nil {
		if len(item.CarouselMedia) == 0 {
			return false, err
		}
		for i := range item.CarouselMedia {
			v, err = downloadAndStoreFeed(db, &item.CarouselMedia[i], c, guser, feed)
			if err != nil {
				return v, fmt.Errorf("carousel media %d: %s", item.CarouselMedia[i].Pk, err)
			}
		}
		return v, errIsCarousel
	}
	if feed == nil {
		feed = &Feed{}
	}

	v, feed.Path, feed.Url = false, imgs, goinsta.GetBest(item.Images.Versions)
	if vds != "" {
		v, feed.Path, feed.Url = true, vds, goinsta.GetBest(item.Videos)
	}

	copyItemToFeed(item, feed)
	err = db.Save(feed).Error
	if err != nil {
		return v, fmt.Errorf("failed creating feed: %s\n", err)
	}
	c.logger.Printf("Downloaded in %s\n", feed.Path)
	return v, nil
}

func downloadAndStoreStory(db *gorm.DB, item *goinsta.Item, c *nConn, title string, guser *goinsta.User, story *Stories) (bool, error) {
	v, output := false, path.Join(*outDir, strconv.FormatInt(guser.ID, 10), "highlights", title)
	imgs, vds, err := item.Download(output, "")
	if err != nil {
		return false, fmt.Errorf("error downloading item %d: %s\n", item.Pk, err)
	}
	if story == nil {
		story = &Stories{}
	}
	story.Title = title
	story.Path, story.Url = imgs, goinsta.GetBest(item.Images.Versions)
	if vds != "" {
		// if item is a video is not an image. (xd)
		v, story.Path, story.Url = true, vds, goinsta.GetBest(item.Videos)
	}
	if title != "" {
		story.Highlight = true
	}
	copyItemToStory(item, story)

	// update actual value if exists
	db.Save(story)
	c.logger.Printf("Downloaded in %s\n", story.Path)
	return v, nil
}
