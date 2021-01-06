package job

import (
	"context"
	"fmt"
	"reflect"
	"runtime"

	"github.com/globalsign/mgo/bson"
)

// RealRunner the real runner
type RealRunner func(ctx context.Context, result chan string) (err error)

// PluginRunner the runer
type PluginRunner interface {
	Run(ctx context.Context, action string, result chan string) error
	Endwith(error) error
}

// RunnerFactory get runner
type RunnerFactory func(bson.ObjectId) (PluginRunner, error)

// hash of db.col and order
var allplugin map[string]RunnerFactory

func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
func getStructType(i interface{}) string {
	t := reflect.TypeOf(i)
	if t.Kind() == reflect.Ptr {
		return t.Elem().Name()
	}
	return t.Name()
}

// Registerplugin save all Runfunc to map
func Registerplugin(col string, p RunnerFactory) {
	if allplugin == nil {
		allplugin = make(map[string]RunnerFactory)
	}
	if _, ok := allplugin[col]; ok {
		panic("the plugin name conflict")
	}
	allplugin[col] = p
}

func factoryPlugin(col string) (p RunnerFactory, err error) {
	p, ok := allplugin[col]
	if !ok {
		err = fmt.Errorf("not plugin name=%v", col)
	}
	return
}
