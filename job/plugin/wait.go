package plugin

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

type machinewait struct{}

func (p *machinewait) Run() error {
	logrus.Info("wait machine to run")
	//panic("panic in wait")
	time.Sleep(time.Second)
	logrus.Info("wait machine to run end")
	return fmt.Errorf("test wait failed")
}

func init() {
	register("machinewait", func() pluginer { return &machinewait{} })
}
