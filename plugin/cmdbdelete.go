package plugin

import (
	"context"
	"haha/job"
	"time"
)

type cmdbdelete struct{}

func (p *cmdbdelete) Conf(b []byte) error { return nil }

func (p *cmdbdelete) Run(ctx context.Context, result chan string) error {
	result <- "wait cmdb delete machine"
	time.Sleep(time.Second * 10)
	result <- "wait cmdb delete machine end"
	return nil
}

func init() {
	job.Register("cmdbdelete", func() job.Pluginer { return &cmdbdelete{} })
}
