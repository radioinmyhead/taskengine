package plugin

import "haha/job"

func init() {
	job.Register("machinewait", func() job.Pluginer { return &machinewait{} })
	job.Register("cmdbdelete", func() job.Pluginer { return &cmdbdelete{} })
	job.Register("cloudinit", func() job.Pluginer { return &cloudinit{} })
	job.Register("callapi", func() job.Pluginer { return &call{} })
}
