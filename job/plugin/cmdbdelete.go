package plugin

import (
	"time"

	"github.com/sirupsen/logrus"
)

type cmdbdelete struct{}

func (p *cmdbdelete) Run() error {
	logrus.Info("wait cmdb delete machine")
	time.Sleep(time.Second * 10)
	logrus.Info("wait cmdb delete machine end")
	return nil
}

func init() {
	register("cmdbdelete", func() pluginer { return &cmdbdelete{} })
}
