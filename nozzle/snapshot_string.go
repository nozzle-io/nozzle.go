package nozzle

// snapshotString copies a bounded C char array without changing its byte encoding.
func snapshotString[T ~int8 | ~uint8](chars []T) string {
	end := len(chars)
	for i, ch := range chars {
		if ch == 0 {
			end = i
			break
		}
	}
	bytes := make([]byte, end)
	for i, ch := range chars[:end] {
		bytes[i] = byte(ch)
	}
	return string(bytes)
}
