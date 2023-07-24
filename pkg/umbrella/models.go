package umbrella

type ResponseStatusMetaDestinations struct {
	Status Status        `json:"status"`
	Meta   Meta          `json:"meta"`
	Data   []Destination `json:"data"`
}

type ResponseStatusMetaDestiantionList struct {
	Status Status          `json:"status"`
	Meta   Meta            `json:"meta"`
	Data   DestinationList `json:"data"`
}

type ResponseStatusMetaDestinationLists struct {
	Status Status            `json:"status"`
	Meta   Meta              `json:"meta"`
	Data   []DestinationList `json:"data"`
}

type ResponseStatusDestinationList struct {
	Status Status          `json:"status"`
	Data   DestinationList `json:"data"`
}

type Status struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

type Meta struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

type Destination struct {
	ID          string `json:"id"`
	Destination string `json:"destination"`
	Type        string `json:"type"`
	Comment     string `json:"comment"`
	CreatedAt   string `json:"createdAt"`
}

type NewDestination struct {
	Destination string `json:"destination"`
}

type DestinationListMeta struct {
	DestinationCount int `json:"destinationCount"`
	DomainCount      int `json:"domainCount"`
	URLCount         int `json:"urlCount"`
	IPv4Count        int `json:"ipv4Count"`
	ApplicationCount int `json:"applicationCount"`
}

type DestinationList struct {
	ID                   int                 `json:"id"`
	OrganizationID       int                 `json:"organizatioNId"`
	Access               string              `json:"access"`
	IsGlobal             bool                `json:"isGlobal"`
	Name                 string              `json:"name"`
	ThirdpartyCategoryId string              `json:"thirdpartyCategoryId"`
	CreatedAt            int                 `json:"createdAt"`
	ModifiedAt           int                 `json:"modifiedAt"`
	IsMspDefault         bool                `json:"isMspDefault"`
	MarkedForDeletion    bool                `json:"markedForDeletion"`
	BundleTypeId         int                 `json:"bundleTypeId"`
	Meta                 DestinationListMeta `json:"meta"`
}
