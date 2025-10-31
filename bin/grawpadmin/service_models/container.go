package service_models

import "github.com/google/uuid"

type ServiceContainer struct {
	Uuid        uuid.UUID
	Name        string
	DockerId    string
	IsAvailable bool
}

type NewServiceContainerOpts struct {
	Uuid     uuid.UUID
	Name     string
	DockerId string
}

func NewServiceContainer(opts NewServiceContainerOpts) (ServiceContainer, error) {
	validateUuidOrGenerateNewUuid(&opts.Uuid)

	var sc ServiceContainer
	sc.Uuid = opts.Uuid
	sc.Name = opts.Name
	sc.DockerId = opts.DockerId
	sc.IsAvailable = false
	return sc, nil
}
