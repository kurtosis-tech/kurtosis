// Code generated by mockery v2.29.0. DO NOT EDIT.

package startosis_packages

import (
	io "io"

	startosis_errors "github.com/kurtosis-tech/kurtosis/core/server/api_container/server/startosis_engine/startosis_errors"
	mock "github.com/stretchr/testify/mock"
)

// MockPackageContentProvider is an autogenerated mock type for the PackageContentProvider type
type MockPackageContentProvider struct {
	mock.Mock
}

type MockPackageContentProvider_Expecter struct {
	mock *mock.Mock
}

func (_m *MockPackageContentProvider) EXPECT() *MockPackageContentProvider_Expecter {
	return &MockPackageContentProvider_Expecter{mock: &_m.Mock}
}

// ClonePackage provides a mock function with given fields: packageId
func (_m *MockPackageContentProvider) ClonePackage(packageId string) (string, string, *startosis_errors.InterpretationError) {
	ret := _m.Called(packageId)

	var r0 string
	var r1 string
	var r2 *startosis_errors.InterpretationError
	if rf, ok := ret.Get(0).(func(string) (string, string, *startosis_errors.InterpretationError)); ok {
		return rf(packageId)
	}
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(packageId)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string) string); ok {
		r1 = rf(packageId)
	} else {
		r1 = ret.Get(1).(string)
	}

	if rf, ok := ret.Get(2).(func(string) *startosis_errors.InterpretationError); ok {
		r2 = rf(packageId)
	} else {
		if ret.Get(2) != nil {
			r2 = ret.Get(2).(*startosis_errors.InterpretationError)
		}
	}

	return r0, r1, r2
}

// MockPackageContentProvider_ClonePackage_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'ClonePackage'
type MockPackageContentProvider_ClonePackage_Call struct {
	*mock.Call
}

// ClonePackage is a helper method to define mock.On call
//   - packageId string
func (_e *MockPackageContentProvider_Expecter) ClonePackage(packageId interface{}) *MockPackageContentProvider_ClonePackage_Call {
	return &MockPackageContentProvider_ClonePackage_Call{Call: _e.mock.On("ClonePackage", packageId)}
}

func (_c *MockPackageContentProvider_ClonePackage_Call) Run(run func(packageId string)) *MockPackageContentProvider_ClonePackage_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockPackageContentProvider_ClonePackage_Call) Return(_a0 string, _a1 string, _a2 *startosis_errors.InterpretationError) *MockPackageContentProvider_ClonePackage_Call {
	_c.Call.Return(_a0, _a1, _a2)
	return _c
}

func (_c *MockPackageContentProvider_ClonePackage_Call) RunAndReturn(run func(string) (string, string, *startosis_errors.InterpretationError)) *MockPackageContentProvider_ClonePackage_Call {
	_c.Call.Return(run)
	return _c
}

// GetAbsoluteLocatorForRelativeModuleLocator provides a mock function with given fields: packageId, relativeOrAbsoluteModulePath
func (_m *MockPackageContentProvider) GetAbsoluteLocatorForRelativeModuleLocator(packageId string, relativeOrAbsoluteModulePath string) (string, *startosis_errors.InterpretationError) {
	ret := _m.Called(packageId, relativeOrAbsoluteModulePath)

	var r0 string
	var r1 *startosis_errors.InterpretationError
	if rf, ok := ret.Get(0).(func(string, string) (string, *startosis_errors.InterpretationError)); ok {
		return rf(packageId, relativeOrAbsoluteModulePath)
	}
	if rf, ok := ret.Get(0).(func(string, string) string); ok {
		r0 = rf(packageId, relativeOrAbsoluteModulePath)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string, string) *startosis_errors.InterpretationError); ok {
		r1 = rf(packageId, relativeOrAbsoluteModulePath)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*startosis_errors.InterpretationError)
		}
	}

	return r0, r1
}

// MockPackageContentProvider_GetAbsoluteLocatorForRelativeModuleLocator_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAbsoluteLocatorForRelativeModuleLocator'
type MockPackageContentProvider_GetAbsoluteLocatorForRelativeModuleLocator_Call struct {
	*mock.Call
}

// GetAbsoluteLocatorForRelativeModuleLocator is a helper method to define mock.On call
//   - packageId string
//   - relativeOrAbsoluteModulePath string
func (_e *MockPackageContentProvider_Expecter) GetAbsoluteLocatorForRelativeModuleLocator(packageId interface{}, relativeOrAbsoluteModulePath interface{}) *MockPackageContentProvider_GetAbsoluteLocatorForRelativeModuleLocator_Call {
	return &MockPackageContentProvider_GetAbsoluteLocatorForRelativeModuleLocator_Call{Call: _e.mock.On("GetAbsoluteLocatorForRelativeModuleLocator", packageId, relativeOrAbsoluteModulePath)}
}

