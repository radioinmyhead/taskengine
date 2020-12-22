package db

import (
	"haha/job"
	"haha/plugin"
	"time"

	"github.com/globalsign/mgo"
)

// conf
var dburi string
var dbname string

// globle
var session *mgo.Session

//var db *mgo.Database

func Init() {
	dburi = "mongodb://192.168.236.169/"
	dbname = "haha"

	session, err := mgo.DialWithTimeout(dburi+dbname, time.Second)
	if err != nil {
		panic(err)
	}
	session.SetMode(mgo.Monotonic, true)
	//db = session.DB(dbname)

	job.SetDb(session.DB(dbname))
	plugin.SetDB(session.DB(dbname))
}

//func C(name string) *mgo.Collation {
//	return session.DB(dbname).C(name)
//}
