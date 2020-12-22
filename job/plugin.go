package job

import (
	"context"
	"fmt"
)

type Pluginer interface {
	Run(ctx context.Context, result chan string) error
	Conf([]byte) error
}

var allplugin map[string]func() Pluginer

func Register(name string, p func() Pluginer) {
	if allplugin == nil {
		allplugin = make(map[string]func() Pluginer)
	}
	allplugin[name] = p
}

func getPlugin(name string) (func() Pluginer, error) {
	p, ok := allplugin[name]
	if ok {
		return p, nil
	}
	return nil, fmt.Errorf("not plugin name=%v", name)
}
