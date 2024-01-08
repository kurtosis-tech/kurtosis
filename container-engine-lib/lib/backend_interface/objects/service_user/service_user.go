package service_user

type UID int64
type GID int64

type ServiceUser struct {
	uid UID
	gid GID
}

func NewServiceUser(uid UID, gid GID) *ServiceUser {
	return &ServiceUser{uid: uid, gid: gid}
}

func (su *ServiceUser) getUID() UID {
	return su.uid
}

func (su *ServiceUser) getGID() GID {
	return su.gid
}
