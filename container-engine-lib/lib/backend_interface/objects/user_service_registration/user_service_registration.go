package user_service_registration

import "net"

type UserServiceRegistration struct {
	id ServiceID
	ipAddress net.IP
}

func (registration *UserServiceRegistration) GetID() ServiceID {
	return registration.id
}
func (registration *UserServiceRegistration) GetIPAddress() net.IP {
	return registration.ipAddress
}