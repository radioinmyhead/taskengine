package plugin

import (
	"time"

	"github.com/sirupsen/logrus"
)

type cloudinit struct{}

func (p *cloudinit) Run() error {
	logrus.Info("run cloud init")
	time.Sleep(time.Second * 3)
	logrus.Info("run cloud init end")
	return nil
}

func init() {
	register("cloudinit", func() pluginer { return &cloudinit{} })
}
