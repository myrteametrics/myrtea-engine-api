package situation

import (
	"sync"

	"github.com/jmoiron/sqlx"
)

// Repository is a storage interface which can be implemented by multiple backend
// (in-memory map, sql database, in-memory cache, file system, ...)
// It allows standard CRUD operation on situations
type Repository interface {
	Get(id int64, groups []int64) (Situation, bool, error)
	GetByName(name string, groups []int64) (Situation, bool, error)
	Create(situation Situation) (int64, error)
	Update(id int64, situation Situation) error
	Delete(id int64) error
	GetAll(groups []int64) (map[int64]Situation, error)
	GetAllByRuleID(groups []int64, ruleID int64) (map[int64]Situation, error)
	IsInGroups(id int64, groups []int64) (bool, error)
	GetRules(id int64) ([]int64, error)
	SetRules(id int64, rules []int64) error
	AddRule(tx *sqlx.Tx, id int64, ruleID int64) error
	RemoveRule(tx *sqlx.Tx, id int64, ruleID int64) error
	GetSituationsByFactID(factID int64, ignoreIsObject bool) ([]Situation, error)
	GetFacts(id int64) ([]int64, error)
	CreateTemplateInstance(situationID int64, instance TemplateInstance) (int64, error)
	UpdateTemplateInstance(instanceID int64, instance TemplateInstance) error
	DeleteTemplateInstance(instanceID int64) error
	GetTemplateInstance(instanceID int64) (TemplateInstance, bool, error)
	GetAllTemplateInstances(situationID int64) (map[int64]TemplateInstance, error)
}

var (
	_globalRepositoryMu sync.RWMutex
	_globalRepository   Repository
)

// R is used to access the global repository singleton
func R() Repository {
	_globalRepositoryMu.RLock()
	defer _globalRepositoryMu.RUnlock()

	repository := _globalRepository
	return repository
}

// ReplaceGlobals affect a new repository to the global repository singleton
func ReplaceGlobals(repository Repository) func() {
	_globalRepositoryMu.Lock()
	defer _globalRepositoryMu.Unlock()

	prev := _globalRepository
	_globalRepository = repository
	return func() { ReplaceGlobals(prev) }
}
