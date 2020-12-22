package plugin

import (
	"context"
	"fmt"
	"haha/job"
	"time"
)

type installpackage struct {
	//order *model.CreateMachine
}

func (p *installpackage) Conf(b []byte) (err error) {
	//tmp := &model.CreateMachine{}
	//err = json.Unmarshal(b, tmp)
	//if err != nil {
	//	return
	//}
	//p.order = tmp
	return
}

func (p *installpackage) Run(ctx context.Context, result chan string) (err error) {
	result <- "call plugin api"
	//result <- fmt.Sprintf("hello %s", p.order.Oper)
	time.Sleep(time.Second)
	//err = p.order.SetIDs([]string{"A", "B", "C"})
	//result <- "call plugin api end"
	fmt.Println("call api", err)
	return
}

func init() {
	job.Register("installpackage", func() job.Pluginer { return &installpackage{} })
}
