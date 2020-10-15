package goafweb

import "errors"

var ErrNotFound = errors.New("Database Error: Resource not found.")
var ErrPWInvalid = errors.New("Authentication error: password invalid.")
var ErrAuth = errors.New("Authentication error: username/password invalid")
