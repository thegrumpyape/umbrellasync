package umbrella

// URL Build Strings
const (
	authPath     = "https://%s/auth/%s"
	deployPath   = "https://%s/deployments/%s"
	adminPath    = "https://%s/admin/%s"
	policiesPath = "https://%s/policies/%s"
	reportsPath  = "https://%s/reports/%s"
)

// Auth Endpoints
const (
	endpointCreateAuthorizationToken = "%s/token"
)

// Policies Endpoints
const (
	endpointGetDestinationLists                   = "%s/destinationlists"
	endpointCreateDestinationList                 = "%s/destinationlists"
	endpointUpdateDestinationList                 = "%s/destinationlists/%d"
	endpointDeleteDestinationList                 = "%s/destinationlists/%d"
	endpointGetDestinationList                    = "%s/destinationlists/%d"
	endpointGetDestinationsInDestinationList      = "%s/destinationlists/%d/destinations"
	endpointAddDestinationsToDestinationList      = "%s/destinationlists/%d/destinations"
	endpointDeleteDestinationsFromDestinationList = "%s/destinationlists/%d/destinations/remove"
)