func (_c *MockPackageContentProvider_GetAbsoluteLocatorForRelativeModuleLocator_Call) Run(run func(packageId string, relativeOrAbsoluteModulePath string)) *MockPackageContentProvider_GetAbsoluteLocatorForRelativeModuleLocator_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *MockPackageContentProvider_GetAbsoluteLocatorForRelativeModuleLocator_Call) Return(_a0 string, _a1 *startosis_errors.InterpretationError) *MockPackageContentProvider_GetAbsoluteLocatorForRelativeModuleLocator_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockPackageContentProvider_GetAbsoluteLocatorForRelativeModuleLocator_Call) RunAndReturn(run func(string, string) (string, *startosis_errors.InterpretationError)) *MockPackageContentProvider_GetAbsoluteLocatorForRelativeModuleLocator_Call {
	_c.Call.Return(run)
	return _c
}

// GetModuleContents provides a mock function with given fields: fileInsidePackageUrl
func (_m *MockPackageContentProvider) GetModuleContents(fileInsidePackageUrl string) (string, *startosis_errors.InterpretationError) {
	ret := _m.Called(fileInsidePackageUrl)

	var r0 string
	var r1 *startosis_errors.InterpretationError
	if rf, ok := ret.Get(0).(func(string) (string, *startosis_errors.InterpretationError)); ok {
		return rf(fileInsidePackageUrl)
	}
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(fileInsidePackageUrl)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string) *startosis_errors.InterpretationError); ok {
		r1 = rf(fileInsidePackageUrl)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*startosis_errors.InterpretationError)
		}
	}

	return r0, r1
}

// MockPackageContentProvider_GetModuleContents_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetModuleContents'
type MockPackageContentProvider_GetModuleContents_Call struct {
	*mock.Call
}

// GetModuleContents is a helper method to define mock.On call
//   - fileInsidePackageUrl string
func (_e *MockPackageContentProvider_Expecter) GetModuleContents(fileInsidePackageUrl interface{}) *MockPackageContentProvider_GetModuleContents_Call {
	return &MockPackageContentProvider_GetModuleContents_Call{Call: _e.mock.On("GetModuleContents", fileInsidePackageUrl)}
}

func (_c *MockPackageContentProvider_GetModuleContents_Call) Run(run func(fileInsidePackageUrl string)) *MockPackageContentProvider_GetModuleContents_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockPackageContentProvider_GetModuleContents_Call) Return(_a0 string, _a1 *startosis_errors.InterpretationError) *MockPackageContentProvider_GetModuleContents_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockPackageContentProvider_GetModuleContents_Call) RunAndReturn(run func(string) (string, *startosis_errors.InterpretationError)) *MockPackageContentProvider_GetModuleContents_Call {
	_c.Call.Return(run)
	return _c
}

// GetOnDiskAbsoluteFilePath provides a mock function with given fields: fileInsidePackageUrl
func (_m *MockPackageContentProvider) GetOnDiskAbsoluteFilePath(fileInsidePackageUrl string) (string, *startosis_errors.InterpretationError) {
	ret := _m.Called(fileInsidePackageUrl)

	var r0 string
	var r1 *startosis_errors.InterpretationError
	if rf, ok := ret.Get(0).(func(string) (string, *startosis_errors.InterpretationError)); ok {
		return rf(fileInsidePackageUrl)
	}
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(fileInsidePackageUrl)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string) *startosis_errors.InterpretationError); ok {
		r1 = rf(fileInsidePackageUrl)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*startosis_errors.InterpretationError)
		}
	}

	return r0, r1
}

// MockPackageContentProvider_GetOnDiskAbsoluteFilePath_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetOnDiskAbsoluteFilePath'
type MockPackageContentProvider_GetOnDiskAbsoluteFilePath_Call struct {
	*mock.Call
}

// GetOnDiskAbsoluteFilePath is a helper method to define mock.On call
//   - fileInsidePackageUrl string
func (_e *MockPackageContentProvider_Expecter) GetOnDiskAbsoluteFilePath(fileInsidePackageUrl interface{}) *MockPackageContentProvider_GetOnDiskAbsoluteFilePath_Call {
	return &MockPackageContentProvider_GetOnDiskAbsoluteFilePath_Call{Call: _e.mock.On("GetOnDiskAbsoluteFilePath", fileInsidePackageUrl)}
}

func (_c *MockPackageContentProvider_GetOnDiskAbsoluteFilePath_Call) Run(run func(fileInsidePackageUrl string)) *MockPackageContentProvider_GetOnDiskAbsoluteFilePath_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockPackageContentProvider_GetOnDiskAbsoluteFilePath_Call) Return(_a0 string, _a1 *startosis_errors.InterpretationError) *MockPackageContentProvider_GetOnDiskAbsoluteFilePath_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockPackageContentProvider_GetOnDiskAbsoluteFilePath_Call) RunAndReturn(run func(string) (string, *startosis_errors.InterpretationError)) *MockPackageContentProvider_GetOnDiskAbsoluteFilePath_Call {
	_c.Call.Return(run)
	return _c
}

