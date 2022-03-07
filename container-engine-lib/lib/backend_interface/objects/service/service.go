package service

// Object that represents POINT-IN-TIME information about an user service
// Store this object and continue to reference it at your own risk!!!
type Service struct {
	id string
}

func NewService(id string) *Service {
	return &Service{id: id}
}

func (service *Service) GetID() string {
	return service.id
}
