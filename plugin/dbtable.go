package plugin

import "github.com/globalsign/mgo"

var db *mgo.Database

func SetDB(d *mgo.Database) {
	db = d
}

const (
	DBMachineCreate string = "machinecreate"
	DBMachineInit   string = "cloudinit"
)
