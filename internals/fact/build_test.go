package fact

import (
	"testing"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/fact/lexer"
	"github.com/spf13/viper"
)

// BuildFacts : This functions creates the facts
// from the file provided

func TestBuildFactsFromFileNoFile(t *testing.T) {
	_, errs := BuildFactsFromFile("testdata", "not_a_file")
	if len(errs) == 0 {
		t.Error("file doesn't exists but no error returned")
	}
}
func TestBuildFactsFromFileValid(t *testing.T) {
	viper.AddConfigPath("testdata")
	viper.SetConfigName("model_entities")
	viper.ReadInConfig()
	entities := viper.GetStringSlice("entities")
	l, err := lexer.New(entities)
	if err != nil {
		t.Error(err)
	}
	lexer.ReplaceGlobals(l)

	_, errs := BuildFactsFromFile("testdata", "facts")
	if errs != nil {
		t.Error(errs)
	}
}

func TestBuildFactsFromFileInvalid(t *testing.T) {
	viper.AddConfigPath("testdata")
	viper.SetConfigName("model_entities")
	viper.ReadInConfig()
	entities := viper.GetStringSlice("entities")
	l, err := lexer.New(entities)
	if err != nil {
		t.Error(err)
	}
	lexer.ReplaceGlobals(l)

	facts, _ := BuildFactsFromFile("testdata", "facts_invalid")
	if len(facts) > 0 {
		for _, fact := range facts {
			t.Log(fact)
		}
		t.Error("All facts should be invalid")
	}
}