// GetOnDiskAbsolutePackagePath provides a mock function with given fields: packageId
func (_m *MockPackageContentProvider) GetOnDiskAbsolutePackagePath(packageId string) (string, *startosis_errors.InterpretationError) {
	ret := _m.Called(packageId)

	var r0 string
	var r1 *startosis_errors.InterpretationError
	if rf, ok := ret.Get(0).(func(string) (string, *startosis_errors.InterpretationError)); ok {
		return rf(packageId)
	}
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(packageId)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string) *startosis_errors.InterpretationError); ok {
		r1 = rf(packageId)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*startosis_errors.InterpretationError)
		}
	}

	return r0, r1
}

// MockPackageContentProvider_GetOnDiskAbsolutePackagePath_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetOnDiskAbsolutePackagePath'
type MockPackageContentProvider_GetOnDiskAbsolutePackagePath_Call struct {
	*mock.Call
}

// GetOnDiskAbsolutePackagePath is a helper method to define mock.On call
//   - packageId string
func (_e *MockPackageContentProvider_Expecter) GetOnDiskAbsolutePackagePath(packageId interface{}) *MockPackageContentProvider_GetOnDiskAbsolutePackagePath_Call {
	return &MockPackageContentProvider_GetOnDiskAbsolutePackagePath_Call{Call: _e.mock.On("GetOnDiskAbsolutePackagePath", packageId)}
}

func (_c *MockPackageContentProvider_GetOnDiskAbsolutePackagePath_Call) Run(run func(packageId string)) *MockPackageContentProvider_GetOnDiskAbsolutePackagePath_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string))
	})
	return _c
}

func (_c *MockPackageContentProvider_GetOnDiskAbsolutePackagePath_Call) Return(_a0 string, _a1 *startosis_errors.InterpretationError) *MockPackageContentProvider_GetOnDiskAbsolutePackagePath_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockPackageContentProvider_GetOnDiskAbsolutePackagePath_Call) RunAndReturn(run func(string) (string, *startosis_errors.InterpretationError)) *MockPackageContentProvider_GetOnDiskAbsolutePackagePath_Call {
	_c.Call.Return(run)
	return _c
}

// StorePackageContents provides a mock function with given fields: packageId, packageContent, overwriteExisting
func (_m *MockPackageContentProvider) StorePackageContents(packageId string, packageContent io.Reader, overwriteExisting bool) (string, *startosis_errors.InterpretationError) {
	ret := _m.Called(packageId, packageContent, overwriteExisting)

	var r0 string
	var r1 *startosis_errors.InterpretationError
	if rf, ok := ret.Get(0).(func(string, io.Reader, bool) (string, *startosis_errors.InterpretationError)); ok {
		return rf(packageId, packageContent, overwriteExisting)
	}
	if rf, ok := ret.Get(0).(func(string, io.Reader, bool) string); ok {
		r0 = rf(packageId, packageContent, overwriteExisting)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string, io.Reader, bool) *startosis_errors.InterpretationError); ok {
		r1 = rf(packageId, packageContent, overwriteExisting)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*startosis_errors.InterpretationError)
		}
	}

	return r0, r1
}

// MockPackageContentProvider_StorePackageContents_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'StorePackageContents'
type MockPackageContentProvider_StorePackageContents_Call struct {
	*mock.Call
}

// StorePackageContents is a helper method to define mock.On call
//   - packageId string
//   - packageContent io.Reader
//   - overwriteExisting bool
func (_e *MockPackageContentProvider_Expecter) StorePackageContents(packageId interface{}, packageContent interface{}, overwriteExisting interface{}) *MockPackageContentProvider_StorePackageContents_Call {
	return &MockPackageContentProvider_StorePackageContents_Call{Call: _e.mock.On("StorePackageContents", packageId, packageContent, overwriteExisting)}
}

func (_c *MockPackageContentProvider_StorePackageContents_Call) Run(run func(packageId string, packageContent io.Reader, overwriteExisting bool)) *MockPackageContentProvider_StorePackageContents_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(io.Reader), args[2].(bool))
	})
	return _c
}

func (_c *MockPackageContentProvider_StorePackageContents_Call) Return(_a0 string, _a1 *startosis_errors.InterpretationError) *MockPackageContentProvider_StorePackageContents_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockPackageContentProvider_StorePackageContents_Call) RunAndReturn(run func(string, io.Reader, bool) (string, *startosis_errors.InterpretationError)) *MockPackageContentProvider_StorePackageContents_Call {
	_c.Call.Return(run)
	return _c
}

type mockConstructorTestingTNewMockPackageContentProvider interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockPackageContentProvider creates a new instance of MockPackageContentProvider. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockPackageContentProvider(t mockConstructorTestingTNewMockPackageContentProvider) *MockPackageContentProvider {
	mock := &MockPackageContentProvider{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
