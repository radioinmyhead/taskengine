package plugin

import "github.com/globalsign/mgo"

var db *mgo.Database

func SetDB(d *mgo.Database) {
	db = d
}

type DbCollection string

const (
	DBMachineCreate DbCollection = "machinecreate"
	DBMachineInit   DbCollection = "cloudinit"
)
