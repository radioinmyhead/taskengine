package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"haha/job"
	"time"

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
	cm.Base.Tasklist = []string{"machinecreate-callapi", "machinecreate-init", "machinecreate-check"}
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

func (cm *CreateMachine) Conf(data []byte) (err error) {
	tmp, err := NewCreateMachineFromJson(data)
	if err != nil {
		return
	}
	*cm = *tmp
	return nil
}

func (cm *CreateMachine) Step1(ctx context.Context, result chan string) (err error) {
	result <- "call plugin api"
	time.Sleep(time.Second)
	err = cm.SetIDs([]string{"A", "B", "C"})
	result <- "call plugin api end"
	fmt.Println("call api", err)
	return
}

func (cm *CreateMachine) Step2(ctx context.Context, result chan string) (err error) {
	fmt.Println("call init")

	list := []*Cloudinit{}
	for _, ip := range cm.IDs {
		ci := NewCloudinit(ip)
		err = ci.Insert()
		if err != nil {
			return
		}
		err = job.NewDbjob("cloudinit", []string{"Cloudinit"}, "cloudinit", ci.ID).Start()
		if err != nil {
			return
		}
		list = append(list, ci)
	}
	for {
		i := 0
		for _, ci := range list {
			finish, err := job.CheckFinishbyContextid(ci.ID)
			if err != nil {
				return err
			}
			if !finish {
				i++
				continue
			}
		}
		if i > 0 {
			time.Sleep(time.Second)
		} else {
			break
		}
	}
	fmt.Println("call init end")
	return nil
}

func (cm *CreateMachine) Step3(ctx context.Context, result chan string) (err error) {
	return nil
}

type CreateOrderS1 struct{ CreateMachine }

func (s *CreateOrderS1) Run(ctx context.Context, result chan string) (err error) {
	return s.Step1(ctx, result)
}

func init() {
	job.Register("machinecreate-callapi", func() job.Pluginer { return &CreateOrderS1{} })
}

type CreateOrderS2 struct{ CreateMachine }

func (s *CreateOrderS2) Run(ctx context.Context, result chan string) (err error) {
	return s.Step2(ctx, result)
}
func init() {
	job.Register("machinecreate-init", func() job.Pluginer { return &CreateOrderS2{} })
}

type CreateOrderS3 struct{ CreateMachine }

func (s *CreateOrderS3) Run(ctx context.Context, result chan string) (err error) {
	return s.Step3(ctx, result)
}
func init() {
	job.Register("machinecreate-check", func() job.Pluginer { return &CreateOrderS3{} })
}
