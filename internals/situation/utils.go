package situation

func getTranslateValue(translateOpt ...bool) bool {
	if len(translateOpt) > 0 {
		return translateOpt[0]
	}
	return true
}
