package resources

// stringValue safely dereferences a string pointer
func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ptrInt32Value safely dereferences an int32 pointer
func ptrInt32Value(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

// ptrInt64Value safely dereferences an int64 pointer
func ptrInt64Value(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}
