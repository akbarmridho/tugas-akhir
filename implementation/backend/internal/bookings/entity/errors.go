package entity

import "errors"

var InternalTicketLockError = errors.New("internal error during lock acquiring for seats")
