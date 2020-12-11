package plugin

import "fmt"

type pluginer interface {
	Run() error
}

var allplugin map[string]func() pluginer

func register(name string, p func() pluginer) {
	if allplugin == nil {
		allplugin = make(map[string]func() pluginer)
	}
	allplugin[name] = p
}

func GetPlugin(name string) (func() pluginer, error) {
	p, ok := allplugin[name]
	if ok {
		return p, nil
	}
	return nil, fmt.Errorf("not plugin name=%v", name)
}
