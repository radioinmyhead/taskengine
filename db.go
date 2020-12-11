package main

import (
	"haha/job"
	"time"

	"gopkg.in/mgo.v2"
)

var db *mgo.Database

func init() {
	uri := "mongodb://192.168.236.169/"
	name := "haha"
	session, err := mgo.DialWithTimeout(uri+name, time.Second)
	if err != nil {
		panic(err)
	}
	session.SetMode(mgo.Monotonic, true)
	db = session.DB(name)

	job.SetDB(db)
}
