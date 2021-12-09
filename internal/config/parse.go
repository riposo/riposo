package config

import (
	"encoding"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type parseFunc func(interface{}) error

func newParseFunc(configFile string, env Env) parseFunc {
	if env == nil {
		env = OSEnv{}
	}

	return func(v interface{}) error {
		rv := reflect.ValueOf(v)
		if rv.Kind() != reflect.Ptr {
			return fmt.Errorf("config must be a pointer, but received %T", v)
		}

		if rv = rv.Elem(); rv.Kind() != reflect.Struct {
			return fmt.Errorf("config must be a struct, but received %T", v)
		}

		if err := parseYAML(configFile, v); err != nil {
			return err
		}

		if err := parseEnv(env, "RIPOSO_", rv); err != nil {
			return err
		}

		return nil
	}
}

// ----------------------------------------------------------------------------

func parseYAML(fname string, v interface{}) error {
	if fname == "" {
		return nil
	}

	f, err := os.Open(fname)
	if err != nil {
		return fmt.Errorf("unable to read config file %q: %w", fname, err)
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(v); err != nil {
		return fmt.Errorf("unable to parse config file %q: %w", fname, err)
	}
	return nil
}

func parseEnv(env Env, prefix string, rv reflect.Value) error {
	rt := rv.Type()

	// iterate over each struct field
	for i := 0; i < rt.NumField(); i++ {
		if err := setField(env, prefix, rv.Field(i), rt.Field(i)); err != nil {
			return err
		}
	}

	return nil
}

// Elements here are inspired by github.com/kelseyhightower/envconfig.
// Copyright Kelsey Hightower (MIT License).
func setField(env Env, prefix string, fv reflect.Value, sf reflect.StructField) error {
	// skip if cannot be set
	if !fv.CanSet() {
		return nil
	}

	// dereference pointers and init values
	for fv.Type().Kind() == reflect.Ptr {
		if fv.IsNil() {
			// nil pointer to a non-struct: leave it alone
			if fv.Type().Elem().Kind() != reflect.Struct {
				break
			}
			fv.Set(reflect.New(fv.Type().Elem()))
		}

		fv = fv.Elem()
	}

	// get field name, use yaml tags if present
	name := sf.Tag.Get("yaml")
	if name == "-" {
		return nil
	} else if name == "" {
		name = sf.Name
	}
	name = strings.ToUpper(name)
	fullName := prefix + name

	// parse nested struct
	if fv.Kind() == reflect.Struct {
		if _, ok := textUnmarshaler(fv); !ok {
			if err := parseEnv(env, fullName+"_", fv); err != nil {
				return err
			}
			return nil
		}
	}

	// get value from env
	val := env.Get(fullName)
	if val == "" {
		// skip if already set
		if !fv.IsZero() {
			return nil
		}

		// fallback on default
		val = sf.Tag.Get("default")
	}

	// skip if no value
	if val == "" {
		return nil
	}

	// set encoding.TextUnmarshaler
	if u, ok := textUnmarshaler(fv); ok {
		if err := u.UnmarshalText([]byte(val)); err != nil {
			return fmt.Errorf("%s: %w", fullName, err)
		}
		return nil
	}

	// set simple values
	if err := setFieldValue(fv, val); err != nil {
		return fmt.Errorf("%s: %w", fullName, err)
	}

	return nil
}

func textUnmarshaler(fv reflect.Value) (encoding.TextUnmarshaler, bool) {
	if !fv.CanInterface() {
		return nil, false
	}

	if u, ok := fv.Interface().(encoding.TextUnmarshaler); ok {
		return u, ok
	}

	if fv.CanAddr() {
		if u, ok := fv.Addr().Interface().(encoding.TextUnmarshaler); ok {
			return u, ok
		}
	}
	return nil, false
}

var durationType = reflect.TypeOf(time.Duration(0))

func setFieldValue(fv reflect.Value, val string) error {
	ft := fv.Type()
	fk := ft.Kind()

	switch fk {
	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		fv.SetBool(b)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(val, ft.Bits())
		if err != nil {
			return err
		}
		fv.SetFloat(f)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch ft {
		case durationType: // special case: time.Duration
			d, err := time.ParseDuration(val)
			if err != nil {
				return err
			}
			fv.SetInt(int64(d))
			return nil
		default:
			i, err := strconv.ParseInt(val, 0, ft.Bits())
			if err != nil {
				return err
			}
			fv.SetInt(i)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(val, 0, ft.Bits())
		if err != nil {
			return err
		}
		fv.SetUint(i)
	case reflect.String:
		fv.SetString(val)
	case reflect.Slice:
		if ft.Elem().Kind() == reflect.Uint8 { // special case: []byte
			fv.Set(reflect.ValueOf([]byte(val)))
		} else {
			vals := strings.Split(val, ",")
			vv := reflect.MakeSlice(ft, len(vals), len(vals))
			for i, val := range vals {
				if err := setFieldValue(vv.Index(i), strings.TrimSpace(val)); err != nil {
					return fmt.Errorf("%s: %w", val, err)
				}
			}
			fv.Set(vv)
		}
	case reflect.Map:
		vv := reflect.MakeMap(ft)
		if err := yaml.NewDecoder(strings.NewReader(val)).Decode(vv.Interface()); err != nil {
			return err
		}
		fv.Set(vv)
	default:
		return fmt.Errorf("cannot decode into %v", ft)
	}

	return nil
}
