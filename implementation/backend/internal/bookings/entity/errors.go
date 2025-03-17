package entity

import "errors"

var ResultDataLengthNotMatch = errors.New("the result data length does not match with the param length")

var UpdatedDataLengthNotMatch = errors.New("the updated data length does not match with the param length")
