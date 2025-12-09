package server

// GRPCComponent returns metadata for the gRPC internal server.
func GRPCComponent() Component {
	return defaultComponents()[3]
}
