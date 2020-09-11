package coordinator

import (
	"sync"

	"github.com/myrteametrics/myrtea-sdk/v4/modeler"

	"go.uber.org/zap"
)

// Master is a singleton component which operates and coordinates Myrtea instances
type Master struct {
	Initialized bool
	Instances   map[string]*Instance
}

var instance *Master
var once sync.Once

// GetInstance returns a pointer to a singleton of the coordinator master component
// This singleton must be initialized at least one time with Initialize()
func GetInstance() *Master {
	once.Do(func() {
		instance = &Master{false, map[string]*Instance{}}
	})
	return instance
}

// InitInstance initialize the a new instance
func (master *Master) InitInstance(instanceName string, urls []string, models map[int64]modeler.Model) error {

	instance := &Instance{false, instanceName, urls, nil, map[string]*LogicalIndex{}}

	if err := instance.initialize(models); err != nil {
		zap.L().Error("instance.initialize()", zap.Error(err))
		return err
	}

	master.Instances[instanceName] = instance
	master.Initialized = true
	return nil
}
