package vt

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/vmkteam/zenrpc/v2"
)

const (
	FieldErrorRequired  = "required"
	FieldErrorMax       = "max"
	FieldErrorMin       = "min"
	FieldErrorIncorrect = "incorrect"
	FieldErrorUnique    = "unique"
	FieldErrorFormat    = "format"
	FieldErrorLen       = "len"
)

const (
	CustomStatusTag = "status"
	CustomAliasTag  = "alias"

	fieldPathSeparator = "."
)

var errorMap = map[string]string{
	"max":           FieldErrorMax,
	"min":           FieldErrorMin,
	"required":      FieldErrorRequired,
	"gt":            FieldErrorRequired,
	"len":           FieldErrorLen,
	CustomStatusTag: FieldErrorIncorrect,
	CustomAliasTag:  FieldErrorFormat,
}

var validate = newPlaygroundValidator()

func newPlaygroundValidator() *validator.Validate {
	validate := validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	_ = validate.RegisterValidationCtx(CustomStatusTag, validateStatus)
	_ = validate.RegisterValidationCtx(CustomAliasTag, validateAlias)
	return validate
}

func validateStatus(_ context.Context, fl validator.FieldLevel) bool {
	id := int(fl.Field().Int())
	return NewStatus(id) != nil
}

var aliasRegex = regexp.MustCompile(`^([0-9a-z-])+$`)

func validateAlias(_ context.Context, fl validator.FieldLevel) bool {
	return aliasRegex.Match([]byte(fl.Field().String()))
}

type FieldError struct {
	Field      string                `json:"field"`
	Error      string                `json:"error"`
	Constraint *FieldErrorConstraint `json:"constraint,omitempty"` // Help with generating an error message.
}

type FieldErrorConstraint struct {
	Max int `json:"max,omitempty"` // Max value for field.
	Min int `json:"min,omitempty"` // Min value for field.
}

type FieldErrorConstraintFunc func(*FieldErrorConstraint)

// NewFieldErrorConstraint return FieldErrorConstraint by validation tag name.
func NewFieldErrorConstraint(errorName string, param string) *FieldErrorConstraint {
	switch errorName {
	case FieldErrorMin:
		value, err := strconv.Atoi(param)
		if err != nil {
			return nil
		}
		return &FieldErrorConstraint{Min: value}
	case FieldErrorMax:
		value, err := strconv.Atoi(param)
		if err != nil {
			return nil
		}
		return &FieldErrorConstraint{Max: value}
	default:
		return nil
	}
}

func NewFieldError(e validator.FieldError) FieldError {
	tag := e.Tag()
	splitted := strings.Split(e.Namespace(), fieldPathSeparator)[1:]
	field := strings.Join(splitted, fieldPathSeparator)
	fe := FieldError{Field: field, Error: tag}
	if mappedError, ok := errorMap[tag]; ok {
		fe.Error = mappedError
	}
	fe.Constraint = NewFieldErrorConstraint(fe.Error, e.Param())
	return fe
}

type Validator struct {
	fields []FieldError
	err    error
}

func (v *Validator) Fields() []FieldError {
	if len(v.fields) == 0 {
		return []FieldError{}
	}
	return v.fields

}

func (v *Validator) Append(field string, err string, constraintFuncs ...FieldErrorConstraintFunc) {
	f := FieldError{Field: field, Error: err}
	if len(constraintFuncs) > 0 {
		c := &FieldErrorConstraint{}
		for _, fn := range constraintFuncs {
			fn(c)
		}
		f.Constraint = c
	}
	v.fields = append(v.fields, f)
}

func (v *Validator) SetInternalError(err error) {
	v.err = err
}

func (v *Validator) HasInternalError() bool {
	return v.err != nil
}

func (v *Validator) HasErrors() bool {
	return len(v.fields) != 0 || v.HasInternalError()
}

func (v *Validator) Error() error {
	if v.err != nil {
		return InternalError(v.err)
	} else if len(v.fields) != 0 {
		return ValidationError(v.fields)
	}
	return nil
}

func (v *Validator) CheckBasic(ctx context.Context, item interface{}) {
	v.SetInternalError(nil)
	err := validate.StructCtx(ctx, item)
	if err == nil {
		return
	}

	var playgroundValidationErrors validator.ValidationErrors

	if errors.As(err, &playgroundValidationErrors) {
		for _, fieldError := range playgroundValidationErrors {
			v.fields = append(v.fields, NewFieldError(fieldError))
		}
	} else {
		v.SetInternalError(err)
	}
}

func InternalError(err error) *zenrpc.Error {
	return zenrpc.NewError(http.StatusInternalServerError, err)
}

func ValidationError(fieldErrors []FieldError) *zenrpc.Error {
	return &zenrpc.Error{Code: http.StatusBadRequest, Data: fieldErrors, Message: "Validation err"}
}
