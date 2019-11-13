package util

import "strings"

// HTTPMethod returns the currect http method
func HTTPMethod(arg string) string {
	methods := []string{"GET", "HEAD", "POST", "PUT", "DELETE", "CONNECT", "OPTIONS", "TRACE"}
	arg = strings.ToUpper(arg)
	for _, m := range methods {
		if m == arg {
			return arg
		}
	}
	return "GET"
}
