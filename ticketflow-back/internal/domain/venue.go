package domain

// Venue is a physical place (a hall) where events take place.
type Venue struct {
	ID      string
	Name    string
	Address string
	// Schema holds the seating plan as raw JSON.
	// Parsing it is not a domain concern.
	Schema string
}
