package plugin

import "github.com/globalsign/mgo"

var db *mgo.Database
var setdblist []func()

func SetDB(d *mgo.Database) {
	db = d
	for _, fn := range setdblist {
		fn()
	}
}

const (
	DBMachineCreate string = "machinecreate"
	DBMachineInit   string = "cloudinit"
)
