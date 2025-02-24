package utility

func PtrOfString(val string) *string {
	oval := val
	return &oval
}

func PtrOfInt32(val int32) *int32 {
	oval := val
	return &oval
}

func PtrOfInt(val int) *int {
	oval := val
	return &oval
}

func PtrOfInt64(val int64) *int64 {
	oval := val
	return &oval
}

func PtrOfBool(val bool) *bool {
	oval := val
	return &oval
}
