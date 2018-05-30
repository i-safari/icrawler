package main

import (
	"github.com/ahmdrz/goinsta"
)

type User struct {
	Id         int64  `gorm:"primary_key"`
	Username   string `gorm:"type:varchar(32);not null"`
	FullName   string `gorm:"type:varchar(64)"`
	Biography  string `gorm:"size:255"`
	Email      string `gorm:"type:varchar(64)"`
	Phone      string `gorm:"type:varchar(20)"`
	Media      int
	Followers  int
	Following  int
	Highlights int
	IsPrivate  bool
}

func copyGuserToUser(guser *goinsta.User, user *User) {
	user.Id = guser.ID
	user.Username = guser.Username
	user.FullName = guser.FullName
	user.Biography = guser.Biography
	user.Email = guser.PublicEmail
	user.Phone = guser.PublicPhoneNumber
	user.Media = guser.MediaCount
	user.Followers = guser.FollowerCount
	user.Following = guser.FollowingCount
	user.IsPrivate = guser.IsPrivate
}

type Feed struct {
	//User     User   `gorm:"foreignkey:UserId"`
	UserId   int64  `gorm:"primary_key" sql:"TYPE:int not null REFERENCES users(id)"`
	Id       int64  `gorm:"primary_key" sql:"TYPE:int not null"`
	Username string `gorm:"type:varchar(32);NOT NULL"`
	Path     string `gorm:"size:2048"`
	Url      string `gorm:"size:2048"`
	Deleted  bool
}

func copyItemToFeed(item *goinsta.Item, feed *Feed) {
	feed.UserId = item.User.ID
	feed.Id = item.Pk
	feed.Username = item.User.Username
}

type Stories struct {
	Feed
	Title     string `gorm:"varchar(32)"`
	Highlight bool
}

func copyItemToStory(item *goinsta.Item, story *Stories) {
	story.Id = item.Pk
	story.UserId = item.User.ID
	story.Username = item.User.Username
}
