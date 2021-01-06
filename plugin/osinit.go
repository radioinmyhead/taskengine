package plugin

import (
	"context"
	"fmt"
	"haha/job"
	"time"

	"github.com/globalsign/mgo/bson"
)

type Cloudinit struct {
	ID     bson.ObjectId `bson:"_id"`
	IP     string
	Status string
}

func NewCloudinit(ip string) *Cloudinit {
	ret := &Cloudinit{
		ID: bson.NewObjectId(),
		IP: ip,
	}
	return ret
}

func FindCloudinit(id bson.ObjectId) (job.PluginRunner, error) {
	ret := &Cloudinit{}
	err := db.C(string(DBMachineInit)).FindId(id).One(ret)
	return ret, err
}

func (p *Cloudinit) Insert() error {
	return db.C(string(DBMachineInit)).Insert(p)
}

func (p *Cloudinit) Endwith(ret error) (err error) {
	p.Status = "fail"
	if ret == nil {
		p.Status = "succ"
	}
	set := bson.M{"$set": bson.M{"status": p.Status}}
	err = db.C(string(DBMachineInit)).UpdateId(p.ID, set)
	return
}

func (p *Cloudinit) create(ctx context.Context, result chan string) error {
	result <- fmt.Sprintf("run cloud init ip=%v", p.IP)
	time.Sleep(time.Second * 3)
	result <- fmt.Sprintf("cloud init end ip=%v", p.IP)
	return nil
}

func (p *Cloudinit) installpackage(ctx context.Context, result chan string) (err error) {
	result <- "call plugin api"
	//result <- fmt.Sprintf("hello %s", p.order.Oper)
	time.Sleep(time.Second)
	result <- fmt.Sprint("in install package", p)
	//err = p.order.SetIDs([]string{"A", "B", "C"})
	result <- "call plugin api end"
	//fmt.Println("call api", err)
	return
}

func (p *Cloudinit) reboot(ctx context.Context, result chan string) error {
	result <- "wait machine to run"
	//panic("panic in wait")
	time.Sleep(time.Second)
	result <- fmt.Sprint("in install package", p)
	result <- "wait machine to run end"
	return fmt.Errorf("test wait failed")
}

func (p *Cloudinit) Run(ctx context.Context, action string, result chan string) (err error) {
	dic := map[string]job.RealRunner{
		"machine_init_create":         p.create,
		"machine_init_installpackage": p.installpackage,
		"machine_init_reboot":         p.reboot,
	}

	fn, ok := dic[action]
	if !ok {
		return fmt.Errorf("not support action=%v", action)

	}
	return fn(ctx, result)
}

func init() {
	job.Registerplugin(string(DBMachineInit), FindCloudinit)
}
