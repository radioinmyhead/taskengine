package model

import "github.com/globalsign/mgo"

type DbCollection string

var db *mgo.Database

func SetDB(d *mgo.Database) {
	db = d
}

const (
	DBMachineCreate DbCollection = "machinecreate"
)

type Base struct {
	Col      string   `json:"col"`
	Jobname  string   `json:"jobname"`
	Tasklist []string `json:"tasklist"`
}
