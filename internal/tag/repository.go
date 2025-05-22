package tag

import "sync"

// Repository is a storage interface which can be implemented by multiple backend
// (in-memory map, sql database, in-memory cache, file system, ...)
// It allows standard CRUD operation on tags
type Repository interface {
	Create(tag Tag) (int64, error)
	Get(id int64) (Tag, bool, error)
	Update(tag Tag) error
	Delete(id int64) error
	GetAll() ([]Tag, error)

	CreateLinkWithSituation(tagID int64, situationID int64) error
	DeleteLinkWithSituation(tagID int64, situationID int64) error
	GetTagsBySituationId(situationId int64) ([]Tag, error)

	CreateLinkWithTemplateInstance(tagID int64, templateInstanceID int64) error
	DeleteLinkWithTemplateInstance(tagID int64, templateInstanceID int64) error
	GetTagsByTemplateInstanceId(templateInstanceId int64) ([]Tag, error)

	GetSituationsTags() (map[int64][]Tag, error)
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
