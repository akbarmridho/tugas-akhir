package entity

import "errors"

var InternalTicketLockError = errors.New("internal error during lock acquiring for seats")

var IssueTicketError = errors.New("internal error during issuing tickets")

var IssuedTicketFetchError = errors.New("internal error during fetch issued ticket")

var LockNotAcquiredError = errors.New("cannot acquire lock for the given seats")
