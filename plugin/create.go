package plugin

import (
	"context"
	"fmt"
	"haha/job"
	"time"

	"github.com/globalsign/mgo/bson"
	"github.com/sirupsen/logrus"
)

type process struct {
	Machineid string
	Initid    string
	Status    string
}

// MachineCreate the order of MachineCreate
type MachineCreate struct {
	ID      bson.ObjectId `bson:"_id"`              // db id
	Oper    string        `bson:"oper" form:"oper"` // args:oper
	Plan    string        `bson:"plan" form:"plan"` // args:plan
	Num     int           `bson:"num" form:"num"`   // args:num
	Process []process     // process
	Status  string        // order status
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
	list := []process{}
	for _, id := range ids {
		i := process{
			Machineid: id,
		}
		list = append(list, i)
	}
	cm.Process = list

	set := bson.M{"$set": bson.M{"process": cm.Process}}
	err = db.C(DBMachineCreate).UpdateId(cm.ID, set)
	return
}
func (cm *MachineCreate) SetProcess(p process) (err error) {
	fmt.Println("bef:", cm.Process)
	for i, process := range cm.Process {
		if process.Machineid == p.Machineid {
			cm.Process[i] = p
		}
	}
	fmt.Println("aft:", cm.Process)
	set := bson.M{"$set": bson.M{"process": cm.Process}}
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
	for _, process := range cm.Process {
		if process.Initid == "" {
			ip := process.Machineid
			ci := NewCloudinit(ip)
			if dberr := ci.Upsert(); dberr != nil {
				return dberr
			}
			err = job.NewDbjob("cloudinit", []string{"machine_init_create", "machine_init_installpackage", "machine_init_reboot"},
				DBMachineInit, ci.ID).Start()
			if err != nil {
				return err
			}
			process.Initid = ci.ID.Hex()
			if dberr := cm.SetProcess(process); dberr != nil {
				return dberr
			}
		} else {
			logrus.Info("alread start init")
		}
	}
	fmt.Println("in cloud init, new process", cm.Process)
	for _, process := range cm.Process {
		if process.Status != "" {
			process.Status = ""
			if dberr := cm.SetProcess(process); dberr != nil {
				return dberr
			}
		}
	}
	fmt.Println("in cloud init, start to wait", cm.Process)
	for i := 0; i < 300; i++ {
		count := 0
		for _, process := range cm.Process {
			if process.Status != "" {
				continue
			}
			fmt.Println("check process id=", process.Machineid)
			finish, err := job.CheckFinishbyContextid(bson.ObjectIdHex(process.Initid))
			if err != nil {
				fmt.Println("check process id=", process.Machineid)
				return err
			}
			if !finish {
				fmt.Println("check process id=", process.Machineid, "not finish")
				count++
				continue
			}

			fmt.Println("find finish", process)
			succ, err := job.CheckSuccbyContextid(bson.ObjectIdHex(process.Initid))
			if err != nil {
				return err
			}
			if succ {
				process.Status = "succ"
			} else {
				process.Status = "fail"
			}
			fmt.Println("set status data=", process)
			if dberr := cm.SetProcess(process); dberr != nil {
				return dberr
			}
		}
		if count > 0 {
			time.Sleep(time.Second)
		}
	}

	failnum := 0
	for _, process := range cm.Process {
		if process.Status == "fail" {
			failnum++
		}
	}
	if failnum > 0 {
		err = fmt.Errorf("cloud init failed,failed")
		return

	}
	return nil
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
