package sensor

// State is for embedding into event states
type State struct {
	Lastupdated string
}

// EmptyState is an empty struct used to indicate no state was parsed
type EmptyState struct{}
