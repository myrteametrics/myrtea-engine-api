package models

type keyContext int

const (
	//ContextKeyUser is used as key to add the user data in the request context
	ContextKeyUser keyContext = iota
	//ContextKeyLoggerR is used as key to add the value of the http.Request at the CustomLogger middlewere execution
	ContextKeyLoggerR
	//UserLogin is used as key to add the userloging into the request context
	UserLogin
)
