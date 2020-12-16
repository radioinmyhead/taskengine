package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"haha/model"
	"time"
)

type call struct {
	order *model.CreateMachine
}

func (p *call) Conf(b []byte) (err error) {
	tmp := &model.CreateMachine{}
	err = json.Unmarshal(b, tmp)
	if err != nil {
		return
	}
	p.order = tmp
	return
}

func (p *call) Run(ctx context.Context, result chan string) (err error) {
	result <- "call plugin api"
	result <- fmt.Sprintf("hello %s", p.order.Oper)
	time.Sleep(time.Second)
	err = p.order.SetIDs([]string{"A", "B", "C"})
	result <- "call plugin api end"
	fmt.Println("call api", err)
	return
}

func init() {
	register("callapi", func() pluginer { return &call{} })
}
