package service

type ServiceID string

// Object that represents POINT-IN-TIME information about an user service
// Store this object and continue to reference it at your own risk!!!
type Service struct {
	id ServiceID
}

func NewService(id ServiceID) *Service {
	return &Service{id: id}
}

func (service *Service) GetID() ServiceID {
	return service.id
}
