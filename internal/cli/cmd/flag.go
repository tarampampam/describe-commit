package cmd

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type (
	Flagger interface {
		// IsSet returns true if the flag was set during parsing (current value != default value).
		IsSet() bool

		// Help returns the flag names and usage.
		Help() (names string, usage string)

		// Apply applies the flag to the flag set.
		Apply(*flag.FlagSet) error

		// Validate validates the flag value.
		Validate(*Command) error

		// RunAction runs the flag action and returns an error if any.
		RunAction(*Command) error
	}

	FlagType interface {
		bool | int | int64 | string | uint | uint64 | float64 | time.Duration
	}

	Flag[T FlagType] struct {
		Names        []string                // e.g. "config-file", "c"
		Usage        string                  // e.g. "path to the configuration file"
		Default      T                       // default value
		EnvVars      []string                // e.g. "CONFIG_FILE"
		Validator    func(*Command, T) error // value validation function
		Action       func(*Command, T) error // an action to run when the flag is set
		ValueSetFrom flagValueSource         // the source of the value (from where the value was taken)

		// the actual value will be placed here after parsing.
		// a pointer is used to allow the value to be set externally
		Value *T
	}
)

var ( // ensure that Flag[T] implements Flagger
	_ Flagger = (*Flag[bool])(nil)
	_ Flagger = (*Flag[int])(nil)
	_ Flagger = (*Flag[int64])(nil)
	_ Flagger = (*Flag[string])(nil)
	_ Flagger = (*Flag[uint])(nil)
	_ Flagger = (*Flag[uint64])(nil)
	_ Flagger = (*Flag[float64])(nil)
	_ Flagger = (*Flag[time.Duration])(nil)
)

type flagValueSource = byte

const (
	FlagValueSourceNone    flagValueSource = iota // not set
	FlagValueSourceDefault                        // set from the default value
	FlagValueSourceEnv                            // set from the environment variable
	FlagValueSourceFlag                           // set from the flag value (command line argument)
)

func (f *Flag[T]) IsSet() bool {
	if f.Value == nil {
		return false // uninitialized flag
	}

	switch f.ValueSetFrom {
	case FlagValueSourceNone, FlagValueSourceDefault:
		return false
	default:
		return *f.Value != f.Default
	}
}

func (f *Flag[T]) Help() (names string, usage string) {
	var b strings.Builder

	b.Grow(len(f.Usage) + 64) //nolint:mnd

	for i, name := range f.Names {
		if i > 0 {
			b.WriteString(", ")
		}

		if len(name) == 1 {
			b.WriteRune('-')
		} else {
			b.WriteString("--")
		}

		b.WriteString(name)

		if _, ok := any(*new(T)).(bool); !ok {
			b.WriteString(`="…"`)
		}
	}

	names = b.String()

	b.Reset()

	b.WriteString(f.Usage)
	b.WriteString(" (default: ")
	b.WriteString(fmt.Sprintf("%v", f.Default))
	b.WriteRune(')')

	if len(f.EnvVars) > 0 {
		b.WriteString(" [")

		for i, envVar := range f.EnvVars {
			if i > 0 {
				b.WriteString(", ")
			}

			b.WriteRune('$')
			b.WriteString(envVar)
		}

		b.WriteRune(']')
	}

	usage = b.String()

	return
}

// parseString parses the string value and returns the value of the generic type.
func (f *Flag[T]) parseString(s string) (T, error) {
	var empty T // readonly value

	switch any(empty).(type) {
	case bool:
		v, err := strconv.ParseBool(s)
		if err != nil {
			return empty, err
		}

		return any(v).(T), nil
	case int:
		v, err := strconv.Atoi(s)
		if err != nil {
			return empty, err
		}

		return any(v).(T), nil
	case int64:
		v, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return empty, err
		}

		return any(v).(T), nil
	case string:
		return any(s).(T), nil
	case uint:
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return empty, err
		}

		return any(uint(v)).(T), nil
	case uint64:
		v, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return empty, err
		}

		return any(v).(T), nil
	case float64:
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return empty, err
		}

		return any(v).(T), nil
	case time.Duration:
		v, err := time.ParseDuration(s)
		if err != nil {
			return empty, err
		}

		return any(v).(T), nil
	}

	return empty, fmt.Errorf("unsupported flag type: %T", empty)
}

// envValue returns the value from the environment variable if it was set.
// The function returns:
//   - The value
//   - A boolean flag indicating whether the value was found
//   - The name of the environment variable
//   - And an error if any
func (f *Flag[T]) envValue() (value T, found bool, envName string, _ error) {
	var empty T // readonly value

	for _, name := range f.EnvVars {
		if envValue, ok := os.LookupEnv(name); ok {
			envValue = strings.Trim(envValue, " \t\n\r") // trim unnecessary characters

			parsed, err := f.parseString(envValue)
			if err != nil {
				return empty, true, name, err // empty value, found, env name, error
			}

			return parsed, true, name, nil // parsed value, found, env name, no error
		}
	}

	return empty, false, "", nil // empty value, not found, no env name, no error
}

func (f *Flag[T]) setValue(v T, src flagValueSource) {
	if f.Value == nil {
		f.Value = new(T)
	}

	*f.Value, f.ValueSetFrom = v, src
}

func (f *Flag[T]) Apply(s *flag.FlagSet) error {
	// set the default flag value
	f.setValue(f.Default, FlagValueSourceDefault)

	var envParsingErr error

	// get the value from the environment variable (before flag parsing)
	if v, found, envName, err := f.envValue(); found && err == nil {
		f.setValue(v, FlagValueSourceEnv)
	} else if err != nil {
		// store the error for later
		envParsingErr = fmt.Errorf("failed to parse the environment variable %s: %w", envName, err)
	}

	switch any(*new(T)).(type) {
	case bool:
		var fn = func(string) error {

			// since we have a boolean flag, we need to set the value to true if the flag was provided
			// without taking into account the value
			f.setValue(any(true).(T), FlagValueSourceFlag)

			return envParsingErr
		}

		for _, name := range f.Names {
			s.BoolFunc(name, f.Usage, fn)
		}
	case
		int,
		int64,
		string,
		uint,
		uint64,
		float64,
		time.Duration:
		var fn = func(in string) error {
			if v, err := f.parseString(in); err == nil {
				f.setValue(v, FlagValueSourceFlag)
			} else {
				return err
			}

			return envParsingErr
		}

		for _, name := range f.Names {
			s.Func(name, f.Usage, fn)
		}
	}

	return nil
}

func (f *Flag[T]) Validate(c *Command) error {
	if f.Validator == nil {
		return nil
	}

	if f.Value == nil {
		return errors.New("flag value is nil")
	}

	return f.Validator(c, *f.Value)
}

func (f *Flag[T]) RunAction(c *Command) error {
	if f.Action == nil {
		return nil
	}

	if f.Value == nil {
		return errors.New("flag value is nil")
	}

	return f.Action(c, *f.Value)
}
