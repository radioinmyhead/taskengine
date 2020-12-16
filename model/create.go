package model

import (
	"encoding/json"

	"github.com/globalsign/mgo/bson"
)

type CreateMachine struct {
	Base `json:"base"`
	ID   bson.ObjectId `json:"_id" bson:"_id"`               // db id
	Oper string        `json:"oper" bson:"oper" form:"oper"` // args:oper
	Plan string        `json:"plan" bson:"plan" form:"plan"` // args:plan
	Num  int           `json:"num" bson:"num" form:"num"`    // args:num
	IDs  []string      `json:"ids" bson:"ids"`               // result
}

func NewCreateMachine(oper, plan string, num int) (ret *CreateMachine, err error) {
	ret = &CreateMachine{
		ID:   bson.NewObjectId(),
		Oper: oper,
		Plan: plan,
		Num:  num,
	}
	ret.Init()
	return
}
func NewCreateMachineFromJson(b []byte) (ret *CreateMachine, err error) {
	ret = &CreateMachine{}
	err = json.Unmarshal(b, ret)
	if err != nil {
		return nil, err
	}
	ret.Init()
	return ret, err
}

func (cm *CreateMachine) Init() {
	cm.Base.Col = string(DBMachineCreate)
	cm.Base.Jobname = "machine-create"
	cm.Base.Tasklist = []string{"callapi", "machinewait", "cloudinit"}
}

func (cm *CreateMachine) Insert() (err error) {
	err = db.C(string(DBMachineCreate)).Insert(cm)
	return
}

func (cm *CreateMachine) SetIDs(ids []string) (err error) {
	cm.IDs = ids
	set := bson.M{"$set": bson.M{"ids": ids}}
	err = db.C(string(DBMachineCreate)).UpdateId(cm.ID, set)
	return
}
