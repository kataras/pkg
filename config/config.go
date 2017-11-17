package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/kataras/pkg/zerocheck"
	"github.com/kataras/survey"
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
	fileDecoder   FileDecoder
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

// Load loads a specific YAML configuration file
// and sets that to "dest".
// If the configuration file didn't contain any sensetive fields
// and the fields are tagged as 'required' then
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
		return tryToAsk(dest, nil, opts)
	}

	// get the abs
	// which will try to find the 'fullpath' from current workind dir too.
	yamlAbsPath, err := filepath.Abs(fullpath)
	if err != nil {
		return tryToAsk(dest, err, opts)
	}

	// read the raw contents of the file.
	data, err := ioutil.ReadFile(yamlAbsPath)
	if err != nil {
		return tryToAsk(dest, err, opts)
	}

	// convert the file's contents to the configuration and keep the error.
	err = opts.fileDecoder(data, dest)

	return tryToAsk(dest, err, opts)
}

func tryToAsk(dest interface{}, prev error, opts options) error {
	// try to ask;
	// if not enabled then it will return the file decoder's error.
	// if enabled but nothing to ask then it will return the file decoder's error.
	// if enabled and asked, so settings are set-ed, then skip the file decoder's error and return nil.
	if !opts.disableSurvey {
		if asked := TryAsk(dest); !asked {
			return prev
		}
	}

	return prev
}

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

	v := reflect.ValueOf(dest).Elem()
	typElem := reflect.TypeOf(dest).Elem() // the struct's type.
	fields := lookupFields(typElem, -1)
	for _, f := range fields {
		fieldVal := v.FieldByIndex(f.Index)

		var unusedAns interface{}

		if f.Required && zerocheck.IsZero(fieldVal) {
			fieldTyp := fieldVal.Type()
			// this will error
			// because it can't convert it correctly, that's why
			// we have the set-logic into our custom makeValidator.
			survey.AskOne(makePrompt(fieldTyp, f), &unusedAns, makeValidator(fieldTyp, fieldVal))
		}
	}

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

// SurveyTimeLayout is the layout that is used
// when a field is a time.Time type and it's required.
// the user type a time in string, and in order to this
// string to be converted to a time.Time and be set-ed
// to the setting field it needs to have a known layout.
//
// Defaults to "Mon, 02 Jan 2006 15:04:05 GMT".
var SurveyTimeLayout = "Mon, 02 Jan 2006 15:04:05 GMT"

func makeValidator(fieldTyp reflect.Type, fieldVal reflect.Value) survey.Validator {
	return func(answer interface{}) error {
		// answer can be bool(if confirmation) or string otherwise.
		if _, ok := answer.(bool); ok {
			// if it was bool, then we're ready to set it as it's.
			fieldVal.Set(reflect.ValueOf(answer))
			return nil
		}

		var got interface{}

		ans, ok := answer.(string)
		// this will never errored with the current survey's implementation but we make
		// this check for good and worst.
		if !ok {
			gotTyp := reflect.TypeOf(answer)
			return fmt.Errorf("invalid type of value passed, expected: %s but got %s", fieldTyp.Name(), gotTyp.Name())
		}

		switch fieldTyp.Kind() {
		case reflect.Int:
			// we could use dynamic "bits" variable but we don't.
			got, _ = strconv.Atoi(ans)
		case reflect.Int32:
			got, _ = strconv.ParseInt(ans, 10, 32)
		case reflect.Int64:
			got, _ = strconv.ParseInt(ans, 10, 64)
		case reflect.Float32:
			got, _ = strconv.ParseFloat(ans, 32)
		case reflect.Float64:
			got, _ = strconv.ParseInt(ans, 10, 64)
		case reflect.Bool:
			// this is already checked before but keep for future, if I decide to remove confirmation dialogs.
			got, _ = strconv.ParseBool(ans)
		case reflect.Slice:
			got = strings.Split(ans, ",")
		case reflect.Struct:
			if fieldTyp.AssignableTo(reflect.TypeOf(time.Time{})) {
				// if setting is a struct and it's time.
				got, _ = time.Parse(SurveyTimeLayout, ans)
			}
		default:
			got = ans
		}

		fieldVal.Set(reflect.ValueOf(got))
		return nil
	}
}

func isZero(v reflect.Value) bool {
	return zerocheck.IsZero(v)
}
