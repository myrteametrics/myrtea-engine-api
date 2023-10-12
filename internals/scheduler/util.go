package scheduler

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	KeyMetadataDependsOn     = "depends_on_metadata"
	ValueMetadataDependsOn   = "depends_on_metadata_value"
	ActionSetValue           = "set"
	ActionPendingValue       = "pending"
	IDSituationDependsOn   = "id_situation_depends_on"
	IDInstanceDependsOn    = "id_instance_depends_on"
)




func parseDuration(duration string) (time.Duration, error) {

	re := regexp.MustCompile(`^(?P<days>\d+d)?\s*(?P<hours>\d+h)?\s*(?P<minutes>\d+m)?\s*(?P<seconds>\d+s)?$`)
	match := re.FindStringSubmatch(duration)

	if match == nil {
		return 0, fmt.Errorf("Invalid duration format")
	}

	days := 0
	totalDuration := time.Duration(0)

	for i, name := range re.SubexpNames() {
		if i != 0 && match[i] != "" {
			value, _ := strconv.Atoi(strings.Trim(match[i], "ydhmsmo"))

			switch name {
			case "days":
				days += value
			case "hours":
				totalDuration += time.Duration(value) * time.Hour
			case "minutes":
				totalDuration += time.Duration(value) * time.Minute
			case "seconds":
				totalDuration += time.Duration(value) * time.Second
			}
		}
	}

	totalDuration += time.Duration(days) * 24 * time.Hour
	return totalDuration, nil
}

func generateKeyAndValues(situation map[string]string) (string, int, int, error) {
    strIdSituationDependsOn, ok1 := situation[IDSituationDependsOn]
    strIdInstanceDependsOn, ok2 := situation[IDInstanceDependsOn]

    if !ok1 || !ok2 {
        return "", 0, 0, fmt.Errorf("couldn't retrieve dependency parameters: missing %s or %s", IDSituationDependsOn, IDInstanceDependsOn)
    }

    idSituationDependsOn, err1 := strconv.Atoi(strIdSituationDependsOn)
    idInstanceDependsOn, err2 := strconv.Atoi(strIdInstanceDependsOn)

    if err1 != nil || err2 != nil {
        return "", 0, 0, fmt.Errorf("error converting ids to int: %v %v", err1, err2)
    }

    key := fmt.Sprintf("%v-%v", idSituationDependsOn, idInstanceDependsOn)
    return key, idSituationDependsOn, idInstanceDependsOn, nil
}

func logDataRetrieval(
	success bool,
	parentSituationID, parentSituationInstanceID int,
	childSituationID, childSituationInstanceID int64,
	err error,
	retrievalTimestamp string,
) {
	if success {
		zap.L().Info(
			"Successfully retrieved the latest history for Parent from database and it's critical",
			zap.Int("ParentSituationID", parentSituationID),
			zap.Int("ParentSituationInstanceID", parentSituationInstanceID),
			zap.String("Timestamp of last evaluation of parent", retrievalTimestamp),
			zap.Int64("ChildSituationID", childSituationID),
			zap.Int64("ChildSituationInstanceID", childSituationInstanceID),
		)		
	} else {
		zap.L().Error(
			"Failed to retrieve the latest history for Parent from database ",
			zap.Int("ParentSituationID", parentSituationID),
			zap.Int("ParentSituationInstanceID", parentSituationInstanceID),
			zap.Int64("ChildSituationID", childSituationID),
			zap.Int64("ChildSituationInstanceID", childSituationInstanceID),
			zap.Error(err),
		)
	}
}

