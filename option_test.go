package optionalv2_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tapp-ai/go-optional-v2"
)

// Custom type for testing fmt.Stringer interface
type CustomType struct {
	ID   int
	Name string
}

func (c CustomType) String() string {
	return fmt.Sprintf("CustomType(ID=%d, Name=%s)", c.ID, c.Name)
}

func TestOption(t *testing.T) {
	// Test Some and None creation
	t.Run("Creation", func(t *testing.T) {
		optSome := optionalv2.Some(42)
		assert.True(t, optSome.IsSome())
		assert.False(t, optSome.IsNone())
		assert.Equal(t, 42, optSome.Unwrap())

		optNone := optionalv2.None[int]()
		assert.False(t, optNone.IsSome())
		assert.True(t, optNone.IsNone())
	})

	// Test Unwrap and UnwrapAsPtr methods
	t.Run("UnwrapMethods", func(t *testing.T) {
		optSome := optionalv2.Some("Hello")
		value := optSome.Unwrap()
		assert.Equal(t, "Hello", value)

		ptr := optSome.UnwrapAsPtr()
		assert.NotNil(t, ptr)
		assert.Equal(t, "Hello", *ptr)

		optNone := optionalv2.None[string]()
		value = optNone.Unwrap()
		assert.Equal(t, "", value) // Zero value for string

		ptr = optNone.UnwrapAsPtr()
		assert.Nil(t, ptr)
	})

	// Test Take and TakeOr methods
	t.Run("TakeMethods", func(t *testing.T) {
		optSome := optionalv2.Some(100)
		value, err := optSome.Take()
		assert.NoError(t, err)
		assert.Equal(t, 100, value)

		optNone := optionalv2.None[int]()
		value, err = optNone.Take()
		assert.Error(t, err)
		assert.Equal(t, optionalv2.ErrNoneValueTaken, err)
		assert.Equal(t, 0, value) // Zero value for int

		value = optNone.TakeOr(200)
		assert.Equal(t, 200, value)

		value = optSome.TakeOr(300)
		assert.Equal(t, 100, value) // Original value, since optSome is Some

		value = optNone.TakeOrElse(func() int {
			return 400
		})
		assert.Equal(t, 400, value)

		value = optSome.TakeOrElse(func() int {
			return 500
		})
		assert.Equal(t, 100, value) // Original value, since optSome is Some
	})

	// Test Or method
	t.Run("OrMethod", func(t *testing.T) {
		optNone := optionalv2.None[int]()
		optSome := optionalv2.Some(10)
		fallback := optionalv2.Some(20)

		result := optNone.Or(fallback)
		assert.True(t, result.IsSome())
		assert.Equal(t, 20, result.Unwrap())

		result = optSome.Or(fallback)
		assert.True(t, result.IsSome())
		assert.Equal(t, 10, result.Unwrap())
	})

	// Test Filter method
	t.Run("FilterMethod", func(t *testing.T) {
		opt := optionalv2.Some(15)
		result := opt.Filter(func(v int) bool {
			return v > 10
		})
		assert.True(t, result.IsSome())
		assert.Equal(t, 15, result.Unwrap())

		result = opt.Filter(func(v int) bool {
			return v < 10
		})
		assert.True(t, result.IsNone())
	})

	// Test IfSome and IfNone methods
	t.Run("IfSomeAndIfNone", func(t *testing.T) {
		optSome := optionalv2.Some("test")
		var called bool

		optSome.IfSome(func(v string) {
			called = true
			assert.Equal(t, "test", v)
		})
		assert.True(t, called)

		called = false
		optSome.IfNone(func() {
			called = true
		})
		assert.False(t, called)

		optNone := optionalv2.None[string]()
		called = false
		optNone.IfSome(func(v string) {
			called = true
		})
		assert.False(t, called)

		called = false
		optNone.IfNone(func() {
			called = true
		})
		assert.True(t, called)
	})

	// Test IfSomeWithError and IfNoneWithError
	t.Run("IfSomeWithErrorAndIfNoneWithError", func(t *testing.T) {
		optSome := optionalv2.Some(5)
		err := optSome.IfSomeWithError(func(v int) error {
			if v < 10 {
				return errors.New("value is less than 10")
			}
			return nil
		})
		assert.Error(t, err)
		assert.Equal(t, "value is less than 10", err.Error())

		optNone := optionalv2.None[int]()
		err = optNone.IfSomeWithError(func(v int) error {
			return errors.New("should not be called")
		})
		assert.NoError(t, err)

		err = optNone.IfNoneWithError(func() error {
			return errors.New("no value present")
		})
		assert.Error(t, err)
		assert.Equal(t, "no value present", err.Error())

		optSome = optionalv2.Some(20)
		err = optSome.IfNoneWithError(func() error {
			return errors.New("should not be called")
		})
		assert.NoError(t, err)
	})

	// Test String method
	t.Run("StringMethod", func(t *testing.T) {
		optSome := optionalv2.Some(42)
		assert.Equal(t, "Some[42]", optSome.String())

		optNone := optionalv2.None[int]()
		assert.Equal(t, "None[]", optNone.String())

		optCustom := optionalv2.Some(CustomType{ID: 1, Name: "Test"})
		assert.Equal(t, "Some[CustomType(ID=1, Name=Test)]", optCustom.String())

		optCustomNone := optionalv2.None[CustomType]()
		assert.Equal(t, "None[]", optCustomNone.String())
	})

	// Test JSON marshalling and unmarshalling
	t.Run("JSONMarshalling", func(t *testing.T) {
		type TestStruct struct {
			Value optionalv2.Option[int] `json:"value,omitempty"`
		}

		// Test marshalling Some with non-zero value
		s := TestStruct{
			Value: optionalv2.Some(10),
		}
		data, err := json.Marshal(s)
		assert.NoError(t, err)
		assert.JSONEq(t, `{"value":10}`, string(data))

		// Test marshalling Some with zero value (should be null)
		s.Value = optionalv2.Some(0)
		data, err = json.Marshal(s)
		assert.NoError(t, err)
		assert.JSONEq(t, `{"value":null}`, string(data))

		// Test marshalling None (should be omitted)
		s.Value = optionalv2.None[int]()
		data, err = json.Marshal(s)
		assert.NoError(t, err)
		assert.JSONEq(t, `{}`, string(data))

		// Test unmarshalling with value
		jsonStr := `{"value": 20}`
		err = json.Unmarshal([]byte(jsonStr), &s)
		assert.NoError(t, err)
		assert.True(t, s.Value.IsSome())
		assert.Equal(t, 20, s.Value.Unwrap())

		// Test unmarshalling with null
		jsonStr = `{"value": null}`
		err = json.Unmarshal([]byte(jsonStr), &s)
		assert.NoError(t, err)
		assert.True(t, s.Value.IsSome())
		assert.Equal(t, 0, s.Value.Unwrap())

		// Test unmarshalling with missing field
		s = TestStruct{}
		jsonStr = `{}`
		err = json.Unmarshal([]byte(jsonStr), &s)
		assert.NoError(t, err)
		assert.True(t, s.Value.IsNone())
	})

	// Test JSON marshalling and unmarshalling with time.Time
	t.Run("JSONTimeMarshalling", func(t *testing.T) {
		type TestStruct struct {
			TimeValue optionalv2.Option[time.Time] `json:"timeValue,omitempty"`
		}

		// Test with value
		s := TestStruct{
			TimeValue: optionalv2.Some(time.Date(2024, 9, 13, 0, 0, 0, 0, time.UTC)),
		}
		data, err := json.Marshal(s)
		assert.NoError(t, err)
		assert.JSONEq(t, `{"timeValue":"2024-09-13T00:00:00Z"}`, string(data))

		// Test with null value
		s.TimeValue = optionalv2.Some(time.Time{})
		data, err = json.Marshal(s)
		assert.NoError(t, err)
		assert.JSONEq(t, `{"timeValue":null}`, string(data))

		// Test with missing field
		s = TestStruct{}
		jsonStr := `{}`
		err = json.Unmarshal([]byte(jsonStr), &s)
		assert.NoError(t, err)
		assert.True(t, s.TimeValue.IsNone())

		// Test unmarshalling with value
		jsonStr = `{"timeValue": "2024-09-13T00:00:00Z"}`
		err = json.Unmarshal([]byte(jsonStr), &s)
		assert.NoError(t, err)
		assert.True(t, s.TimeValue.IsSome())
		assert.Equal(t, time.Date(2024, 9, 13, 0, 0, 0, 0, time.UTC), s.TimeValue.Unwrap())

		// Test unmarshalling with null
		jsonStr = `{"timeValue": null}`
		err = json.Unmarshal([]byte(jsonStr), &s)
		assert.NoError(t, err)
		assert.True(t, s.TimeValue.IsSome())
		assert.Equal(t, time.Time{}, s.TimeValue.Unwrap())
	})

	// Test edge cases with zero values
	t.Run("ZeroValues", func(t *testing.T) {
		opt := optionalv2.Some(0)
		assert.True(t, opt.IsSome())
		data, err := json.Marshal(opt)
		assert.NoError(t, err)
		assert.Equal(t, "null", string(data))

		var optUnmarshalled optionalv2.Option[int]
		err = json.Unmarshal([]byte("null"), &optUnmarshalled)
		assert.NoError(t, err)
		assert.True(t, optUnmarshalled.IsSome())
		assert.Equal(t, 0, optUnmarshalled.Unwrap())
	})

	// Test with time.Time type
	t.Run("TimeType", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Second)
		optTime := optionalv2.Some(now)
		assert.True(t, optTime.IsSome())
		assert.Equal(t, now, optTime.Unwrap())

		data, err := json.Marshal(optTime)
		assert.NoError(t, err)
		expectedJSON, _ := json.Marshal(now)
		assert.Equal(t, string(expectedJSON), string(data))

		var optTimeUnmarshalled optionalv2.Option[time.Time]
		err = json.Unmarshal(data, &optTimeUnmarshalled)
		assert.NoError(t, err)
		assert.True(t, optTimeUnmarshalled.IsSome())
		assert.Equal(t, now, optTimeUnmarshalled.Unwrap())
	})

	// Test with pointer types
	t.Run("PointerTypes", func(t *testing.T) {
		value := 10
		opt := optionalv2.Some(&value)
		assert.True(t, opt.IsSome())
		assert.Equal(t, &value, opt.Unwrap())

		data, err := json.Marshal(opt)
		assert.NoError(t, err)
		expectedJSON, _ := json.Marshal(&value)
		assert.Equal(t, string(expectedJSON), string(data))

		var optUnmarshalled optionalv2.Option[*int]
		err = json.Unmarshal(data, &optUnmarshalled)
		assert.NoError(t, err)
		assert.True(t, optUnmarshalled.IsSome())
		assert.Equal(t, &value, optUnmarshalled.Unwrap())
	})

	// Test nil pointer
	t.Run("NilPointer", func(t *testing.T) {
		var ptr *int = nil
		opt := optionalv2.Some(ptr)
		assert.True(t, opt.IsSome())
		assert.Nil(t, opt.Unwrap())

		data, err := json.Marshal(opt)
		assert.NoError(t, err)
		assert.Equal(t, "null", string(data))

		var optUnmarshalled optionalv2.Option[*int]
		err = json.Unmarshal([]byte("null"), &optUnmarshalled)
		assert.NoError(t, err)
		assert.True(t, optUnmarshalled.IsSome())
		assert.Nil(t, optUnmarshalled.Unwrap())
	})

	// Test with complex struct
	t.Run("ComplexStruct", func(t *testing.T) {
		type NestedStruct struct {
			ID   int
			Name string
		}

		type TestStruct struct {
			Data optionalv2.Option[NestedStruct] `json:"data,omitempty"`
		}

		// Test with value
		s := TestStruct{
			Data: optionalv2.Some(NestedStruct{ID: 1, Name: "Nested"}),
		}
		data, err := json.Marshal(s)
		assert.NoError(t, err)
		assert.JSONEq(t, `{"data": {"ID":1, "Name":"Nested"}}`, string(data))

		// Test unmarshalling
		jsonStr := `{"data": {"ID":2, "Name":"Unmarshalled"}}`
		err = json.Unmarshal([]byte(jsonStr), &s)
		assert.NoError(t, err)
		assert.True(t, s.Data.IsSome())
		assert.Equal(t, NestedStruct{ID: 2, Name: "Unmarshalled"}, s.Data.Unwrap())

		// Test with null value
		jsonStr = `{"data": null}`
		err = json.Unmarshal([]byte(jsonStr), &s)
		assert.NoError(t, err)
		assert.True(t, s.Data.IsSome())
		assert.Equal(t, NestedStruct{}, s.Data.Unwrap())

		// Test with missing field
		s = TestStruct{}
		jsonStr = `{}`
		err = json.Unmarshal([]byte(jsonStr), &s)
		assert.NoError(t, err)
		assert.True(t, s.Data.IsNone())
	})

	// Test isNull method (private method, but behavior can be tested)
	t.Run("IsNullBehavior", func(t *testing.T) {
		// Since isNull is a private method, we test its effect via JSON marshalling
		opt := optionalv2.Some(0)
		data, err := json.Marshal(opt)
		assert.NoError(t, err)
		assert.Equal(t, "null", string(data)) // Zero value treated as null

		opt = optionalv2.Some(1)
		data, err = json.Marshal(opt)
		assert.NoError(t, err)
		assert.Equal(t, "1", string(data))

		// Note: This is considered undefined behavior, as the `omitempty` tag is not used
		optNone := optionalv2.None[int]()
		data, err = json.Marshal(optNone)
		assert.NoError(t, err)
		assert.Equal(t, "0", string(data)) // None is marshalled as the zero value when not in a struct

		// Unmarshalling null into Option[int]
		var optUnmarshalled optionalv2.Option[int]
		err = json.Unmarshal([]byte("null"), &optUnmarshalled)
		assert.NoError(t, err)
		assert.True(t, optUnmarshalled.IsSome())
		assert.Equal(t, 0, optUnmarshalled.Unwrap())
	})

	// Test behavior with slices
	t.Run("Slices", func(t *testing.T) {
		opt := optionalv2.Some([]int{1, 2, 3})
		assert.True(t, opt.IsSome())
		assert.Equal(t, []int{1, 2, 3}, opt.Unwrap())

		data, err := json.Marshal(opt)
		assert.NoError(t, err)
		assert.JSONEq(t, `[1,2,3]`, string(data))

		var optUnmarshalled optionalv2.Option[[]int]
		err = json.Unmarshal(data, &optUnmarshalled)
		assert.NoError(t, err)
		assert.True(t, optUnmarshalled.IsSome())
		assert.Equal(t, []int{1, 2, 3}, optUnmarshalled.Unwrap())
	})

	// Test behavior with maps
	t.Run("Maps", func(t *testing.T) {
		opt := optionalv2.Some(map[string]int{"one": 1, "two": 2})
		assert.True(t, opt.IsSome())
		assert.Equal(t, map[string]int{"one": 1, "two": 2}, opt.Unwrap())

		data, err := json.Marshal(opt)
		assert.NoError(t, err)
		expectedJSON, _ := json.Marshal(map[string]int{"one": 1, "two": 2})
		assert.JSONEq(t, string(expectedJSON), string(data))

		var optUnmarshalled optionalv2.Option[map[string]int]
		err = json.Unmarshal(data, &optUnmarshalled)
		assert.NoError(t, err)
		assert.True(t, optUnmarshalled.IsSome())
		assert.Equal(t, map[string]int{"one": 1, "two": 2}, optUnmarshalled.Unwrap())
	})

	// Test behavior with interface types
	t.Run("InterfaceTypes", func(t *testing.T) {
		var i interface{} = "string value"
		opt := optionalv2.Some(i)
		assert.True(t, opt.IsSome())
		assert.Equal(t, "string value", opt.Unwrap())

		data, err := json.Marshal(opt)
		assert.NoError(t, err)
		assert.JSONEq(t, `"string value"`, string(data))

		var optUnmarshalled optionalv2.Option[interface{}]
		err = json.Unmarshal(data, &optUnmarshalled)
		assert.NoError(t, err)
		assert.True(t, optUnmarshalled.IsSome())
		assert.Equal(t, "string value", optUnmarshalled.Unwrap())
	})

	// Test methods with custom types implementing fmt.Stringer
	t.Run("CustomStringer", func(t *testing.T) {
		customValue := CustomType{ID: 2, Name: "Custom"}
		opt := optionalv2.Some(customValue)
		assert.Equal(t, "Some[CustomType(ID=2, Name=Custom)]", opt.String())

		optNone := optionalv2.None[CustomType]()
		assert.Equal(t, "None[]", optNone.String())
	})

	// Test TakeOrElse with side effects
	t.Run("TakeOrElseSideEffects", func(t *testing.T) {
		opt := optionalv2.None[int]()
		var sideEffect int

		value := opt.TakeOrElse(func() int {
			sideEffect = 1
			return 42
		})
		assert.Equal(t, 42, value)
		assert.Equal(t, 1, sideEffect)

		opt = optionalv2.Some(10)
		sideEffect = 0
		value = opt.TakeOrElse(func() int {
			sideEffect = 1
			return 100
		})
		assert.Equal(t, 10, value)
		assert.Equal(t, 0, sideEffect)
	})
}
