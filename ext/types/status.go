package types

// Extension status constants
const (
	// StatusActive indicates the extension is running normally
	StatusActive = "active"
	// StatusInactive indicates the extension is installed but not running
	StatusInactive = "inactive"
	// StatusError indicates the extension encountered an error
	StatusError = "error"
	// StatusInitializing indicates the extension is in initialization process
	StatusInitializing = "initializing"
	// StatusMaintenance indicates the extension is under maintenance
	StatusMaintenance = "maintenance"
	// StatusDisabled indicates the extension has been manually disabled
	StatusDisabled = "disabled"
)
