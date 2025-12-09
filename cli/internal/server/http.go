package server

// HTTPComponent returns metadata for the HTTP API server.
func HTTPComponent() Component {
	return defaultComponents()[0]
}
