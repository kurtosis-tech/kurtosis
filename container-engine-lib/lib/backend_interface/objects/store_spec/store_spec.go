package store_spec

type StoreSpec struct {
	name string
	src  string
}

func NewStoreSpec(src, name string) *StoreSpec {
	return &StoreSpec{
		name: name,
		src:  src,
	}
}

func (storeSpec *StoreSpec) GetName() string {
	return storeSpec.name
}

func (storeSpec *StoreSpec) GetSrc() string {
	return storeSpec.src
}

func (storeSpec *StoreSpec) SetName(name string) {
	storeSpec.name = name
}
