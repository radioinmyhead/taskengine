package plugin

import (
	"context"
	"time"
)

type cloudinit struct{}

func (p *cloudinit) Conf(b []byte) error { return nil }

func (p *cloudinit) Run(ctx context.Context, result chan string) error {
	result <- "run cloud init"
	time.Sleep(time.Second * 3)
	result <- "run cloud init end"
	return nil
}

//func init() {
//	register("cloudinit", func() pluginer { return &cloudinit{} })
//}
