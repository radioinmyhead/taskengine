package plugin

import (
	"context"
	"fmt"
	"haha/job"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/sirupsen/logrus"
)

// MachineCreate the order of MachineCreate
type MachineCreate struct {
	ID     bson.ObjectId `bson:"_id"`              // db id
	Oper   string        `bson:"oper" form:"oper"` // args:oper
	Plan   string        `bson:"plan" form:"plan"` // args:plan
	Num    int           `bson:"num" form:"num"`   // args:num
	IDs    []string      `bson:"ids"`              // result
	Status string        // order status
}

// NewMachineCreate create machine order from httpargs
func NewMachineCreate(oper, plan string, num int) (ret *MachineCreate, err error) {
	ret = &MachineCreate{
		ID:   bson.NewObjectId(),
		Oper: oper,
		Plan: plan,
		Num:  num,
	}
	return
}

// FindMachineCreate get order from db by id
func FindMachineCreate(id bson.ObjectId) (job.PluginRunner, error) {
	ret := &MachineCreate{}
	err := db.C(DBMachineCreate).FindId(id).One(ret)
	return ret, err
}

func (cm *MachineCreate) Upsert() (err error) {
	_, err = db.C(DBMachineCreate).UpsertId(cm.ID, cm)
	return
}

func (cm *MachineCreate) SetIDs(ids []string) (err error) {
	cm.IDs = ids
	set := bson.M{"$set": bson.M{"ids": ids}}
	err = db.C(DBMachineCreate).UpdateId(cm.ID, set)
	return
}

func (cm *MachineCreate) Endwith(ret error) (err error) {
	cm.Status = "succ"
	if ret != nil {
		cm.Status = "fail"
	}
	set := bson.M{"$set": bson.M{"status": cm.Status}}
	err = db.C(DBMachineCreate).UpdateId(cm.ID, set)
	return
}

func (cm *MachineCreate) callapi(ctx context.Context, result chan string) (err error) {
	result <- "call plugin api"
	time.Sleep(time.Second)
	err = cm.SetIDs([]string{"A", "B", "C"})
	result <- "call plugin api end"
	//fmt.Println("call api", err)
	return
}

func (cm *MachineCreate) cloudinit(ctx context.Context, result chan string) (err error) {
	// TODO cant run multi times.
	logrus.Info("call init")
	list := []*Cloudinit{}
	for _, ip := range cm.IDs {
		ci := NewCloudinit(ip)
		err = ci.Upsert()
		if err != nil {
			return
		}
		err = job.NewDbjob("cloudinit", []string{"machine_init_create", "machine_init_installpackage", "machine_init_reboot"},
			string(DBMachineInit), ci.ID).Start()
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
				logrus.Info("call init: check finish: failed, err=", err)
				return err
			}
			if !finish {
				i++
				continue
			}
		}
		if i > 0 {
			logrus.Info("call init: check finish: not finish num=", i)
			time.Sleep(time.Second)
		} else {
			logrus.Info("call init: check finish: finish")
			break
		}
	}
	succ, err := job.CheckSuccbyContextid(list[0].ID)
	if err != nil {
		return
	}
	if !succ {
		err = fmt.Errorf("cloud init failed,failed")
	}

	return err
}

func (cm *MachineCreate) check(ctx context.Context, result chan string) (err error) {
	return nil
}

// Run cm.action(ctx context.Context, result chan string) (err error)
func (cm *MachineCreate) Run(ctx context.Context, action string, result chan string) (err error) {
	dic := map[string]job.RealRunner{
		"machine_create_callapi":   cm.callapi,
		"machine_create_cloudinit": cm.cloudinit,
		"machine_create_check":     cm.check,
	}

	fn, ok := dic[action]
	if !ok {
		return fmt.Errorf("not support action=%v", action)

	}
	return fn(ctx, result)
}

func init() {
	job.Registerplugin(DBMachineCreate, FindMachineCreate)
}
