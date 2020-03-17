package util

import "strings"

// DefaultHTTPMethod for not specified methods
const DefaultHTTPMethod = "GET"

// HTTPMethod returns the currect http method
// Default method: GET
func HTTPMethod(arg string) string {
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE", "CONNECT", "OPTIONS", "TRACE"}

	arg = strings.ToUpper(arg)

	for _, m := range methods {
		if m == arg {
			return arg
		}
	}

	// return default method
	return DefaultHTTPMethod
}
