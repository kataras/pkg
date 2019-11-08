package config

import (
	"fmt"
	"reflect"

	"github.com/AlecAivazis/survey/v2"
)

// TryAsk is called from `Load` if configuration file
// didn't contain the 'required' configuration struct's fields
// but it can be called manually as well, if that's needed.
//
// Remember, "dest" should be a pointer to a struct's instance.
//
// If not any field to be prompted for value then this function does nothing.
// Returns true if it was something to ask, otherwise false.
func TryAsk(dest interface{}) bool {
	if !ok(dest) {
		return false
	}

	visitMissingFields(dest, func(f field, fValue reflect.Value) {
		var unusedAns interface{}
		fieldTyp := fValue.Type()
		survey.AskOne(makePrompt(fieldTyp, f), &unusedAns, makeValidator(fieldTyp, fValue))
	})

	return true
}

func makePrompt(fieldTyp reflect.Type, f field) survey.Prompt {
	fieldName := f.Name

	// if it's a boolean then show a confirmation prompt.
	if fieldTyp.Kind() == reflect.Bool {
		return &survey.Confirm{
			Help:    fmt.Sprintf("The provided type of '%s' should be %s.", fieldName, fieldTyp.Name()),
			Message: fmt.Sprintf("%s?", fieldName),
		}
	}

	// if it's a secret then show a password (replaces text to ****) prompt.
	if f.Secret {
		return &survey.Password{
			Help:    fmt.Sprintf("The provided type of '%s' should be a secret of %s.", fieldName, fieldTyp.Name()),
			Message: fmt.Sprintf("Please type the value for the setting '%s'", fieldName),
		}
	}

	// otherwise show an input with a default value as well in parenthesis ().
	zero := reflect.Zero(fieldTyp)
	var def string
	if zero.IsValid() && zero.CanInterface() {
		def = fmt.Sprintf("%v", zero.Interface())
	}

	return &survey.Input{
		Default: def,
		Help:    fmt.Sprintf("The provided type of '%s' should be %s.", fieldName, fieldTyp.Name()),
		Message: fmt.Sprintf("Please type the value for the setting '%s'", fieldName),
	}
}

func makeValidator(fieldTyp reflect.Type, fieldVal reflect.Value) survey.AskOpt {
	validator := func(gotValue interface{}) error {
		// gotValue can be bool(if confirmation) or string otherwise.
		if _, ok := gotValue.(bool); ok {
			// if it was bool, then we're ready to set it as it's.
			fieldVal.Set(reflect.ValueOf(gotValue))
			return nil
		}

		// value to set.
		value := parseValue(gotValue, fieldTyp)
		if value == nil {
			gotTyp := reflect.TypeOf(value)
			return fmt.Errorf("invalid type of value passed, expected: %s but got %s", fieldTyp.Name(), gotTyp.Name())
		}

		fieldVal.Set(reflect.ValueOf(value))
		return nil
	}

	return func(options *survey.AskOptions) error {
		options.Validators = append(options.Validators, validator)
		return nil
	}
}
