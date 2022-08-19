package situation

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/myrteametrics/myrtea-engine-api/v5/internals/fact"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// BuildSituationsFromFile reads an ensemble of situations from the provided file and returns them
// Deprecated: Situation loading from YAML file has been replaced by standard REST API
func BuildSituationsFromFile(path string, file string) (map[int64]*Situation, []error) {

	if fact.R() == nil {
		return nil, []error{errors.New("fact repository not initialized")}
	}

	conf := viper.New()
	conf.SetConfigType("yaml")
	conf.AddConfigPath(path)
	conf.SetConfigName(file)

	if err := conf.ReadInConfig(); err != nil {
		zap.L().Error(fmt.Sprintf("initializeConfig.ReadInConfig: %s", err))
		return nil, []error{err}
	}

	situations := make(map[int64]*Situation, 0)
	var errs []error

	situationsRaw := conf.GetStringMapStringSlice("situations")
	for key, situationFacts := range situationsRaw {
		facts := make([]int64, 0)

		for _, v := range situationFacts {
			f, found, err := fact.R().GetByName(v)
			if err != nil {
				errs = append(errs, err)
				continue
			}
			if !found {
				errs = append(errs, errors.New("fact "+v+" not found in fact repository"))
				continue
			}
			facts = append(facts, f.ID)
		}

		idSituation, err := strconv.ParseInt(key, 10, 64)
		if err == nil {
			zap.L().Error("Error on parsing situation id", zap.String("situationID", key))
		}

		situations[idSituation] = &Situation{Name: key, Facts: facts}
	}
	return situations, errs
}
