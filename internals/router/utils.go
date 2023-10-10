package router

import "go.uber.org/zap"

func sliceDeduplicate(stringSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range stringSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func errorInfo(message string, err error) {
	if err != nil {
		zap.L().Info(message, zap.Error(err))
	}
}
