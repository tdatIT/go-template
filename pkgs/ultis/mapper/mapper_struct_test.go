package mapper_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/tdatIT/go-template/pkgs/ultis/mapper"
)

// Test for Copy function
func TestCopy(t *testing.T) {
	type Source struct {
		Field1 string
		Field2 int
	}

	type Destination struct {
		Field1 string
		Field2 int
	}

	src := Source{Field1: "test", Field2: 100}
	var dest Destination

	err := mapper.Copy(&dest, &src)
	assert.NoError(t, err)
	assert.Equal(t, src.Field1, dest.Field1)
	assert.Equal(t, src.Field2, dest.Field2)
}

// Test for CopyIgnoreEmpty function
func TestCopyIgnoreEmpty(t *testing.T) {
	type Source struct {
		Field1 string
		Field2 int
	}

	type Destination struct {
		Field1 string
		Field2 int
	}

	src := Source{Field1: "", Field2: 100}
	var dest Destination

	err := mapper.CopyIgnoreEmpty(&dest, &src)
	assert.NoError(t, err)
	assert.Equal(t, 100, dest.Field2)
	assert.Empty(t, dest.Field1)

	// With non-zero value for Field1
	src.Field1 = "Non-empty"
	err = mapper.CopyIgnoreEmpty(&dest, &src)
	assert.NoError(t, err)
	assert.Equal(t, "Non-empty", dest.Field1)
}

// Test for BindingStruct function
func TestBindingStruct(t *testing.T) {
	type Source struct {
		Field1 string `json:"field1"`
		Field2 int    `json:"field2"`
	}

	type Destination struct {
		Field1 string `json:"field1"`
		Field2 int    `json:"field2"`
	}

	src := Source{Field1: "test", Field2: 100}
	var dest Destination

	err := mapper.BindingStruct(src, &dest)
	assert.NoError(t, err)
	assert.Equal(t, "test", dest.Field1)
	assert.Equal(t, 100, dest.Field2)
}

// Test for BindingAndValidate function
func TestBindingAndValidate(t *testing.T) {
	type Model struct {
		Field1 string `json:"field1"`
		Field2 int    `json:"field2"`
	}

	validator := func(v interface{}) error {
		m := v.(Model)
		if m.Field1 == "" {
			return errors.New("Field1 cannot be empty")
		}
		return nil
	}

	detail := map[string]interface{}{"field1": "valid", "field2": 100}
	model, err := mapper.BindingAndValidate[Model](detail, validator)
	assert.NoError(t, err)
	assert.Equal(t, "valid", model.Field1)

	// Test invalid validation
	detail = map[string]interface{}{"field1": "", "field2": 100}
	model, err = mapper.BindingAndValidate[Model](detail, validator)
	assert.Error(t, err)
	assert.EqualError(t, err, "Field1 cannot be empty")
}

// Test for StructToMap function
func TestStructToMap(t *testing.T) {
	type TestStruct struct {
		Field1 string  `json:"field1"`
		Field2 int     `json:"field2"`
		Field3 *string `json:"field3"`
	}

	field3 := "optional"
	input := TestStruct{Field1: "value1", Field2: 42, Field3: &field3}

	result := mapper.StructToMap(input, true)
	assert.Equal(t, "value1", result["field1"])
	assert.Equal(t, 42, result["field2"])
	assert.Equal(t, "optional", result["field3"])

	// ignoreNilFiled=true: nil pointer field must be omitted
	input.Field3 = nil
	result = mapper.StructToMap(input, true)
	_, exists := result["field3"]
	assert.False(t, exists)

	// ignoreNilFiled=false: nil pointer field must appear as nil (no panic)
	result = mapper.StructToMap(input, false)
	val, exists := result["field3"]
	assert.True(t, exists)
	assert.Nil(t, val)

	// pointer input must not panic
	assert.NotPanics(t, func() {
		mapper.StructToMap(&input, false)
	})
}

// Test for GetJsonStringify function
func TestGetJsonStringify(t *testing.T) {
	type TestStruct struct {
		Field1 string `json:"field1"`
		Field2 int    `json:"field2"`
	}

	input := TestStruct{Field1: "value1", Field2: 42}
	jsonStr := mapper.GetJsonStringify(input)
	expected := `{"field1":"value1","field2":42}`
	assert.JSONEq(t, expected, jsonStr)

	// Test with marshaling error
	var invalidInput = make(chan int)
	jsonStr = mapper.GetJsonStringify(invalidInput)
	assert.Empty(t, jsonStr)
}
