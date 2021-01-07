package job

import (
	"context"
	"fmt"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/sirupsen/logrus"
)

const (
	timeoutDuration time.Duration = time.Second
	timeoutCount    int           = 6
	jobStatusSucc   string        = "succ"
	jobStatusFail   string        = "fail"
)

type jobtask struct {
	ID     bson.ObjectId `bson:"_id"`
	Name   string
	Status string
	Err    string
	Stime  time.Time
	Etime  time.Time
	Logs   []string
}

func newJobtask(name string) jobtask {
	return jobtask{ID: bson.NewObjectId(), Name: name}
}

type jobcontext struct {
	Col string
	ID  bson.ObjectId `bson:"_id"`
}

func CheckFinishbyContextid(id bson.ObjectId) (finish bool, err error) {
	query := bson.M{"context._id": id, "status": ""}
	n, err := db.C(dbtable).Find(query).Count()
	if err != nil {
		return
	}

	return n == 0, nil
}

func CheckSuccbyContextid(id bson.ObjectId) (succ bool, err error) {
	query := bson.M{"context._id": id, "status": "fail"}
	n, err := db.C(dbtable).Find(query).Count()
	if err != nil {
		return
	}
	return n == 0, nil
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
	force   bool         `bson:"-"` // 仅仅用于重试，避免重试中被其他worder抢去
}

// NewDbjob create a new job
func NewDbjob(name string, list []string, table string, id bson.ObjectId) *dbjob {
	ret := &dbjob{}
	ret.Jobid = bson.NewObjectId()
	ret.Stime = time.Now()
	ret.Jobname = name
	ret.Context.Col = table
	ret.Context.ID = id
	for _, k := range list {
		ret.Jobtask = append(ret.Jobtask, newJobtask(k))
	}
	return ret
}

func (j *dbjob) upsert() (err error) { // upsert
	query := bson.M{"_id": j.Jobid}
	_, err = db.C(dbtable).Upsert(query, j)
	return
}

func (j *dbjob) checktimeout() bool {
	if j.Heart.IsZero() {
		return true
	}
	return time.Now().Sub(j.Heart) > timeoutDuration*time.Duration(timeoutCount)
}

