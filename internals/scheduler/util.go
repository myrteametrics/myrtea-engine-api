package scheduler

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
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
