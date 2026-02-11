package scheduler

import (
	"context"
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
	JobID     string    `json:"jobId"`     // Deterministic job ID
	CreatedAt time.Time `json:"createdAt"` // When the action was created
	ExpiresAt time.Time `json:"expiresAt"` // When the action will expire (CreatedAt + TTL)
}

// BoostManager is an in-memory service that tracks which jobs need to be boosted or reverted
type BoostManager struct {
	mu         sync.RWMutex
	boostList  map[string]*BoostAction // Jobs that should be boosted
	revertList map[string]*BoostAction // Jobs that should revert to normal
	ttl        time.Duration           // Time-to-live for actions
	ctx        context.Context         // Application context
	cancel     context.CancelFunc      // Cancel function for cleanup
	wg         sync.WaitGroup          // Wait group for graceful shutdown
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

	if _globalBoostManager != nil {
		zap.L().Info("Stopping previous BoostManager...")
		_globalBoostManager.Stop()
	}

	prev := _globalBoostManager
	_globalBoostManager = bm

	bm.Start()

	return func() { ReplaceGlobalBoostManager(prev) }
}

// NewBoostManager creates a new BoostManager with the given TTL
func NewBoostManager() *BoostManager {
	ttl := viper.GetDuration("BOOST_LIFETIME")
	if ttl == 0 {
		ttl = 5 * time.Minute
		zap.L().Warn("BOOST_LIFETIME not configured, using default", zap.Duration("ttl", ttl))
	}

	boostCtx, cancel := context.WithCancel(context.Background())

	bm := &BoostManager{
		boostList:  make(map[string]*BoostAction),
		revertList: make(map[string]*BoostAction),
		ttl:        ttl,
		ctx:        boostCtx,
		cancel:     cancel,
	}

	return bm
}

// Start begins the background cleanup goroutine
func (bm *BoostManager) Start() {
	bm.wg.Add(1)
	go bm.cleanupRoutine()
	zap.L().Info("BoostManager started",
		zap.Duration("ttl", bm.ttl),
		zap.String("cleanupInterval", "1m"))
}

// cleanupRoutine runs the periodic cleanup in a goroutine
func (bm *BoostManager) cleanupRoutine() {
	defer bm.wg.Done()

	cleanupInterval := 1 * time.Minute
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()

	zap.L().Info("Cleanup routine started", zap.Duration("interval", cleanupInterval))

	for {
		select {
		case <-ticker.C:
			bm.cleanup()
		case <-bm.ctx.Done():
			zap.L().Info("BoostManager cleanup routine stopping due to context cancellation")
			bm.cleanup()
			return
		}
	}
}

// Stop stops the background cleanup goroutine and waits for it to finish
func (bm *BoostManager) Stop() {
	zap.L().Info("Stopping BoostManager...")

	// Cancel the context to signal goroutine to stop
	bm.cancel()

	// Wait for cleanup routine to finish
	bm.wg.Wait()

	zap.L().Info("BoostManager stopped gracefully")
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

	now := time.Now()
	bm.boostList[jobID] = &BoostAction{
		JobID:     jobID,
		CreatedAt: now,
		ExpiresAt: now.Add(bm.ttl),
	}

	zap.L().Info("Job added to boost list",
		zap.String("jobId", jobID),
		zap.Time("expiresAt", bm.boostList[jobID].ExpiresAt))
}

// addToRevertList removes the job from both lists then adds it to the revert list
func (bm *BoostManager) addToRevertList(jobID string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	delete(bm.boostList, jobID)
	delete(bm.revertList, jobID)

	now := time.Now()
	bm.revertList[jobID] = &BoostAction{
		JobID:     jobID,
		CreatedAt: now,
		ExpiresAt: now.Add(bm.ttl),
	}

	zap.L().Info("Job added to revert list",
		zap.String("jobId", jobID),
		zap.Time("expiresAt", bm.revertList[jobID].ExpiresAt))
}

// GetBoostList returns a copy of the current boost list (only non-read items)
func (bm *BoostManager) GetBoostList() []BoostAction {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	result := make([]BoostAction, 0, len(bm.boostList))
	for _, action := range bm.boostList {
		result = append(result, *action)
	}
	return result
}

// GetRevertList returns a copy of the current revert list (only non-read items)
func (bm *BoostManager) GetRevertList() []BoostAction {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	result := make([]BoostAction, 0, len(bm.revertList))
	for _, action := range bm.revertList {
		result = append(result, *action)
	}
	return result
}

// AcknowledgeBoost marks a job in the boost list as read and removes it
func (bm *BoostManager) AcknowledgeBoost(jobID string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	_, ok := bm.boostList[jobID]
	if !ok {
		return fmt.Errorf("job %s not found in boost list", jobID)
	}

	delete(bm.boostList, jobID)

	zap.L().Debug("Boost action acknowledged", zap.String("jobId", jobID))
	return nil
}

// AcknowledgeRevert marks a job in the revert list as read and removes it
func (bm *BoostManager) AcknowledgeRevert(jobID string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	_, ok := bm.revertList[jobID]
	if !ok {
		return fmt.Errorf("job %s not found in revert list", jobID)
	}

	delete(bm.revertList, jobID)

	zap.L().Debug("Revert action acknowledged", zap.String("jobId", jobID))
	return nil
}

// cleanup removes expired items (based on individual ExpiresAt) from both lists
func (bm *BoostManager) cleanup() {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	now := time.Now()
	boostCleaned := 0
	revertCleaned := 0

	for jobID, action := range bm.boostList {
		if now.After(action.ExpiresAt) {
			delete(bm.boostList, jobID)
			boostCleaned++

			age := now.Sub(action.CreatedAt)
			overtime := now.Sub(action.ExpiresAt)

			zap.L().Debug("Cleaned up expired boost action",
				zap.String("jobId", jobID),
				zap.Duration("age", age),
				zap.Duration("overtime", overtime),
				zap.Time("createdAt", action.CreatedAt),
				zap.Time("expiresAt", action.ExpiresAt))
		}
	}

	for jobID, action := range bm.revertList {
		if now.After(action.ExpiresAt) {
			delete(bm.revertList, jobID)
			revertCleaned++

			age := now.Sub(action.CreatedAt)
			overtime := now.Sub(action.ExpiresAt)

			zap.L().Debug("Cleaned up expired revert action",
				zap.String("jobId", jobID),
				zap.Duration("age", age),
				zap.Duration("overtime", overtime),
				zap.Time("createdAt", action.CreatedAt),
				zap.Time("expiresAt", action.ExpiresAt))
		}
	}

	if boostCleaned > 0 || revertCleaned > 0 {
		zap.L().Info("Cleanup completed",
			zap.Int("boostCleaned", boostCleaned),
			zap.Int("revertCleaned", revertCleaned),
			zap.Int("boostRemaining", len(bm.boostList)),
			zap.Int("revertRemaining", len(bm.revertList)))
	}
}

// IsExpired checks if an action has expired
func (action *BoostAction) IsExpired() bool {
	return time.Now().After(action.ExpiresAt)
}

// TimeUntilExpiration returns the duration until this action expires
func (action *BoostAction) TimeUntilExpiration() time.Duration {
	remaining := action.ExpiresAt.Sub(time.Now())
	if remaining < 0 {
		return 0
	}
	return remaining
}
