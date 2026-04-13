package domain

import "errors"

// ErrNotFound is returned by any LocationService implementation when a CEP
// cannot be resolved to a city. Handlers check this to return HTTP 404.
var ErrNotFound = errors.New("not found")
