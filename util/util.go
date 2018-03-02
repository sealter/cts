package util

import (
	"path"
	"runtime"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
)

// FuncName get current scope of function name
func FuncName() string {
	p := make([]uintptr, 1)
	runtime.Callers(2, p)
	fullname := runtime.FuncForPC(p[0]).Name()

	_, name := path.Split(fullname)
	return name
}

// Decode convert an arbitrary map[string]interface{} into a Go structure.
func Decode(m map[string]interface{}, i interface{}) error {
	decoder, err := mapstructure.NewDecoder(
		&mapstructure.DecoderConfig{
			WeaklyTypedInput: true,
			Result:           i,
		})
	if err != nil {
		return errors.Wrap(err, FuncName())
	}

	err = decoder.Decode(m)
	if err != nil {
		return errors.Wrap(err, FuncName())
	}
	return nil
}
