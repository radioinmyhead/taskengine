package plugin

import (
	"context"
	"fmt"
	"time"
)

type machinewait struct{}

func (p *machinewait) Conf(b []byte) error { return nil }

func (p *machinewait) Run(ctx context.Context, result chan string) error {
	result <- "wait machine to run"
	//panic("panic in wait")
	time.Sleep(time.Second)
	result <- "wait machine to run end"
	return fmt.Errorf("test wait failed")
}

func init() {
	register("machinewait", func() pluginer { return &machinewait{} })
}
