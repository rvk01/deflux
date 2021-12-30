package sensor

// State contains properties that are provided by all sensors
// It is embedded in specific sensors' State
type State struct {
	Lastupdated string
}

// EmptyState is an empty struct used to indicate no state was parsed
type EmptyState struct{}
