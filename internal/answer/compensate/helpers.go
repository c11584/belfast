package compensate

func boolToUint32(value bool) uint32 {
	if value {
		return 1
	}
	return 0
}
