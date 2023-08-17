package umbrellasync

import (
	"log"

	"github.com/thegrumpyape/umbrellasync/pkg/umbrella"
)

type UmbrellaSync struct {
	client           umbrella.UmbrellaService
	destinationLists []umbrella.DestinationList
	logger           *log.Logger
}
