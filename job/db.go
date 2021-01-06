package job

import (
	"github.com/globalsign/mgo"
)

var dbtable string = "job"

// db
var db *mgo.Database

// SetDb set db for package
func SetDb(d *mgo.Database) {
	db = d
}
