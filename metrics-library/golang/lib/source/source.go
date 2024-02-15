package source

var (
	KurtosisCLISource    = Source{"kurtosis-cli"}
	KurtosisEngineSource = Source{"kurtosis-engine"}
	KurtosisCoreSource   = Source{"kurtosis-core"}
)

// We declare Source as a struct to protect the value, so clients of this library can not create
// their own implementation of Source because key is a private key and there is not constructor
// It's called as Struct-based Enums, can se more here: https://threedots.tech/post/safer-enums-in-go/
type Source struct {
	key string
}

func (src *Source) GetKey() string {
	return src.key
}
