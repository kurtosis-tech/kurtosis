package module

// Object that represents POINT-IN-TIME information about an Kurtosis Module
// Store this object and continue to reference it at your own risk!!!
type Module struct {
	id string
}

func NewModule(id string) *Module {
	return &Module{id: id}
}




