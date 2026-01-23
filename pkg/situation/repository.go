package situation

import (
	"sync"

	"github.com/jmoiron/sqlx"
)

// Repository is a storage interface which can be implemented by multiple backend
// (in-memory map, sql database, in-memory cache, file system, ...)
// It allows standard CRUD operation on situations
type Repository interface {
	Get(id int64, parseGlobalVariables ...bool) (Situation, bool, error)
	GetByName(name string, parseGlobalVariables ...bool) (Situation, bool, error)
	Create(situation Situation) (int64, error)
	Update(id int64, situation Situation) error
	Delete(id int64) error
	GetAll(parseGlobalVariables ...bool) (map[int64]Situation, error)
	GetAllByIDs(ids []int64, parseGlobalVariables ...bool) (map[int64]Situation, error)
	GetAllByRuleID(ruleID int64, parseGlobalVariables ...bool) (map[int64]Situation, error)

	GetRules(id int64) ([]int64, error)
	SetRules(id int64, rules []int64) error
	AddRule(tx *sqlx.Tx, id int64, ruleID int64) error
	RemoveRule(tx *sqlx.Tx, id int64, ruleID int64) error
	GetSituationsByFactID(factID int64, ignoreIsObject bool, parseGlobalVariables ...bool) ([]Situation, error)
	GetFacts(id int64) ([]int64, error)

	CreateTemplateInstance(situationID int64, instance TemplateInstance) (int64, error)
	UpdateTemplateInstance(instanceID int64, instance TemplateInstance) error
	DeleteTemplateInstance(instanceID int64) error
	GetTemplateInstance(instanceID int64, parseParameters ...bool) (TemplateInstance, bool, error)
	GetAllTemplateInstances(situationID int64, parseParameters ...bool) (map[int64]TemplateInstance, error)
	GetAllTemplateInstancesByIDs(ids []int64, parseParameters ...bool) (map[int64]TemplateInstance, error)

	GetAllTemplateInstancesByRuleID(ruleID int64, parseParameters ...bool) (map[int64]TemplateInstance, error)

	GetSituationOverview() ([]SituationOverview, error)
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
