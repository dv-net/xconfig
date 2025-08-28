package validate

import (
	"reflect"

	"github.com/dv-net/xconfig/plugins"
)

type CustomValidator func(any) error

type validate interface {
	Validate() error
}

type validator struct {
	config          any
	customValidator []CustomValidator
}

// New returns an validator plugin.
// It accepts a list of CustomValidator functions.
//
// By default, it will validate the struct with the Validate() method.
// If the struct does not have a Validate() method, it will be skipped.
//
// If you want to add custom validation, you can pass a list of CustomValidator functions.
// The CustomValidator function should accept an interface{} and return an error.
//
// Example:
//
//	type MyStruct struct {
//		Str string
//	}
//
//	func (m MyStruct) Validate() error {
//		if m.Str == "" {
//			return fmt.Errorf("Str is empty")
//		}
//		return nil
//	}
func New(validators ...CustomValidator) plugins.Plugin {
	v := &validator{}
	for _, validator := range validators {
		if validator == nil {
			continue
		}
		v.customValidator = append(v.customValidator, validator)
	}
	return v
}

func (v *validator) Parse() error {
	if v == nil {
		return nil
	}

	if err := validateElem(v.config); err != nil {
		return err
	}

	val := reflect.ValueOf(v.config).Elem()
	for i := range val.NumField() {
		if err := validateElem(val.Field(i).Addr().Interface()); err != nil {
			return err
		}
	}

	for _, validator := range v.customValidator {
		if err := validator(v.config); err != nil {
			return err
		}
	}

	return nil
}

func (v *validator) Walk(config any) error {
	v.config = config
	return nil
}

func validateElem(elem any) error {
	// try to validate with Validate() error
	if tmp, ok := elem.(validate); ok {
		if err := tmp.Validate(); err != nil {
			return err
		}
	}

	return nil
}
