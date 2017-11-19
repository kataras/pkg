package config

import (
	"errors"
	"flag"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/kataras/pkg/zerocheck"
	"gopkg.in/yaml.v2"
)

// Here are some optional settings that can be passed to the `Load` func.
//
// Guidelines:
// i) if boolean && defaults to false then its prerefixed with "disable".
// ii) if boolean & defaults to true then its prefixed with "enable".
// Zero boolean values should be the defaults as well.
//
// iii) any other setting has a valid default value which is initialized at `Load`.
//
// All options are prefixed with the words: "With" or "Without".
type options struct {
	disableSurvey bool
	// if not nil then it will scan for flags after the file loading, file is always first but it can be disabled.
	flagSet     *flag.FlagSet
	fileDecoder FileDecoder
	/// TODO:
	// if users ask then it will be a good idea
	// to add a loader via os environment variables as well.
}

// Option should be implement by all options, it's used to set the `options`.
type Option func(o *options)

// WithoutSurvey disables the prompts about missing configuration fields.
// Defaults to false; survey is enabled.
func WithoutSurvey(o *options) {
	o.disableSurvey = true
}

// CommandLine flagset is a shortcut for the `flag.CommandLine`
// so end-users wont have to import the flag package to use the `WithFlags`.
var CommandLine = flag.CommandLine

// WithFlags enables the config to be loaded from specific flag set( i.e flag.CommandLine).
// The flags may or may not be parsed before.
//
// Note that the end-user should declare the needed flags otherwise this will PANIC (at least for now).
// It scans the flags after the file decoder and before the survey; loading from file is always first action but it can be disabled if needed.
func WithFlags(set *flag.FlagSet) Option {
	return func(o *options) {
		o.flagSet = set
	}
}

// FileDecoder is the supported kind of function that
// are allowed to be passed as custom function to decode file's contents
// and unmarshal to a specific configuration struct
// before asking for missing settings via `os.Stdin` when survey is not disabled.
type FileDecoder func(fileContents []byte, dest interface{}) (err error)

// WithFileDecoder changes the default decoder/unmarshaler which is used
// to set the "dest" configuration using a specific file's contents.
// To disable file decoding entirely pass nil.
//
// Defaults to a 'YAML' unmarshaler.
func WithFileDecoder(fileDecoder FileDecoder) Option {
	return func(o *options) {
		o.fileDecoder = fileDecoder
	}
}

// ErrBad fired when bad value of "dest" passed.
var ErrBad = errors.New("dest should be a non-nil pointer to a struct")

func ok(dest interface{}) bool {
	if dest == nil {
		return false
	}
	// no indevitual fields are allowed only pointers to struct.
	typ := reflect.TypeOf(dest)
	typElem := typ.Elem() // the struct's type.
	if typ.Kind() != reflect.Ptr || typElem.Kind() != reflect.Struct {
		return false
	}

	return true
}

// Load fills the "dest", which should be a non-nil pointer to a struct value,
// based a specific configuration file,
// by default YAML but this can be changed with an `Option`.
// If the configuration file didn't contain any sensetive fields
// and the fields are not tagged as 'config:"-"' then
// it prompts the user to define these fields' values from
// the `os.Stdin` using the survey package.
//
// The "dest" should be a pointer to a struct value
// and may be filled before this call.
//
// Returns an error if something bad happened like
// bad yaml-formated file.
func Load(fullpath string, dest interface{}, optional ...Option) error {
	if !ok(dest) {
		return ErrBad
	}

	// default options.
	opts := options{
		fileDecoder:   yaml.Unmarshal,
		disableSurvey: false,
		flagSet:       nil,
	}

	for _, opt := range optional {
		opt(&opts)
	}

	// when error is nil:
	// - if file decoder is disabled
	// - if file decoder did its job without errors
	// when error is not nil:
	// - if file decoder is not disabled AND
	//  * an error occurred when loading the file AND
	//	 * survey is disabled or failed to ask anything.
	//
	// If survey is enabled it will try to ask at any state of error, so
	// we can reduce errors for file not find if survey successfully asked all the required
	// settings.

	if opts.fileDecoder == nil {
		// file decoder was disabled.
		// nothing to set, file load disabled and survey has nothing to ask, so pass
		// nil as the prev error.
		return next(dest, nil, opts)
	}

	// get the abs
	// which will try to find the 'fullpath' from current workind dir too.
	f, err := filepath.Abs(fullpath)
	if err != nil {
		return next(dest, err, opts)
	}

	// read the raw contents of the file.
	data, err := ioutil.ReadFile(f)
	if err != nil {
		return next(dest, err, opts)
	}

	// convert the file's contents to the configuration and keep the error.
	err = opts.fileDecoder(data, dest)

	return next(dest, err, opts)
}

