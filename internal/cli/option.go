package cli

import "github.com/urfave/cli/v3"

type option[T any] struct {
	Value T
	isSet bool
}

// Set sets the option value.
func (o *option[T]) Set(v T) { o.Value, o.isSet = v, true }

// SetIfNotNil sets the option value if the provided value is not nil.
func (o *option[T]) SetIfNotNil(v *T) {
	if v != nil {
		o.Set(*v)
	}
}

// SetFromFlagIfUnset sets the option value if:
//   - Flag value was provided by the user
//   - The option is not set yet
func (o *option[T]) SetFromFlagIfUnset(c *cli.Command, flagName string, getter func(flagName string) T) {
	if c.IsSet(flagName) || !o.isSet {
		o.Set(getter(flagName))
	}
}
