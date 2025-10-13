package evaluator

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/myrteametrics/myrtea-engine-api/v5/internal/rule"
	"github.com/myrteametrics/myrtea-sdk/v5/ruleeng"
)

var (
	_globalMu      sync.RWMutex
	_globalREngine = make(map[string]*ruleeng.RuleEngine, 0)
	_lastUpdate    = make(map[string]time.Time, 0)
)

// GetEngine return a specific engine
func GetEngine(engineID string) (*ruleeng.RuleEngine, bool) {
	_globalMu.RLock()
	defer _globalMu.RUnlock()
	re, ok := _globalREngine[engineID]
	if ok {
		return re, true
	}
	return nil, false
}

// CloneEngine inits a new engine based on an existing one
func CloneEngine(engineID string, cloneRuleBase bool, cloneKnowledgeBase bool) (*ruleeng.RuleEngine, error) {
	if re, ok := GetEngine(engineID); !ok {
		return nil, fmt.Errorf("Engine with ID %s does not exist", engineID)
	} else {
		clone := ruleeng.NewRuleEngine()
		if cloneRuleBase {
			clone.SetRules(re.GetRulesBase())
		}
		if cloneKnowledgeBase {
			clone.SetKnowledge(re.GetKnowledgeBase())
		}
		return clone, nil
	}
}

// InitEngine inits an engine if it does not exist
func InitEngine(engineID string) error {
	if _, ok := GetEngine(engineID); ok {
		return fmt.Errorf("Engine with ID %s already exists", engineID)
	}

	_globalMu.Lock()
	defer _globalMu.Unlock()

	_globalREngine[engineID] = ruleeng.NewRuleEngine()
	_lastUpdate[engineID] = time.Now()

	rules, err := rule.R().GetAll()
	if err != nil {
		return errors.New("couldn't read rules " + err.Error())
	}

	for _, rule := range rules {
		if rule.Enabled {
			_globalREngine[engineID].InsertRule(&rule)
		}
	}
	return nil
}

// UpdateEngine updates an engine if it exists
func UpdateEngine(engineID string) error {
	if _, ok := GetEngine(engineID); !ok {
		return fmt.Errorf("Engine with ID %s does not exist", engineID)
	}

	_globalMu.Lock()
	defer _globalMu.Unlock()

	now := time.Now()
	rules, err := rule.R().GetAllModifiedFrom(_lastUpdate[engineID])
	if err != nil {
		return errors.New("couldn't read modified rules " + err.Error())
	}
	_lastUpdate[engineID] = now
	for _, rule := range rules {
		if rule.Enabled {
			_globalREngine[engineID].InsertRule(&rule)
		} else {
			_globalREngine[engineID].RemoveRule(rule.ID)
		}
	}

	return nil
}
