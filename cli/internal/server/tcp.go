package server

// TCPComponent returns metadata for the TCP Sync server.
func TCPComponent() Component {
	return defaultComponents()[1]
}
