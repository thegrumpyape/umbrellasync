package comparison

// Compares BlockFile with Destinations from DestinationList
func Compare(blocklistData []string, destinationListData []string) ([]string, []string) {
	var destsToAdd, destsToDelete []string

	blMap, dlMap := make(map[string]bool), make(map[string]bool)

	for _, item := range blocklistData {
		blMap[item] = true
	}

	for _, item := range destinationListData {
		dlMap[item] = true
	}

	for key := range blMap {
		if !dlMap[key] {
			destsToAdd = append(destsToAdd, key)
		}
	}

	for key := range dlMap {
		if !blMap[key] {
			destsToDelete = append(destsToDelete, key)
		}
	}

	return destsToAdd, destsToDelete
}
