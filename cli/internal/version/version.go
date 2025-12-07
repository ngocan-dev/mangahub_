package version

// CLIVersion is the current MangaHub CLI version. This value can be overridden at build time via -ldflags.
var CLIVersion = "1.3.0"

// BuildTime captures when the binary was built.
var BuildTime = "unknown"

// APICompatibility documents the supported backend API range.
var APICompatibility = "1.3.x"
