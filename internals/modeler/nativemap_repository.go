package modeler

import (
	"errors"
	"strconv"
	"sync"

	"github.com/myrteametrics/myrtea-sdk/v4/modeler"
)

// NativeMapRepository is an implementation of entity.Repository, supported by standard native in-memory map
type NativeMapRepository struct {
	mutex        sync.RWMutex
	modelsByID   map[int64]modeler.Model
	modelsByName map[string]modeler.Model
	nextInt      func() int64
}

func intSeq() func() int64 {
	i := int64(0)
	return func() int64 {
		i++
		return i
	}
}

// NewNativeMapRepository returns a new instance of NativeMapRepository
func NewNativeMapRepository() Repository {
	r := NativeMapRepository{
		modelsByID:   make(map[int64]modeler.Model, 0),
		modelsByName: make(map[string]modeler.Model, 0),
		nextInt:      intSeq(),
	}

	var isr Repository = &r
	return isr
}

// Get search and returns an entity from the repository by its id
func (r *NativeMapRepository) Get(id int64) (modeler.Model, bool, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	model, found := r.modelsByID[id]
	if !found {
		return modeler.Model{}, false, nil
	}
	return model, true, nil
}

// GetByName search and returns an entity from the repository by its name
func (r *NativeMapRepository) GetByName(name string) (modeler.Model, bool, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	model, found := r.modelsByName[name]
	if !found {
		return modeler.Model{}, false, nil
	}
	return model, true, nil
}

// Create stores the given new model in the database as a json.
func (r *NativeMapRepository) Create(model modeler.Model) (int64, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if _, ok := r.modelsByName[model.Name]; ok {
		return -1, errors.New("Model already exists for the Name: " + model.Name)
	}
	modelID := r.nextInt()
	if _, ok := r.modelsByID[modelID]; ok {
		return -1, errors.New("Model already exists for the ID:" + strconv.FormatInt(modelID, 10))
	}

	//This is necessary because within the definition we don't have the id
	model.ID = modelID

	r.modelsByID[modelID] = model
	r.modelsByName[model.Name] = model
	return modelID, nil
}

// Update updates an entity in the repository by its name
func (r *NativeMapRepository) Update(id int64, model modeler.Model) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.modelsByID[id]; !ok {
		return errors.New("Model does not exists for the ID:" + strconv.FormatInt(id, 10))
	}

	//This is necessary because within the definition we don't have the id
	model.ID = id

	r.modelsByID[id] = model
	r.modelsByName[model.Name] = model
	return nil
}

// Delete deletes an entity from the repository by its name
func (r *NativeMapRepository) Delete(id int64) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.modelsByID[id]; !ok {
		return errors.New("Model does not exists for the ID:" + strconv.FormatInt(id, 10))
	}
	f := r.modelsByID[id]
	delete(r.modelsByName, f.Name)
	delete(r.modelsByID, id)
	return nil
}

// GetAll returns all entities in the repository
func (r *NativeMapRepository) GetAll() (map[int64]modeler.Model, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	modelsByID := r.modelsByID
	return modelsByID, nil
}
