package plugin

import (
	"time"

	"github.com/sirupsen/logrus"
)

type call struct{}

func (p *call) Run() error {
	logrus.Info("call plugin api")
	time.Sleep(time.Second)
	logrus.Info("call plugin api end")
	return nil
}

func init() {
	register("callapi", func() pluginer { return &call{} })
}
