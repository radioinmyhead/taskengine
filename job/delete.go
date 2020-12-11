package job

//type machinedelete struct {
//	*dbjob
//}
//
//func NewMachinedelete() *machinedelete {
//	ret := &machinedelete{&dbjob{}}
//	ret.dbjob.Jobid = bson.NewObjectId()
//	for _, k := range []string{"callapi", "cmdbdelete"} {
//		ret.dbjob.Jobtask = append(ret.dbjob.Jobtask, jobtask{Name: k})
//	}
//	ret.dbjob.Jobname = ret.Name()
//	return ret
//}
//
//func (j *machinedelete) Name() string {
//	return "machinedelete"
//}
//
//func (j *machinedelete) Run() error {
//	logrus.Info("delete machine start")
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
//		if e := j.SetCost(pluginname, stime, etime); e != nil {
//			return e
//		}
//		if e := j.Endplugin(pluginname, err); e != nil {
//			return e
//		}
//		if err != nil {
//			return err
//		}
//	}
//	logrus.Info("delete machine end")
//	return nil
//}
//
//func init() {
//	register(func() jobber { return NewMachinedelete() })
//}
//
