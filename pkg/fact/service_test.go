package fact

import (
	"testing"

	"github.com/myrteametrics/myrtea-sdk/v5/engine"
)

func TestCanUseElasticsearchCountAPI(t *testing.T) {
	tests := []struct {
		name string
		fact engine.Fact
		want bool
	}{
		{
			name: "basic count on model without dimensions",
			fact: engine.Fact{
				Model: "my_model",
				Intent: &engine.IntentFragment{
					Operator: engine.Count,
					Term:     "my_model",
				},
			},
			want: true,
		},
		{
			name: "count on specific field uses search API",
			fact: engine.Fact{
				Model: "my_model",
				Intent: &engine.IntentFragment{
					Operator: engine.Count,
					Term:     "group",
				},
			},
			want: false,
		},
		{
			name: "non count operator uses search API",
			fact: engine.Fact{
				Model: "my_model",
				Intent: &engine.IntentFragment{
					Operator: engine.Avg,
					Term:     "my_model",
				},
			},
			want: false,
		},
		{
			name: "count with dimensions uses search API",
			fact: engine.Fact{
				Model: "my_model",
				Intent: &engine.IntentFragment{
					Operator: engine.Count,
					Term:     "my_model",
				},
				Dimensions: []*engine.DimensionFragment{
					{Name: "dimension_1"},
				},
			},
			want: false,
		},
		{
			name: "nil intent uses search API",
			fact: engine.Fact{
				Model: "my_model",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canUseElasticsearchCountAPI(tt.fact); got != tt.want {
				t.Errorf("canUseElasticsearchCountAPI() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildWidgetDataFromCount(t *testing.T) {
	const expectedCount int64 = 42

	widgetData := buildWidgetDataFromCount(expectedCount)
	if widgetData == nil {
		t.Fatal("buildWidgetDataFromCount() returned nil")
	}

	if len(widgetData.Hits) != 0 {
		t.Fatalf("buildWidgetDataFromCount() hits length = %d, want 0", len(widgetData.Hits))
	}

	if widgetData.Aggregates == nil || widgetData.Aggregates.Aggs == nil {
		t.Fatal("buildWidgetDataFromCount() aggregates are nil")
	}

	docCount, ok := widgetData.Aggregates.Aggs["doc_count"]
	if !ok || docCount == nil {
		t.Fatal("buildWidgetDataFromCount() missing doc_count aggregation")
	}

	if docCount.Value != expectedCount {
		t.Fatalf("doc_count value = %v, want %d", docCount.Value, expectedCount)
	}
}
