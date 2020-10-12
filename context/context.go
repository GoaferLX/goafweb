/*
Package context acts as a wrapper around stdlib.Context to provide utility
for storing/retreiving values from Context.
*/
package context

import (
	"context"
	"goafweb"
)

type ctxKey string

// Define userKey as constant so "user" can't be overwritten by anything malicious.
const (
	userKey ctxKey = "user"
)

// WithUser adds a User into Context.
func WithUser(ctx context.Context, user *goafweb.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// GetUser checks the Context to see if the userKey exists.
// Returns goafweb.User or nil.
func GetUser(ctx context.Context) *goafweb.User {
	if userctx := ctx.Value(userKey); userctx != nil {
		if user, ok := userctx.(*goafweb.User); ok {
			return user
		}
	}
	return nil
}
