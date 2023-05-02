package scheduler

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

/*
   la fonction parseDuration prend une chaîne de durée sous la forme "1y 12mo 7d 3h 35m 3s"
   et la convertit en un objet time.Duration en extrayant et en traitant les différentes
   unités de temps (années, mois, jours, heures, minutes et secondes).
   Les années et les mois sont convertis en jours pour faciliter le calcul
   de la durée totale, qui est ensuite retournée par la fonction.
*/

// const (
//
//	daysInOneYear = 365.25
//	daysInOneMonth =  30.44
//
// )
func parseDuration(duration string) (time.Duration, error) {

	/*
	   analyser la chaîne de caractères duration et d'extraire les différentes unités de temps
	   en fonction de leur présence et de leur valeur dans la chaîne.
	*/
	re := regexp.MustCompile(`^(?P<years>\d+y)?\s*(?P<months>\d+mo)?\s*(?P<days>\d+d)?\s*(?P<hours>\d+h)?\s*(?P<minutes>\d+m)?\s*(?P<seconds>\d+s)?$`)
	match := re.FindStringSubmatch(duration)

	if match == nil {
		return 0, fmt.Errorf("Invalid duration format")
	}

	years := 0
	months := 0
	days := 0
	totalDuration := time.Duration(0)

	for i, name := range re.SubexpNames() {
		if i != 0 && match[i] != "" {
			value, _ := strconv.Atoi(strings.Trim(match[i], "ydhmsmo"))

			switch name {
			case "years":
				years = value
			case "months":
				months = value
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

	//  Add years and months using AddDate, then convert the result to days

	/*
	   Premier approche :
	    on depend la date actuelle pour calculer daysInPeriode et peux genereer des resultat incorrects
	*/

	now := time.Now()
	targetDate := now.AddDate(years, months, 0)
	daysInPeriod := int(targetDate.Sub(now).Hours() / 24)
	days += daysInPeriod

	/*
		    deuxiemem  approche :

		    on supposer qu'une année a une durée moyenne de 365,25 jours (en tenant compte
			des années bissextiles) et qu'un mois a une durée moyenne de 30,44 jours
			(en tenant compte des différents nombres de jours dans les mois).
	*/
	// daysInYears := float64(years) * daysInOneYear
	// daysInMonths := float64(months) * daysInOneMonth
	// daysInPeriod := int(daysInYears + daysInMonths)
	// days += daysInPeriod

	// Add days and other time units using Add
	totalDuration += time.Duration(days) * 24 * time.Hour
	return totalDuration, nil
}
