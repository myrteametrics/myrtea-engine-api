package fact

import (
	"errors"
	"strconv"
	"sync"

	"github.com/myrteametrics/myrtea-sdk/v4/engine"
)

// NativeMapRepository is an implementation of entity.Repository, supported by standard native in-memory map
type NativeMapRepository struct {
	mutex       sync.RWMutex
	factsByID   map[int64]engine.Fact
	factsByName map[string]engine.Fact
	nextInt     func() int64
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
		factsByID:   make(map[int64]engine.Fact, 0),
		factsByName: make(map[string]engine.Fact, 0),
		nextInt:     intSeq(),
	}

	var isr Repository = &r
	return isr
}

// Get search and returns an entity from the repository by its id
func (r *NativeMapRepository) Get(id int64) (engine.Fact, bool, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	fact, found := r.factsByID[id]
	if !found {
		return engine.Fact{}, false, nil
	}
	return fact, true, nil
}

// GetByName search and returns an entity from the repository by its name
func (r *NativeMapRepository) GetByName(name string) (engine.Fact, bool, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	fact, found := r.factsByName[name]
	if !found {
		return engine.Fact{}, false, nil
	}
	return fact, true, nil
}

// Create stores the given new fact in the database as a json.
func (r *NativeMapRepository) Create(fact engine.Fact) (int64, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	if _, ok := r.factsByName[fact.Name]; ok {
		return -1, errors.New("fact already exists for the Name: " + fact.Name)
	}
	factID := r.nextInt()
	if _, ok := r.factsByID[factID]; ok {
		return -1, errors.New("fact already exists for the ID:" + strconv.FormatInt(factID, 10))
	}

	//This is necessary because within the definition we don't have the id
	fact.ID = factID

	r.factsByID[factID] = fact
	r.factsByName[fact.Name] = fact
	return factID, nil
}

// Update updates an entity in the repository by its name
func (r *NativeMapRepository) Update(id int64, fact engine.Fact) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, ok := r.factsByID[id]; !ok {
		return errors.New("fact does not exists for the ID:" + strconv.FormatInt(id, 10))
	}

	//This is necessary because within the definition we don't have the id
	fact.ID = id

	r.factsByID[id] = fact
	r.factsByName[fact.Name] = fact
	return nil
}

// Delete deletes an entity from the repository by its name
func (r *NativeMapRepository) Delete(id int64) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, ok := r.factsByID[id]; !ok {
		return errors.New("fact does not exists for the ID:" + strconv.FormatInt(id, 10))
	}
	f := r.factsByID[id]
	delete(r.factsByName, f.Name)
	delete(r.factsByID, id)
	return nil
}

// GetAll returns all entities in the repository
func (r *NativeMapRepository) GetAll() (map[int64]engine.Fact, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	factsByID := r.factsByID
	return factsByID, nil
}

// GetAllByIDs returns all entities in the repository
func (r *NativeMapRepository) GetAllByIDs(ids []int64) (map[int64]engine.Fact, error) {
	r.mutex.RLock()
	factsByID := r.factsByID
	r.mutex.RUnlock()
	return factsByID, nil
}
