package service_user

import (
	"encoding/json"
	"fmt"
	"github.com/kurtosis-tech/stacktrace"
)

type UID int64
type GID int64

const (
	defaultGidValue = -1
	gidIsNotSet     = false
)

type ServiceUser struct {
	privateServiceUser *privateServiceUser
}

type privateServiceUser struct {
	UID      UID
	GID      GID
	IsGIDSet bool
}

func NewServiceUser(uid UID) *ServiceUser {
	internalServiceUser := &privateServiceUser{
		UID:      uid,
		GID:      defaultGidValue,
		IsGIDSet: gidIsNotSet,
	}
	return &ServiceUser{privateServiceUser: internalServiceUser}
}

func (su *ServiceUser) GetUID() UID {
	return su.privateServiceUser.UID
}

func (su *ServiceUser) GetGID() (GID, bool) {
	if !su.privateServiceUser.IsGIDSet {
		return 0, false
	}
	return su.privateServiceUser.GID, true
}

func (su *ServiceUser) SetGID(gid GID) {
	su.privateServiceUser.GID = gid
	su.privateServiceUser.IsGIDSet = true
}

func (su *ServiceUser) GetUIDGIDPairAsStr() string {
	if su.privateServiceUser.IsGIDSet {
		return fmt.Sprintf("%v:%v", su.privateServiceUser.UID, su.privateServiceUser.GID)
	}
	return fmt.Sprintf("%v", su.privateServiceUser.UID)
}

func (su ServiceUser) MarshalJSON() ([]byte, error) {
	return json.Marshal(su.privateServiceUser)
}

func (su *ServiceUser) UnmarshalJSON(data []byte) error {

	// Suppressing exhaustruct requirement because we want an object with zero values
	// nolint: exhaustruct
	unmarshalledPrivateStructPtr := &privateServiceUser{}

	if err := json.Unmarshal(data, unmarshalledPrivateStructPtr); err != nil {
		return stacktrace.Propagate(err, "An error occurred unmarshalling the private struct")
	}

	su.privateServiceUser = unmarshalledPrivateStructPtr
	return nil
}
