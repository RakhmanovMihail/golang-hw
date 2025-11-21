package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var sb strings.Builder
	for i, err := range v {
		sb.WriteString(err.Field + ": " + err.Err.Error())
		if i < len(v)-1 {
			sb.WriteString("; ")
		}
	}
	return sb.String()
}

func Validate(v interface{}) error {
	val := reflect.ValueOf(v)

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Если не структура, то сразу выкидываем ошибку
	if val.Kind() != reflect.Struct {
		return errors.New(ValidationErrors{ValidationError{"Struct", errors.New("expected a struct")}}.Error())
	}

	typ := val.Type()
	var validationErrors ValidationErrors

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldType := typ.Field(i)

		if !field.CanInterface() {
			continue
		}

		tag := fieldType.Tag.Get("validate")
		if tag == "" {
			continue
		}

		rules := parseValidateTag(tag)
		for _, r := range rules {
			if verr := applyRule(field, fieldType.Name, r); verr != nil {
				validationErrors = append(validationErrors, *verr)
			}
		}
	}

	if len(validationErrors) > 0 {
		return errors.New(validationErrors.Error())
	}
	return nil
}

type rule struct {
	name  string
	value string
}

func parseValidateTag(tag string) []rule {
	parts := strings.Split(tag, "|")
	res := make([]rule, 0, len(parts))
	for _, p := range parts {
		kv := strings.SplitN(p, ":", 2)
		if len(kv) != 2 {
			continue
		}
		res = append(res, rule{name: kv[0], value: kv[1]})
	}
	return res
}

func applyRule(field reflect.Value, fieldName string, r rule) *ValidationError {
	var err error
	switch r.name {
	case "len":
		err = validateLen(field, r.value)
	case "min":
		err = validateMin(field, r.value)
	case "max":
		err = validateMax(field, r.value)
	case "in":
		err = validateIn(field, r.value)
	case "regexp":
		err = validateRegexp(field, r.value)
	default:
		return nil
	}
	if err != nil {
		return &ValidationError{Field: fieldName, Err: err}
	}
	return nil
}

type number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

func validateNumericField[T number](field T, ruleValue string, compareFunc func(T, T) bool, errorMsg string) error {
	var val T
	switch any(field).(type) {
	case float32, float64:
		f, err := strconv.ParseFloat(ruleValue, 64)
		if err != nil {
			return fmt.Errorf("invalid value: %v", ruleValue)
		}
		val = T(f)
	case uint, uint8, uint16, uint32, uint64, uintptr:
		u, err := strconv.ParseUint(ruleValue, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid value: %v", ruleValue)
		}
		val = T(u)
	default:
		i, err := strconv.ParseInt(ruleValue, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid value: %v", ruleValue)
		}
		val = T(i)
	}

	if !compareFunc(field, val) {
		return fmt.Errorf(errorMsg, val)
	}
	return nil
}

func validateLen(field reflect.Value, ruleValue string) error {
	expectedLen, err := strconv.Atoi(ruleValue)
	if err != nil {
		return fmt.Errorf("invalid length value: %v", ruleValue)
	}

	validateString := func(s string) error {
		if len(s) != expectedLen {
			return fmt.Errorf("field length must be %d", expectedLen)
		}
		return nil
	}

	//nolint:exhaustive
	switch field.Kind() {
	case reflect.String:
		return validateString(field.String())
	case reflect.Slice, reflect.Array:
		for i := 0; i < field.Len(); i++ {
			elem := field.Index(i)
			if elem.Kind() == reflect.String {
				if err := validateString(elem.String()); err != nil {
					return fmt.Errorf("element %d: %w", i, err)
				}
			}
		}
		return nil
	default:
		return fmt.Errorf("field type does not support length validation")
	}
}

func validateMin(field reflect.Value, ruleValue string) error {
	//nolint:exhaustive
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return validateNumericField(
			field.Int(),
			ruleValue,
			func(a, b int64) bool { return a >= b },
			"value must be at least %v",
		)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return validateNumericField(
			field.Uint(),
			ruleValue,
			func(a, b uint64) bool { return a >= b },
			"value must be at least %v",
		)
	case reflect.Float32, reflect.Float64:
		return validateNumericField(
			field.Float(),
			ruleValue,
			func(a, b float64) bool { return a >= b },
			"value must be at least %v",
		)
	default:
		return fmt.Errorf("field does not support min validation")
	}
}

func validateMax(field reflect.Value, ruleValue string) error {
	//nolint:exhaustive
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return validateNumericField(
			field.Int(),
			ruleValue,
			func(a, b int64) bool { return a <= b },
			"value must be at most %v",
		)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return validateNumericField(
			field.Uint(),
			ruleValue,
			func(a, b uint64) bool { return a <= b },
			"value must be at most %v",
		)
	case reflect.Float32, reflect.Float64:
		return validateNumericField(
			field.Float(),
			ruleValue,
			func(a, b float64) bool { return a <= b },
			"value must be at most %v",
		)
	default:
		return fmt.Errorf("field does not support max validation")
	}
}

func validateIn(field reflect.Value, ruleValue string) error {
	allowedValues := strings.Split(ruleValue, ",")
	if len(allowedValues) == 0 {
		return fmt.Errorf("no allowed values specified")
	}

	checkInSlice := func(value string) error {
		for _, v := range allowedValues {
			if value == v {
				return nil
			}
		}
		return fmt.Errorf("value must be one of: %s", strings.Join(allowedValues, ", "))
	}

	//nolint:exhaustive
	switch field.Kind() {
	case reflect.String:
		return checkInSlice(field.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return checkInSlice(strconv.FormatInt(field.Int(), 10))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return checkInSlice(strconv.FormatUint(field.Uint(), 10))
	default:
		return fmt.Errorf("field does not support 'in' validation")
	}
}

func validateRegexp(field reflect.Value, ruleValue string) error {
	if field.Kind() != reflect.String {
		return fmt.Errorf("field must be a string for regular expression validation")
	}

	re, err := regexp.Compile(ruleValue)
	if err != nil {
		return fmt.Errorf("invalid regular expression: %v", ruleValue)
	}
	if !re.MatchString(field.String()) {
		return fmt.Errorf("value does not match the pattern")
	}

	return nil
}
