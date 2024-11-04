package domain

//go:generate go run github.com/abice/go-enum

// HTTPMethod ...
// ENUM(unspecified, get, put, post, delete, patch)
type HTTPMethod uint8
