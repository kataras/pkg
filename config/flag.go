package config

import (
	"flag"
	"os"
	"reflect"
	"strings"
)

// TryLoadFlags tries to load the "dest" configuration from a flag set.
// For command line applications the flagset you should provide is the `flag.CommandLine`.
//
// If flags are not parsed then it parses by omitting the first argument, which is the
// executable name for command line applications (os.Args[1:]).
// The flags may or may not be parsed already.
//
// Note that the end-user should declare the needed flags otherwise this will panic.
func TryLoadFlags(set *flag.FlagSet, dest interface{}) error {
	if !ok(dest) {
		return ErrBad
	}

	if !set.Parsed() {
		if err := set.Parse(os.Args[1:]); err != nil {
			return err
		}
	}

	visitMissingFields(dest, func(f field, fValue reflect.Value) {
		name := strings.ToLower(f.Name) // even if customized Name is capitalized, flag's name should be all lowercase.
		arg := set.Lookup(name)
		if arg != nil {
			value := parseString(arg.Value.String(), fValue.Type())
			fValue.Set(reflect.ValueOf(value))
		}
	})

	return nil
}
