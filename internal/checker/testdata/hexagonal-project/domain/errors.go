package domain

import "errors"

var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("conflict")

// GlobalCache is a mutable global — should trigger global_mutable_state.
var GlobalCache = map[string]interface{}{}
