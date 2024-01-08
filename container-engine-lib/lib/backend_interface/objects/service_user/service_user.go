package service_user

type UID int64
type GID int64

const (
	defaultGidValue = -1
	gidIsNotSet     = false
)

type ServiceUser struct {
	uid      UID
	gid      GID
	isGIDSet bool
}

func NewServiceUser(uid UID) *ServiceUser {
	return &ServiceUser{uid: uid, gid: defaultGidValue, isGIDSet: gidIsNotSet}
}

func (su *ServiceUser) GetUID() UID {
	return su.uid
}

func (su *ServiceUser) GetGID() GID {
	return su.gid
}

func (su *ServiceUser) SetGID(gid GID) {
	su.gid = gid
	su.isGIDSet = true
}
