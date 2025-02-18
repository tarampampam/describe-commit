package app

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

	Flag[T bool | int | int64 | string | uint | uint64 | float64 | time.Duration] struct {
		Names     []string                // e.g. "config-file", "c"
		Usage     string                  // e.g. "path to the configuration file"
		Default   T                       // default value
		EnvVars   []string                // e.g. "CONFIG_FILE"
		Validator func(*Command, T) error // value validation function
		Action    func(*Command, T) error // an action to run when the flag is set
		Value     *T                      // actual value (set after parsing)
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

func (f *Flag[T]) IsSet() bool {
	if f.Value == nil {
		return false
	}

	return f.Default != *f.Value
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

		if _, ok := any(f.Value).(*bool); !ok {
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

func (f *Flag[T]) Apply(s *flag.FlagSet) error {
	var empty T

	// initialize the value pointer if was not set before
	if f.Value == nil {
		f.Value = &empty
	}

	var valueFromEnv string

	// get the value from the environment variable
	for _, envVar := range f.EnvVars {
		if v, ok := os.LookupEnv(envVar); ok {
			valueFromEnv = strings.Trim(v, " \t\n\r")

			break
		}
	}

	switch any(f.Default).(type) { // use the [Flag.Default] value type to determine the flag type
	case bool:
		var def = any(f.Default).(bool)

		if valueFromEnv != "" {
			if v, err := strconv.ParseBool(valueFromEnv); err == nil {
				def = v
			}
		}

		for _, name := range f.Names {
			s.BoolVar(any(f.Value).(*bool), name, def, f.Usage)
		}
	case int:
		var def = any(f.Default).(int)

		if valueFromEnv != "" {
			if v, err := strconv.Atoi(valueFromEnv); err == nil {
				def = v
			}
		}

		for _, name := range f.Names {
			s.IntVar(any(f.Value).(*int), name, def, f.Usage)
		}
	case int64:
		var def = any(f.Default).(int64)

		if valueFromEnv != "" {
			if v, err := strconv.ParseInt(valueFromEnv, 10, 64); err == nil {
				def = v
			}
		}

		for _, name := range f.Names {
			s.Int64Var(any(f.Value).(*int64), name, def, f.Usage)
		}
	case string:
		var def = any(f.Default).(string)

		if valueFromEnv != "" {
			def = valueFromEnv
		}

		for _, name := range f.Names {
			s.StringVar(any(f.Value).(*string), name, def, f.Usage)
		}
	case uint:
		var def = any(f.Default).(uint)

		if valueFromEnv != "" {
			if v, err := strconv.ParseUint(valueFromEnv, 10, 64); err == nil {
				def = uint(v)
			}
		}

		for _, name := range f.Names {
			s.UintVar(any(f.Value).(*uint), name, def, f.Usage)
		}
	case uint64:
		var def = any(f.Default).(uint64)

		if valueFromEnv != "" {
			if v, err := strconv.ParseUint(valueFromEnv, 10, 64); err == nil {
				def = v
			}
		}

		for _, name := range f.Names {
			s.Uint64Var(any(f.Value).(*uint64), name, def, f.Usage)
		}
	case float64:
		var def = any(f.Default).(float64)

		if valueFromEnv != "" {
			if v, err := strconv.ParseFloat(valueFromEnv, 64); err == nil {
				def = v
			}
		}

		for _, name := range f.Names {
			s.Float64Var(any(f.Value).(*float64), name, def, f.Usage)
		}
	case time.Duration:
		var def = any(f.Default).(time.Duration)

		if valueFromEnv != "" {
			if v, err := time.ParseDuration(valueFromEnv); err == nil {
				def = v
			}
		}

		for _, name := range f.Names {
			s.DurationVar(any(f.Value).(*time.Duration), name, def, f.Usage)
		}
	default:
		return fmt.Errorf("unsupported flag type: %T", f.Value)
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
