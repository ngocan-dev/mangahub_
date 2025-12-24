package progress

func icon(ok bool) string {
	if ok {
		return "✓"
	}
	return "✗"
}
