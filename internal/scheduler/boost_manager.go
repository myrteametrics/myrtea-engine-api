package scheduler

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/internal/model"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/metadata"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// BoostAction represents a pending boost or revert action for a job
type BoostAction struct {
	JobID     string    `json:"jobId"`
	CreatedAt time.Time `json:"createdAt"`
	Read      bool      `json:"read"`
}

// BoostManager is an in-memory service that tracks which jobs need to be boosted or reverted
type BoostManager struct {
	mu         sync.RWMutex
	boostList  map[string]*BoostAction
	revertList map[string]*BoostAction
	ttl        time.Duration
	stopChan   chan struct{}
}

var (
	_globalBoostManagerMu sync.RWMutex
	_globalBoostManager   *BoostManager
)

// BM returns the global BoostManager singleton
func BM() *BoostManager {
	_globalBoostManagerMu.RLock()
	defer _globalBoostManagerMu.RUnlock()
	return _globalBoostManager
}

// ReplaceGlobalBoostManager sets the global BoostManager singleton
func ReplaceGlobalBoostManager(bm *BoostManager) func() {
	_globalBoostManagerMu.Lock()
	defer _globalBoostManagerMu.Unlock()
	prev := _globalBoostManager
	_globalBoostManager = bm
	return func() { ReplaceGlobalBoostManager(prev) }
}

// NewBoostManager creates a new BoostManager with the given TTL
func NewBoostManager() *BoostManager {
	ttl := viper.GetDuration("BOOST_LIFETIME")
	return &BoostManager{
		boostList:  make(map[string]*BoostAction),
		revertList: make(map[string]*BoostAction),
		ttl:        ttl,
		stopChan:   make(chan struct{}),
	}
}

// Start begins the background cleanup goroutine
func (bm *BoostManager) Start() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				bm.cleanup()
			case <-bm.stopChan:
				return
			}
		}
	}()
	zap.L().Info("BoostManager started")
}

// Stop stops the background cleanup goroutine
func (bm *BoostManager) Stop() {
	close(bm.stopChan)
	zap.L().Info("BoostManager stopped")
}

// Evaluate processes metadata and boost info to decide if a job should be boosted or reverted
func (bm *BoostManager) Evaluate(metadatas []metadata.MetaData, boostInfo model.BoostInfo) {
	var value string
	for _, md := range metadatas {
		v, ok := md.Value.(string)
		if !ok {
			continue
		}

		normalized := strings.ToLower(strings.TrimSpace(v))

		if normalized == model.Critical.String() || normalized == model.Ok.String() {
			value = v
			break
		}
	}

	if value == "" {
		return
	}

	switch value {
	case model.Critical.String():
		if boostInfo.Active {
			return
		}
		bm.addToBoostList(boostInfo.JobID)

	case model.Ok.String():
		if !boostInfo.Active {
			return
		}
		if boostInfo.Quota <= boostInfo.Used {
			return
		}
		bm.addToRevertList(boostInfo.JobID)
	}
}

// addToBoostList removes the job from both lists then adds it to the boost list
func (bm *BoostManager) addToBoostList(jobID string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	delete(bm.boostList, jobID)
	delete(bm.revertList, jobID)

	bm.boostList[jobID] = &BoostAction{
		JobID:     jobID,
		CreatedAt: time.Now(),
	}
	zap.L().Info("Job added to boost list", zap.String("jobId", jobID))
}

// addToRevertList removes the job from both lists then adds it to the revert list
func (bm *BoostManager) addToRevertList(jobID string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	delete(bm.boostList, jobID)
	delete(bm.revertList, jobID)

	bm.revertList[jobID] = &BoostAction{
		JobID:     jobID,
		CreatedAt: time.Now(),
	}
	zap.L().Info("Job added to revert list", zap.String("jobId", jobID))
}

// GetBoostList returns a copy of the current boost list
func (bm *BoostManager) GetBoostList() []BoostAction {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	result := make([]BoostAction, 0, len(bm.boostList))
	for _, action := range bm.boostList {
		if action.Read {
			continue
		}
		result = append(result, *action)
	}
	return result
}

// GetRevertList returns a copy of the current revert list
func (bm *BoostManager) GetRevertList() []BoostAction {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	result := make([]BoostAction, 0, len(bm.revertList))
	for _, action := range bm.revertList {
		if action.Read {
			continue
		}
		result = append(result, *action)
	}
	return result
}

// AcknowledgeBoost marks a job in the boost list as read
func (bm *BoostManager) AcknowledgeBoost(jobID string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	action, ok := bm.boostList[jobID]
	if !ok {
		return fmt.Errorf("job %s not found in boost list", jobID)
	}
	action.Read = true
	delete(bm.boostList, jobID)
	return nil
}

// AcknowledgeRevert marks a job in the revert list as read
func (bm *BoostManager) AcknowledgeRevert(jobID string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	action, ok := bm.revertList[jobID]
	if !ok {
		return fmt.Errorf("job %s not found in revert list", jobID)
	}
	action.Read = true
	delete(bm.boostList, jobID)
	return nil
}

// cleanup removes read items and expired items (TTL) from both lists
func (bm *BoostManager) cleanup() {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	now := time.Now()

	for jobID, action := range bm.boostList {
		if action.Read || now.Sub(action.CreatedAt) > bm.ttl {
			delete(bm.boostList, jobID)
			zap.L().Debug("Cleaned up boost action", zap.String("jobId", jobID), zap.Bool("read", action.Read))
		}
	}

	for jobID, action := range bm.revertList {
		if action.Read || now.Sub(action.CreatedAt) > bm.ttl {
			delete(bm.revertList, jobID)
			zap.L().Debug("Cleaned up revert action", zap.String("jobId", jobID), zap.Bool("read", action.Read))
		}
	}
}
