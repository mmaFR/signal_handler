package signal_handler

import "errors"

var ErrAlreadyStarted = errors.New("handler already started")
var ErrNotRunning = errors.New("handler not running")
