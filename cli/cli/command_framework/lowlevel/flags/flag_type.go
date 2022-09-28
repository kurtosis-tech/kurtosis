package flags

//go:generate go run github.com/dmarkham/enumer -type=FlagType -trimprefix=FlagType_ -transform=lower
type FlagType int
const (
	FlagType_String FlagType = iota  // This is intentionally the first value, meaning it will be the emptyval/default
	FlagType_Uint32
	FlagType_Bool
)