func (j *dbjob) lock() error {
	query := bson.M{"_id": j.Jobid, "heart": j.Heart}
	j.Heart = time.Now()
	set := bson.M{"$set": bson.M{"heart": j.Heart}}

	//logrus.Info("update heartbeat", j.Heart)
	err := db.C(dbtable).Update(query, set)
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

	if !j.force {
		if !j.checktimeout() {
			err = fmt.Errorf("Start job: lock failed: no timeout time=%v", j.Heart)
			return
		}
	}

	err = j.lock()
	if err != nil {
		return
	}

	go func() {
		j.ticker = time.NewTicker(timeoutDuration)
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
		j.Status = jobStatusSucc
	} else {
		j.Status = jobStatusFail
	}
	j.Err = fmt.Sprintf("%v", ret)

	// update
	err := db.C(dbtable).Update(bson.M{"_id": j.Jobid}, bson.M{"$set": bson.M{
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

func (j *dbjob) pluginSetStart(jobtaskid bson.ObjectId) error {
	start := time.Now()
	for i, task := range j.Jobtask {
		if task.ID.Hex() == jobtaskid.Hex() {
			j.Jobtask[i].Stime = start
		}
	}
	query := bson.M{"_id": j.Jobid, "jobtask._id": jobtaskid}
	set := bson.M{"$set": bson.M{
		"jobtask.$.stime": start,
	}}
	err := db.C(string(dbtable)).Update(query, set)
	return err
}

func (j *dbjob) pluginAppendlog(jobtaskid bson.ObjectId) (ch chan string) {
	ch = make(chan string)
	go func() {
		for log := range ch {
			log = fmt.Sprintf("%s: %s", time.Now().Format(time.RFC3339), log)
			query := bson.M{"_id": j.Jobid, "jobtask._id": jobtaskid}
			set := bson.M{"$push": bson.M{
				"jobtask.$.logs": log,
			}}
			if err := db.C(string(dbtable)).Update(query, set); err != nil {
				logrus.Info(err)
				// TODO, db error
			}
		}
	}()
	return
}

func (j *dbjob) pluginEnd(jobtaskid bson.ObjectId, ret error) error {
	end := time.Now()
	status := jobStatusSucc
	if ret != nil {
		status = jobStatusFail
	}
	query := bson.M{"_id": j.Jobid, "jobtask._id": jobtaskid}
	set := bson.M{"$set": bson.M{
		"jobtask.$.err":    fmt.Sprintf("%v", ret),
		"jobtask.$.status": status,
		"jobtask.$.etime":  end,
	}}
	err := db.C(dbtable).Update(query, set)
	return err
}

func (j *dbjob) run() error {
	ctx := context.Background()

	// get contesxt by col and id
	argsFactory, err := factoryPlugin(j.Context.Col)
	if err != nil {
		return err
	}
	args, err := argsFactory(j.Context.ID)
	if err != nil {
		return err
	}

	if dberr := args.Upsert(); dberr != nil {
		return dberr
	}

	logrus.Info("start:", j.Jobname, "\targs=", args)

	for _, task := range j.Jobtask {
		// jump the runned task
		if task.Status != "" {
			logrus.Info("conti:", j.Jobname, ":", task.Name)
			continue
		}

		logrus.Info("start:", j.Jobname, ":", task.Name, "\targs=", args)

		// get result channel
		result := j.pluginAppendlog(task.ID)
		defer close(result)

		// run plugin
		if dberr := j.pluginSetStart(task.ID); dberr != nil {
			logrus.Info("  end:", j.Jobname, "\terr=", dberr, "\targs=", args)
			return dberr
		}
		//err = pfac(args, ctx, result)
		err = args.Run(ctx, task.Name, result)

		// set plugin end-result to db
		if dberr := j.pluginEnd(task.ID, err); dberr != nil {
			logrus.Info("  end:", j.Jobname, "\terr=", dberr, "\targs=", args)
			return dberr
		}

		if err != nil {
			break
		}
	}
	// order endwith succ
	if dberr := args.Endwith(err); dberr != nil {
		logrus.Info("  end:", j.Jobname, "\terr=", err, "\tdberr=", dberr, "\targs=", args)
		return dberr
	}
	logrus.Info("  end:", j.Jobname, "\terr=", err, "\targs=", args)
	return err
}

// GetRunningJobs get all jobs in running
func GetRunningJobs() (list []*dbjob, err error) {
	query := bson.M{"status": ""}
	list = []*dbjob{}
	err = db.C("job").Find(query).All(&list)
	if err == mgo.ErrNotFound {
		err = nil
	}
	return
}

// ContinueJobs continue all failed jobs
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

// 更新，添加新的job，同时吧整个单子标记成完成
func (j *dbjob) inserttask(pos int, taskname string) (err error) {
	j.Err = ""
	j.Status = ""
	j.Heart = time.Now()
	task := newJobtask(taskname)
	set := bson.M{
		"$push": bson.M{
			"jobtask": bson.M{
				"$each":     []interface{}{task},
				"$position": pos,
			},
		},
		"$set": bson.M{
			"heart":  j.Heart,
			"err":    j.Err,
			"status": j.Status,
		},
	}
	err = db.C(dbtable).UpdateId(j.Jobid, set)
	if err != nil {
		return
	}
	err = db.C(dbtable).FindId(j.Jobid).One(j)
	return
}

func (j *dbjob) retry() (err error) {
	// get failed, insert db , clean the job status, continue the job
	for i := len(j.Jobtask) - 1; i >= 0; i-- {
		task := j.Jobtask[i]
		if task.Status == "" {
			continue
		} else if task.Status == jobStatusFail {
			return j.inserttask(i+1, task.Name)
		} else {
			return fmt.Errorf("something wrong")
		}
	}
	return fmt.Errorf("not found failed task")
}

func RetryJob(id string) (err error) {
	j := &dbjob{}
	query := bson.M{
		"_id":    bson.ObjectIdHex(id),
		"status": jobStatusFail,
	}
	err = db.C(dbtable).Find(query).One(j)
	if err != nil {
		err = fmt.Errorf("RetryJob find job failed: id error or job not failed: id=%v err=%v query=%v", id, err.Error(), query)
		return
	}

	// retry
	err = j.retry()
	if err != nil {
		err = fmt.Errorf("RetryJob retry job failed err=%v", err.Error())
		return
	}

	j.force = true
	return j.Start()
}

func JobFindAll() (list []*dbjob, err error) {
	list = []*dbjob{}
	err = db.C(dbtable).Find(nil).All(&list)
	return
}

func JobFindbyid(id string) (job *dbjob, err error) {
	job = &dbjob{}
	err = db.C(dbtable).FindId(bson.ObjectIdHex(id)).One(&job)
	return
}

func JobFindbyCtx(id string) (list []*dbjob, err error) {
	list = []*dbjob{}
	query := bson.M{"context._id": bson.ObjectIdHex(id)}
	err = db.C(dbtable).Find(query).Limit(1).All(&list)
	return
}
