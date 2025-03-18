package entity

import "errors"

var InternalTicketLockError = errors.New("internal error during lock acquiring for seats")

var IssueTicketError = errors.New("internal error during issuing tickets")
