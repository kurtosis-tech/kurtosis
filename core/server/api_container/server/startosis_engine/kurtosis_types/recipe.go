package kurtosis_types

//
//const (
//	recipeTypeName = "recipe"
//)
//
//type Recipe struct {
//	ipAddress starlark.String
//	ports     *starlark.Dict
//}
//
//func Recipe(ipAddress starlark.String, ports *starlark.Dict) *Recipe {
//	return &Service{
//		ipAddress: ipAddress,
//		ports:     ports,
//	}
//}
//
//// String the starlark.Value interface
//func (rv *Recipe) String() string {
//	buffer := new(strings.Builder)
//	buffer.WriteString(serviceTypeName + "(")
//	buffer.WriteString(ipAddressAttr + "=")
//	buffer.WriteString(fmt.Sprintf("%v, ", rv.ipAddress))
//	buffer.WriteString(portsAttr + "=")
//	buffer.WriteString(fmt.Sprintf("%v)", rv.ports.String()))
//	return buffer.String()
//}
//
//// Type implements the starlark.Value interface
//func (rv *Recipe) Type() string {
//	return serviceTypeName
//}
//
//// Freeze implements the starlark.Value interface
//func (rv *Recipe) Freeze() {
//	// this is a no-op its already immutable
//}
//
//// Truth implements the starlark.Value interface
//func (rv *Recipe) Truth() starlark.Bool {
//	return rv.ipAddress != "" && rv.ports != nil
//}
//
//// Hash implements the starlark.Value interface
//// This shouldn't be hashed, users should use a portId instead
//func (rv *Recipe) Hash() (uint32, error) {
//	return 0, fmt.Errorf("unhashable type: '%v'", serviceTypeName)
//}
//
//// Attr implements the starlark.HasAttrs interface.
//func (rv *Recipe) Attr(name string) (starlark.Value, error) {
//	switch name {
//	case ipAddressAttr:
//		return rv.ipAddress, nil
//	case portsAttr:
//		return rv.ports, nil
//	default:
//		return nil, fmt.Errorf("'%v' has no attribute '%v'", serviceTypeName, name)
//	}
//}
//
//// AttrNames implements the starlark.HasAttrs interface.
//func (rv *Recipe) AttrNames() []string {
//	return []string{ipAddressAttr, portsAttr}
//}
