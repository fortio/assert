// Copyright 2022 Fortio Authors

// Minimalistic drop in eplacement for github.com/stretchr/testify/assert
package assert // import "fortio.org/assert"

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// ObjectsAreEqualValues returns true if a == b (through refection).
func ObjectsAreEqualValues(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

// Errorf is a local variant to get the right line numbers.
func Errorf(t *testing.T, format string, rest ...interface{}) {
	_, file, line, _ := runtime.Caller(2)
	file = file[strings.LastIndex(file, "/")+1:]
	fmt.Printf("%s:%d %s", file, line, fmt.Sprintf(format, rest...))
	t.Fail()
}

// NotEqual checks for a not equal b.
func NotEqual(t *testing.T, a, b interface{}, msg ...string) {
	if ObjectsAreEqualValues(a, b) {
		Errorf(t, "%v unexpectedly equal: %v", a, msg)
	}
}

// EqualValues checks for a equal b.
func EqualValues(t *testing.T, a, b interface{}, msg ...string) {
	if !ObjectsAreEqualValues(a, b) {
		Errorf(t, "%v unexpectedly not equal %v: %v", a, b, msg)
	}
}

// Equal also checks for a equal b.
func Equal(t *testing.T, a, b interface{}, msg ...string) {
	EqualValues(t, a, b, msg...)
}

// NoError checks for no errors (nil).
func NoError(t *testing.T, err error, msg ...string) {
	if err != nil {
		Errorf(t, "expecting no error, got %v: %v", err, msg)
	}
}

// Error checks/expects an error.
func Error(t *testing.T, err error, msg ...string) {
	if err == nil {
		Errorf(t, "expecting an error, didn't get it: %v", msg)
	}
}

// True checks bool is true.
func True(t *testing.T, b bool, msg ...string) {
	if !b {
		Errorf(t, "expecting true, didn't: %v", msg)
	}
}

// False checks bool is false.
func False(t *testing.T, b bool, msg ...string) {
	if b {
		Errorf(t, "expecting false, didn't: %v", msg)
	}
}

// Contains checks that needle is in haystack.
func Contains(t *testing.T, haystack, needle string, msg ...string) {
	if !strings.Contains(haystack, needle) {
		Errorf(t, "%v doesn't contain %v: %v", haystack, needle, msg)
	}
}

// Fail fails the test.
func Fail(t *testing.T, msg string) {
	_, file, line, _ := runtime.Caller(1)
	file = file[strings.LastIndex(file, "/")+1:]
	fmt.Printf("%s:%d %s\n", file, line, msg)
	t.FailNow()
}

// CheckEquals checks if actual == expect and fails the test and logs
// failure (including filename:linenum if they are not equal).
func CheckEquals(t *testing.T, actual interface{}, expected interface{}, msg interface{}) {
	if expected != actual {
		_, file, line, _ := runtime.Caller(1)
		file = file[strings.LastIndex(file, "/")+1:]
		fmt.Printf("%s:%d mismatch!\nactual:\n%+v\nexpected:\n%+v\nfor %+v\n", file, line, actual, expected, msg)
		t.Fail()
	}
}

// Assert is similar to True() under a different name and earlier
// fortio/stats test implementation.
func Assert(t *testing.T, cond bool, msg interface{}) {
	if !cond {
		_, file, line, _ := runtime.Caller(1)
		file = file[strings.LastIndex(file, "/")+1:]
		fmt.Printf("%s:%d assert failure: %+v\n", file, line, msg)
		t.Fail()
	}
}

type hasT interface {
	T() *testing.T
	SetT(*testing.T)
}

// TestSuite to be used as base struct for test suites.
// replaces https://pkg.go.dev/github.com/stretchr/testify@v1.8.0/suite
type TestSuite struct {
	t *testing.T
}

// T returns the current testing.T.
func (s *TestSuite) T() *testing.T {
	return s.t
}

// SetT sets the testing.T in the suite object.
func (s *TestSuite) SetT(t *testing.T) {
	s.t = t
}

type hasSetupTest interface {
	SetupTest()
}
type hasTearDown interface {
	TearDownTest()
}

// Run runs the test suite with SetupTest first and TearDownTest after.
// replaces https://pkg.go.dev/github.com/stretchr/testify/suite#Run
func Run(t *testing.T, suite hasT) {
	suite.SetT(t)
	tests := []testing.InternalTest{}
	methodFinder := reflect.TypeOf(suite)
	var setup hasSetupTest
	if s, ok := suite.(hasSetupTest); ok {
		setup = s
	}
	var tearDown hasTearDown
	if td, ok := suite.(hasTearDown); ok {
		tearDown = td
	}
	for i := 0; i < methodFinder.NumMethod(); i++ {
		method := methodFinder.Method(i)
		//nolint:staticcheck // consider fixing later for perf but this is just to run a few tests.
		if ok, _ := regexp.MatchString("^Test", method.Name); !ok {
			continue
		}
		test := testing.InternalTest{
			Name: method.Name,
			F: func(t *testing.T) {
				method.Func.Call([]reflect.Value{reflect.ValueOf(suite)})
			},
		}
		tests = append(tests, test)
	}
	for _, test := range tests {
		if setup != nil {
			setup.SetupTest()
		}
		t.Run(test.Name, test.F)
		if tearDown != nil {
			tearDown.TearDownTest()
		}
	}
}
