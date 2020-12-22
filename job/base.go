package job

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/sirupsen/logrus"
)

// db
var db *mgo.Database

func SetDb(d *mgo.Database) {
	db = d
}

var timeout = time.Second

type jobtask struct {
	Name   string
	Status string
	Err    string
	Stime  time.Time
	Etime  time.Time
	Logs   []string
}

type jobcontext struct {
	Col string
	ID  bson.ObjectId `bson:"_id"`
}

func (j jobcontext) get() (b []byte, err error) {
	var data interface{}
	err = db.C(j.Col).FindId(j.ID).Sort("-_id").Limit(1).One(&data)
	if err != nil {
		return
	}
	b, err = json.Marshal(data)
	return
}

type dbjob struct {
	Jobid   bson.ObjectId `bson:"_id"`
	Jobname string
	Context jobcontext
	Jobtask []jobtask
	Heart   time.Time
	Status  string
	Err     string
	Stime   time.Time
	Etime   time.Time
	ticker  *time.Ticker `bson:"-"` // 只有这一个是运行状态，其他的都db状态。
}

func NewDbjob(name string, list []string, table string, id bson.ObjectId) *dbjob {
	ret := &dbjob{}
	ret.Jobid = bson.NewObjectId()
	ret.Stime = time.Now()
	ret.Jobname = name
	ret.Context.Col = table
	ret.Context.ID = id
	for _, k := range list {
		ret.Jobtask = append(ret.Jobtask, jobtask{Name: k})
	}
	return ret
}

func (j *dbjob) upsert() (err error) { // upsert
	query := bson.M{"_id": j.Jobid}
	_, err = db.C("job").Upsert(query, j)
	return
}

func (j *dbjob) checktimeout() bool {
	if j.Heart.IsZero() {
		return true
	}
	return time.Now().Sub(j.Heart) > timeout*10
}

func (j *dbjob) lock() error {
	query := bson.M{"_id": j.Jobid, "heart": j.Heart}
	j.Heart = time.Now()
	set := bson.M{"$set": bson.M{"heart": j.Heart}}

	logrus.Info("update heartbeat", j.Heart)
	err := db.C("job").Update(query, set)
	if err != nil {
		logrus.Info("update heart failed", err.Error())
		return err
	}
	return nil
}

func (j *dbjob) Start() (err error) {
	err = j.upsert()
	if err != nil {
		return
	}

	if !j.checktimeout() {
		return fmt.Errorf("lock failed")
	}

	err = j.lock()
	if err != nil {
		return
	}

	go func() {
		j.ticker = time.NewTicker(timeout * 1)
		for range j.ticker.C {
			if j.lock() != nil {
				j.lock()
			}
		}
	}()
	go func() {
		err = j.run()
		e := j.end(err)
		if e != nil {
			logrus.Info(e)
		}
	}()

	return nil
}

func (j *dbjob) end(ret error) error {
	// stop time ticker
	j.ticker.Stop()

	// set values
	if ret == nil {
		j.Status = "succ"
	} else {
		j.Status = "fail"
	}
	j.Err = fmt.Sprintf("%v", ret)

	// update
	err := db.C("job").Update(bson.M{"_id": j.Jobid}, bson.M{"$set": bson.M{
		"err":    j.Err,
		"status": j.Status,
		"etime":  time.Now(),
	}})
	if err != nil {
		logrus.Info("db write failed", err)
		return err
	}
	return nil
}

func (j *dbjob) pluginSetStart(name string) error {
	start := time.Now()
	for i, task := range j.Jobtask {
		if task.Name == name {
			j.Jobtask[i].Stime = start
		}
	}
	query := bson.M{"_id": j.Jobid, "jobtask.name": name}
	set := bson.M{"$set": bson.M{
		"jobtask.$.stime": start,
	}}
	err := db.C("job").Update(query, set)
	return err
}

func (j *dbjob) pluginAppendlog(name string) (ch chan string) {
	ch = make(chan string)
	go func() {
		for log := range ch {
			log = fmt.Sprintf("%s: %s", time.Now().Format(time.RFC3339), log)
			query := bson.M{"_id": j.Jobid, "jobtask.name": name}
			set := bson.M{"$push": bson.M{
				"jobtask.$.logs": log,
			}}
			if err := db.C("job").Update(query, set); err != nil {
				logrus.Info(err)
				// TODO, db error
			}
		}
	}()
	return
}

func (j *dbjob) pluginEnd(name string, ret error) error {
	end := time.Now()
	status := "succ"
	if ret != nil {
		status = "fail"
	}
	query := bson.M{"_id": j.Jobid, "jobtask.name": name}
	set := bson.M{"$set": bson.M{
		"jobtask.$.err":    fmt.Sprintf("%v", ret),
		"jobtask.$.status": status,
		"jobtask.$.etime":  end,
	}}
	err := db.C("job").Update(query, set)
	return err
}

func (j *dbjob) run() error {
	logrus.Info("start:", j.Jobname)
	ctx := context.Background()
	for _, task := range j.Jobtask {
		// jump the runned task
		if task.Status != "" {
			logrus.Info("conti:", j.Jobname, ":", task.Name)
			continue
		}

		logrus.Info("start:", j.Jobname, ":", task.Name)

		// get result channel
		result := j.pluginAppendlog(task.Name)
		defer close(result)

		// get args
		args, err := j.Context.get()
		if err != nil {
			return err
		}
		logrus.Info("get args:", string(args))

		// get plugin
		pfac, err := getPlugin(task.Name)
		if err != nil {
			return err
		}
		p := pfac()

		// config plugin
		err = p.Conf(args)
		if err != nil {
			return err
		}

		// run plugin
		j.pluginSetStart(task.Name)
		err = p.Run(ctx, result)

		// set plugin end-result to db
		dberr := j.pluginEnd(task.Name, err)
		if dberr != nil {
			logrus.Info(dberr)
			return dberr
		}
		if err != nil {
			return err
		}
	}
	fmt.Println(4)
	logrus.Info("  end:", j.Jobname)
	return nil
}

func GetRunningJobs() (list []*dbjob, err error) {
	query := bson.M{"status": ""}
	list = []*dbjob{}
	err = db.C("job").Find(query).All(&list)
	if err == mgo.ErrNotFound {
		err = nil
	}
	return
}

func ContinueJobs() (err error) {
	list, err := GetRunningJobs()
	if err != nil {
		return
	}
	for _, one := range list {
		if !one.checktimeout() {
			continue
		}
		err = one.Start()
		if err != nil {
			return
		}
	}
	return
}
