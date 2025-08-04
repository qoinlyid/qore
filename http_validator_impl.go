package qore

import (
	"errors"
	"fmt"
	"net"
	"reflect"
	"strings"
	"sync"

	"github.com/go-playground/validator/v10"
)

type httpValidatorImpl struct {
	validate *validator.Validate
}

// Compile time check `httpValidatorImpl implements `HttpValidator`.
var _ HttpValidator = (*httpValidatorImpl)(nil)

var httpValidatorErrPool = sync.Pool{
	New: func() any {
		return new(HttpValidatorErr)
	},
}

// Fungsi ini akan mereset dan mengembalikan error ke pool
func putHttpValidatorErr(verr *HttpValidatorErr) {
	if verr == nil {
		return
	}
	verr.ErrMandatory = nil
	verr.ErrFormat = nil
	verr.recycle = nil
	httpValidatorErrPool.Put(verr)
}

// httpValidateIpCidr returns true if given field is valid IP or CIDR, otherwise return false.
func (v *httpValidatorImpl) httpValidateIpCidr(field validator.FieldLevel) bool {
	checker := func(src string) bool {
		if net.ParseIP(src) != nil {
			return true
		}
		_, _, err := net.ParseCIDR(src)
		return err == nil
	}

	switch field.Field().Kind() {
	case reflect.String:
		return checker(field.Field().String())
	case reflect.Slice:
		for i := 0; i < field.Field().Len(); i++ {
			val := field.Field().Index(i)
			if !checker(val.String()) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// httpValidatorDefault creates new default HTTP(S) validator implementation.
func httpValidatorDefault() *httpValidatorImpl {
	validator := &httpValidatorImpl{
		validate: validator.New(),
	}
	validator.validate.RegisterValidation("ip_or_cidr", validator.httpValidateIpCidr)
	validator.validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		tag := field.Tag.Get("json")
		if tag == "-" {
			return ""
		}
		return tag
	})
	return validator
}

func (v *httpValidatorImpl) Validate(source any) *HttpValidatorErr {
	err := v.validate.Struct(source)
	fieldErrs, ok := err.(validator.ValidationErrors)
	if ok {
		// Get http validator error pointer from the pool.
		verr := httpValidatorErrPool.Get().(*HttpValidatorErr)
		verr.recycle = putHttpValidatorErr

		// If field errors is any.
		if len(fieldErrs) > 0 {
			fe := fieldErrs[0]
			field := StringRemoveNonAlphabet(fe.Field())

			// Parse error from struct validator.
			switch {
			default:
				verr.ErrFormat = fmt.Errorf("field [%s] is invalid", field)
			case strings.Contains(fe.Tag(), "required"):
				verr.ErrMandatory = fmt.Errorf("field [%s] cannot be empty", field)
			case fe.Tag() == "min" || fe.Tag() == "max":
				verr.ErrFormat = fmt.Errorf("field [%s] is invalid, %s length %s", field, fe.Tag(), fe.Param())
			}
			return verr
		}

		verr.ErrMandatory = errors.New("validation error: not a field error")
		return verr

	}

	// Valid.
	return nil
}
