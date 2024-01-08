package service_user

import (
	"fmt"
)

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

func (su *ServiceUser) GetGID() (GID, bool) {
	if !su.isGIDSet {
		return 0, false
	}
	return su.gid, true
}

func (su *ServiceUser) SetGID(gid GID) {
	su.gid = gid
	su.isGIDSet = true
}

func (su *ServiceUser) GetUIDGIDPairAsStr() string {
	if su.isGIDSet {
		return fmt.Sprintf("%v:%v", su.uid, su.gid)
	}
	return fmt.Sprintf("%v", su.uid)
}
