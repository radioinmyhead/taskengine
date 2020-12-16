package job

import (
	"fmt"
	"haha/job/plugin"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/sirupsen/logrus"
)

// db

var db *mgo.Database
var timeout time.Duration

func SetDB(d *mgo.Database) {
	db = d
	timeout = time.Second
}

type jobtask struct {
	Name   string
	Status string
	Err    string
	Stime  time.Time
	Etime  time.Time
	Cost   time.Duration
}

type dbjob struct {
	Jobid   bson.ObjectId `bson:"_id"`
	Jobname string
	Jobtask []jobtask
	Heart   time.Time
	Status  string
	Err     string
	Stime   time.Time
	Etime   time.Time
	Cost    time.Duration
	ticker  *time.Ticker `bson:"-"` // 只有这一个是运行状态，其他的都db状态。
}

func NewDbjob(name string, list []string) *dbjob {
	ret := &dbjob{}
	ret.Jobid = bson.NewObjectId()
	ret.Stime = time.Now()
	ret.Jobname = name
	for _, k := range list {
		ret.Jobtask = append(ret.Jobtask, jobtask{Name: k})
	}
	return ret
}

func (j *dbjob) upsert() (err error) { // upsert
	query := bson.M{"_id": j.Jobid}
	info, err := db.C("job").Upsert(query, j)
	logrus.Info(info)
	return
}

func (j *dbjob) checktimeout() bool {
	if j.Heart.IsZero() {
		return true
	}
	return time.Now().Sub(j.Heart) > timeout*6
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
		"cost":   time.Now().Sub(j.Stime),
	}})
	if err != nil {
		logrus.Info("db write failed", err)
		return err
	}
	return nil
}

func (j *dbjob) pluginEnd(name string, ret error, start, end time.Time) error {
	status := "succ"
	if ret != nil {
		status = "fail"
	}
	query := bson.M{"_id": j.Jobid, "jobtask.name": name}
	set := bson.M{"$set": bson.M{
		"jobtask.$.err":    fmt.Sprintf("%v", ret),
		"jobtask.$.status": status,
		"jobtask.$.stime":  start,
		"jobtask.$.etime":  end,
		"jobtask.$.cost":   end.Sub(start),
	}}
	err := db.C("job").Update(query, set)
	return err
}

func (j *dbjob) run() error {
	logrus.Info("start:", j.Jobname)
	for _, task := range j.Jobtask {
		if task.Status != "" {
			logrus.Info("conti:", j.Jobname, ":", task.Name)
			continue
		}
		logrus.Info("start:", j.Jobname, ":", task.Name)
		pfac, err := plugin.GetPlugin(task.Name)
		if err != nil {
			return err
		}
		p := pfac()
		stime := time.Now()
		err = p.Run()
		etime := time.Now()

		dberr := j.pluginEnd(task.Name, err, stime, etime)
		if dberr != nil {
			logrus.Info(dberr)
			return dberr
		}
		if err != nil {
			return err
		}
	}
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
