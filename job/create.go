package job

//type Create struct {
//	*dbjob
//}
//
//func NewCreate() *Create {
//	ret := &Create{&dbjob{}}
//	ret.dbjob.Jobid = bson.NewObjectId()
//	for _, k := range []string{"callapi", "machinewait", "cloudinit"} {
//		ret.dbjob.Jobtask = append(ret.dbjob.Jobtask, jobtask{Name: k})
//	}
//	ret.dbjob.Jobname = ret.Name()
//	return ret
//}
//
//func (j *Create) Name() string {
//	return "machine-create"
//}
//
//func (j *Create) Run() error {
//	logrus.Info("create machine start")
//	for _, jt := range j.dbjob.Jobtask {
//		pluginname := jt.Name
//		pfac, err := plugin.GetPlugin(pluginname)
//		if err != nil {
//			return err
//		}
//		p := pfac()
//		stime := time.Now()
//		err = p.Run()
//		etime := time.Now()
//
//		j.SetCost(pluginname, stime, etime)
//		if e := j.Endplugin(pluginname, err); e != nil {
//			return e
//		}
//		if err != nil {
//			return err
//		}
//	}
//	logrus.Info("create machine end")
//	return nil
//}
//
//func init() {
//	register("",func() jobber {
//		ret := &Create{&dbjob{}}
//		ret.dbjob.Jobid = bson.NewObjectId()
//		for _, k := range []string{"callapi", "machinewait", "cloudinit"} {
//			ret.dbjob.Jobtask = append(ret.dbjob.Jobtask, jobtask{Name: k})
//		}
//		ret.dbjob.Jobname = ret.Name()
//		return ret
//	})
//}
//
