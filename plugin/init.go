package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"haha/job"
	"time"

	"github.com/globalsign/mgo/bson"
)

type Cloudinit struct {
	ID bson.ObjectId `bson:"_id" json:"_id"`
	IP string
}

func NewCloudinit(ip string) *Cloudinit {
	ret := &Cloudinit{
		ID: bson.NewObjectId(),
		IP: ip,
	}
	return ret
}

func (p *Cloudinit) Insert() error {
	return db.C("cloudinit").Insert(p)
}

func (p *Cloudinit) Conf(b []byte) error {
	return json.Unmarshal(b, p)
}

func (p *Cloudinit) Run(ctx context.Context, result chan string) error {
	result <- fmt.Sprintf("run cloud init ip=%v", p.IP)
	time.Sleep(time.Second * 3)
	result <- fmt.Sprintf("cloud init end ip=%v", p.IP)
	return nil
}

func init() {
	job.Register("Cloudinit", func() job.Pluginer { return &Cloudinit{} })
}
