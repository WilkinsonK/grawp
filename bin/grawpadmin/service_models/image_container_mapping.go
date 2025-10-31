package service_models

import "github.com/google/uuid"

type ServiceImageContainerMapping struct {
	ContainerUuid uuid.UUID
	ImageUuid     uuid.UUID
}

func NewServiceImageContainerMapping(containerUuid uuid.UUID, imageUuid uuid.UUID) ServiceImageContainerMapping {
	return ServiceImageContainerMapping{
		ContainerUuid: containerUuid,
		ImageUuid:     imageUuid,
	}
}
