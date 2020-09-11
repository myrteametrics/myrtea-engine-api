package groups

// This is the universal token of groups, giving access as if it was all the groups
const (
	// AllGroups SQL statement for the issues table
	AllGroups int64 = 0
)

//GetTokenAllGroups gets the array with the universal token group
func GetTokenAllGroups() []int64 {
	return []int64{AllGroups}
}

//DeleteTokenAllGroups delete the universal token from groups
func DeleteTokenAllGroups(groups []int64) []int64 {
	newGroups := make([]int64, 0)
	for _, group := range groups {
		if group != AllGroups {
			newGroups = append(newGroups, group)
		}
	}
	return newGroups
}
