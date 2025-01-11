package main

type contextKey string

// to avoid conflicting names that third party packages might be using with r.Context()
const isAuthenticatedKey = contextKey("isAuthenticated")
