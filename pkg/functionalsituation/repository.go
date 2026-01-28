package functionalsituation

import "sync"

// Repository is a storage interface which can be implemented by multiple backend
// (in-memory map, sql database, in-memory cache, file system, ...)
// It allows standard CRUD operation on functional situations
type Repository interface {
	// Basic CRUD operations
	Create(fs FunctionalSituation, createdBy string) (int64, error)
	Get(id int64) (FunctionalSituation, bool, error)
	GetByName(name string, parentID *int64) (FunctionalSituation, bool, error)
	Update(id int64, fs FunctionalSituationUpdate, updatedBy string) error
	Delete(id int64) error
	GetAll() ([]FunctionalSituation, error)

	// Hierarchy operations
	GetChildren(parentID int64) ([]FunctionalSituation, error)
	GetRoots() ([]FunctionalSituation, error)
	GetTree() ([]FunctionalSituation, error)
	GetAncestors(id int64) ([]FunctionalSituation, error)
	MoveToParent(id int64, newParentID *int64) error

	// Template Instance associations
	AddTemplateInstance(fsID int64, instanceID int64, parameters map[string]interface{}, addedBy string) error
	RemoveTemplateInstance(fsID int64, instanceID int64) error
	GetTemplateInstances(fsID int64) ([]int64, error)
	GetTemplateInstancesWithParameters(fsID int64) (map[int64]map[string]interface{}, error)
	GetFunctionalSituationsByInstance(instanceID int64) ([]FunctionalSituation, error)
	GetInstanceReference(instanceID int64) (InstanceReference, bool, error)
	UpdateInstanceReferenceParameters(instanceID int64, parameters map[string]interface{}) error
	GetAllInstanceReferences() ([]InstanceReference, error)

	// Situation associations
	AddSituation(fsID int64, situationID int64, parameters map[string]interface{}, addedBy string) error
	RemoveSituation(fsID int64, situationID int64) error
	GetSituations(fsID int64) ([]int64, error)
	GetSituationsWithParameters(fsID int64) (map[int64]map[string]interface{}, error)
	GetFunctionalSituationsBySituation(situationID int64) ([]FunctionalSituation, error)
	GetSituationReference(situationID int64) (SituationReference, bool, error)
	UpdateSituationReferenceParameters(situationID int64, parameters map[string]interface{}) error
	GetAllSituationReferences() ([]SituationReference, error)

	// Overview operations
	GetOverview() ([]FunctionalSituationOverview, error)
	GetOverviewByID(id int64) (FunctionalSituationOverview, bool, error)

	// Enriched tree with instances and situations
	GetEnrichedTree() ([]FunctionalSituationTreeNode, error)
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
