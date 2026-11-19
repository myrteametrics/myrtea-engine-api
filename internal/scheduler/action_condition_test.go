package scheduler

import (
	"testing"

	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/history"
	"github.com/myrteametrics/myrtea-engine-api/v5/pkg/metadata"
	"github.com/myrteametrics/myrtea-sdk/v5/ruleeng"
)

func conditionedAction(slots ...ruleeng.ActionConditionSlot) ruleeng.Action {
	return ruleeng.DefaultAction{
		Name:                  "create-issue",
		EnableActionCondition: true,
		ActionCondition:       &ruleeng.ActionCondition{Slots: slots},
	}
}

// loadedWindow builds a window already filled, so no database is involved.
// rows[0] is T-1, rows[1] is T-2, ...
func loadedWindow(firedByRow ...[]string) *actionHistoryWindow {
	w := newActionHistoryWindow(1, 1, len(firedByRow))
	w.loaded = true
	for _, fired := range firedByRow {
		row := historyRowWith(fired...)
		w.rows = append(w.rows, row)
	}
	return w
}

func TestIsActionConditionVerified(t *testing.T) {
	current := map[string]ruleeng.Action{"A": ruleeng.DefaultAction{ID: "A", Name: "set"}}

	tests := []struct {
		name   string
		action ruleeng.Action
		window *actionHistoryWindow
		want   bool
	}{
		{
			name:   "condition disabled",
			action: ruleeng.DefaultAction{Name: "create-issue"},
			window: loadedWindow(),
			want:   true,
		},
		{
			name:   "no slot enabled",
			action: conditionedAction(ruleeng.ActionConditionSlot{Offset: 1, Enabled: false, ActionSetID: "A"}),
			window: loadedWindow(),
			want:   true,
		},
		{
			name:   "offset 0 fired",
			action: conditionedAction(ruleeng.ActionConditionSlot{Offset: 0, Enabled: true, ActionSetID: "A"}),
			window: loadedWindow(),
			want:   true,
		},
		{
			name:   "offset 0 not fired",
			action: conditionedAction(ruleeng.ActionConditionSlot{Offset: 0, Enabled: true, ActionSetID: "B"}),
			window: loadedWindow(),
			want:   false,
		},
		{
			name:   "offset 1 fired",
			action: conditionedAction(ruleeng.ActionConditionSlot{Offset: 1, Enabled: true, ActionSetID: "A"}),
			window: loadedWindow([]string{"A"}),
			want:   true,
		},
		{
			name:   "offset 3 fired exactly three ticks ago",
			action: conditionedAction(ruleeng.ActionConditionSlot{Offset: 3, Enabled: true, ActionSetID: "A"}),
			window: loadedWindow(nil, nil, []string{"A"}),
			want:   true,
		},
		{
			name:   "offset 3 not fired while it fired at offset 2",
			action: conditionedAction(ruleeng.ActionConditionSlot{Offset: 3, Enabled: true, ActionSetID: "A"}),
			window: loadedWindow(nil, []string{"A"}, nil),
			want:   false,
		},
		{
			name:   "offset beyond the available history",
			action: conditionedAction(ruleeng.ActionConditionSlot{Offset: 4, Enabled: true, ActionSetID: "A"}),
			window: loadedWindow([]string{"A"}, []string{"A"}),
			want:   false,
		},
		{
			name: "several offsets, all verified",
			action: conditionedAction(
				ruleeng.ActionConditionSlot{Offset: 0, Enabled: true, ActionSetID: "A"},
				ruleeng.ActionConditionSlot{Offset: 1, Enabled: true, ActionSetID: "A"},
				ruleeng.ActionConditionSlot{Offset: 2, Enabled: true, ActionSetID: "A"},
			),
			window: loadedWindow([]string{"A"}, []string{"A"}),
			want:   true,
		},
		{
			name: "several offsets, one missing",
			action: conditionedAction(
				ruleeng.ActionConditionSlot{Offset: 0, Enabled: true, ActionSetID: "A"},
				ruleeng.ActionConditionSlot{Offset: 1, Enabled: true, ActionSetID: "A"},
				ruleeng.ActionConditionSlot{Offset: 2, Enabled: true, ActionSetID: "A"},
			),
			window: loadedWindow([]string{"A"}, nil),
			want:   false,
		},
		{
			name: "a disabled slot is ignored",
			action: conditionedAction(
				ruleeng.ActionConditionSlot{Offset: 1, Enabled: true, ActionSetID: "A"},
				ruleeng.ActionConditionSlot{Offset: 5, Enabled: false, ActionSetID: "A"},
			),
			window: loadedWindow([]string{"A"}),
			want:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := isActionConditionVerified(tc.action, current, tc.window); got != tc.want {
				t.Errorf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

func TestMaxActionConditionDepth(t *testing.T) {
	agenda := []ruleeng.Action{
		ruleeng.DefaultAction{Name: "set"},
		conditionedAction(ruleeng.ActionConditionSlot{Offset: 2, Enabled: true, ActionSetID: "A"}),
		conditionedAction(
			ruleeng.ActionConditionSlot{Offset: 4, Enabled: true, ActionSetID: "A"},
			ruleeng.ActionConditionSlot{Offset: 9, Enabled: false, ActionSetID: "A"},
		),
	}
	if got := maxActionConditionDepth(agenda); got != 4 {
		t.Errorf("expected 4, got %d", got)
	}
	if got := maxActionConditionDepth(nil); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

func TestActionHistoryWindowRejectsOffsetZero(t *testing.T) {
	if loadedWindow([]string{"A"}).actionSetFiredAt(0, "A") {
		t.Errorf("offset 0 is the current tick and never read from the history")
	}
}

func historyRowWith(actionIDs ...string) history.HistorySituationsV4 {
	row := history.HistorySituationsV4{}
	for _, id := range actionIDs {
		row.Metadatas = append(row.Metadatas, metadata.MetaData{Key: "statut", Value: "critical", ActionID: id})
	}
	return row
}
