package functionalsituation

import (
	"testing"
	"time"
)

func TestFunctionalSituation_IsValid(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		fs      FunctionalSituation
		wantErr bool
	}{
		{"valid minimal", FunctionalSituation{Name: "Accounting", Color: "#112233", Icon: "folder", CreatedAt: now, UpdatedAt: now, CreatedBy: "test"}, false},
		{"missing name", FunctionalSituation{Color: "#112233", Icon: "folder"}, true},
		{"name too long", FunctionalSituation{Name: string(make([]byte, 101)), Color: "#112233"}, true},
		{"invalid color", FunctionalSituation{Name: "A", Color: "112233"}, true},
		{"invalid color length", FunctionalSituation{Name: "A", Color: "#123"}, true},
		{"icon too long", FunctionalSituation{Name: "A", Icon: string(make([]byte, 51))}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := tt.fs.IsValid()
			if (err != nil) != tt.wantErr {
				t.Fatalf("IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
			if ok && tt.wantErr {
				t.Fatalf("expected invalid but got valid")
			}
		})
	}
}

func TestFunctionalSituationCreate_IsValid(t *testing.T) {
	tests := []struct {
		name    string
		fsc     FunctionalSituationCreate
		wantErr bool
	}{
		{"valid create", FunctionalSituationCreate{Name: "Ops", Color: "#FFFFFF", Icon: "gear"}, false},
		{"missing name", FunctionalSituationCreate{Color: "#FFFFFF"}, true},
		{"invalid color", FunctionalSituationCreate{Name: "X", Color: "FFFFFF"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := tt.fsc.IsValid()
			if (err != nil) != tt.wantErr {
				t.Fatalf("IsValid() error = %v, wantErr %v", err, tt.wantErr)
			}
			if ok && tt.wantErr {
				t.Fatalf("expected invalid but got valid")
			}
		})
	}
}