func next(dest interface{}, prev error, opts options) error {
	if opts.flagSet != nil {
		if err := TryLoadFlags(opts.flagSet, dest); err != nil {
			return err
		}
	}

	// try to ask;
	// if not enabled then it will return the file decoder's error.
	// if enabled but nothing to ask then it will return the file decoder's error.
	// if enabled and asked, so settings are set-ed, then skip the file decoder's error and return nil.
	if !opts.disableSurvey {
		TryAsk(dest)
	}

	return prev
}

func visitMissingFields(dest interface{}, fn func(f field, fValue reflect.Value)) {
	v := reflect.ValueOf(dest).Elem()
	typElem := reflect.TypeOf(dest).Elem() // the struct's type.
	fields := lookupFields(typElem, field{})
	for _, f := range fields {
		fieldVal := v.FieldByIndex(f.Index)
		if f.Required && zerocheck.IsZero(fieldVal) {
			fn(f, fieldVal)
		}
	}
}

// check if gotValue is the same as fieldKind, if yes then set it's
// second, check if it's string, then take that string and try to parse
// in the fieldKind's value.
func parseValue(gotValue interface{}, fieldTyp reflect.Type) (value interface{}) {
	switch v := gotValue.(type) {
	case bool:
		return parseBool(v, fieldTyp)
	case string:
		return parseString(v, fieldTyp)
	case int:
		return parseInt(v, fieldTyp)
	default:
		return nil
	}
}

// TimeLayout is the layout that is used
// when a field is a time.Time type and it's required.
// the user type a time in string, and in order to this
// string to be converted to a time.Time and be set-ed
// to the setting field it needs to have a known layout.
//
// Defaults to "Mon, 02 Jan 2006 15:04:05 GMT".
var TimeLayout = "Mon, 02 Jan 2006 15:04:05 GMT"

// parses a string based on the wanted field's type of kind and
// returns the result value that will be set-ed to the field by the caller.
func parseString(got string, fieldTyp reflect.Type) (value interface{}) {
	want := fieldTyp.Kind()
	switch want {
	case reflect.String:
		return got // we had a string and we want a string, just return.
	case reflect.Int:
		// we could use dynamic "bits" variable but we don't.
		value, _ = strconv.Atoi(got)
	case reflect.Int32:
		value, _ = strconv.ParseInt(got, 10, 32)
	case reflect.Int64:
		value, _ = strconv.ParseInt(got, 10, 64)
	case reflect.Float32:
		value, _ = strconv.ParseFloat(got, 32)
	case reflect.Float64:
		value, _ = strconv.ParseInt(got, 10, 64)
	case reflect.Bool:
		// this is already checked before but keep for future, if I decide to remove confirmation dialogs.
		value, _ = strconv.ParseBool(got)
	case reflect.Slice:
		value = strings.Split(got, ",")
	case reflect.Struct:
		if fieldTyp.AssignableTo(reflect.TypeOf(time.Time{})) {
			// if setting is a struct and it's time.
			value, _ = time.Parse(TimeLayout, got)
		}
	}

	return
}

// parseInt parses an int based on the wanted field's type of kind and
// returns the result value that will be set-ed to the field by the caller.
func parseInt(got int, fieldTyp reflect.Type) (value interface{}) {
	want := fieldTyp.Kind()
	switch want {
	case reflect.String:
		return strconv.Itoa(got)
	case reflect.Int:
		// we had an int and we want an int, just return.
		return got
	case reflect.Int32:
		// b := byte(got % 256)
		return int32(got)
	case reflect.Int64:
		return int64(got)
	case reflect.Float32:
		return float32(got)
	case reflect.Float64:
		return float64(got)
	case reflect.Bool:
		if got <= 0 {
			return false
		}
		return true
	}
	return nil
}

func parseBool(got bool, fieldTyp reflect.Type) interface{} {
	want := fieldTyp.Kind()
	switch want {
	case reflect.String:
		return strconv.FormatBool(got)
	case reflect.Int:
		if !got {
			return 0
		}
		return 1
	case reflect.Int32:
		if !got {
			return int32(0)
		}
		return int32(1)
	case reflect.Int64:
		if !got {
			return int64(0)
		}
		return int64(1)
	case reflect.Float32:
		if !got {
			return float32(0)
		}
		return float32(1)
	case reflect.Float64:
		if !got {
			return float64(0)
		}
		return float64(1)
	case reflect.Bool:
		return got
	}

	return nil
}

func isZero(v reflect.Value) bool {
	return zerocheck.IsZero(v)
}
