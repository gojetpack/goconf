package goconf

// Hector Oliveros - 2019
// hector.oliveros.leon@gmail.com

import (
	"errors"
	"fmt"
	"github.com/gojetpack/pyos"
	"github.com/joho/godotenv"
	"os"
	"reflect"
	"strconv"
)

const EnvTagName string = "env"
const DefaultCase caseType = ScreamingSnake

type ExtractorOptions struct {
	// The path of the env file
	// i.e: ./env/.env.prod
	EnvFile string

	// If it is true then all those fields that do not
	// contain the "env" tag will be omitted from the process
	OmitNotTagged bool

	// If the field has no tag and OmitNotTagged is false then
	// the field name is converted to this case type and later
	// is searched in the os env.
	// The default value is ScreamingSnake. i.e: "ANY_KIND_OF_STRING"
	EnvNameCaseType caseType
}

func (e *ExtractorOptions) mergeWithDefault() {
	if e.EnvNameCaseType == caseType("") {
		e.EnvNameCaseType = DefaultCase
	}
}

type ExtractorArgs struct {
	// Allow to configure the env extraction
	Options ExtractorOptions

	// It must be an even array of elements.
	// For each tuple:
	//  - The first element will be a pointer to the object in which the configuration will be saved.
	//  - The second element will be the prefix for this configuration
	Configs []interface{}
}

// Extract system environment variables, parse them and save them
func Extract(args ExtractorArgs) error {
	args.Options.mergeWithDefault()
	if len(args.Options.EnvFile) > 0 {
		if !pyos.Path.Exist(args.Options.EnvFile) {
			return errors.New("the configuration file doesnt exist: " + args.Options.EnvFile)
		}
		err := godotenv.Load(args.Options.EnvFile)
		if err != nil {
			return err
		}
	}
	for i := 0; i < len(args.Configs); i += 2 {
		err := extractByIdx(i, args)
		if err != nil {
			return err
		}
	}
	return nil
}

func extractByIdx(idx int, args ExtractorArgs) error {
	//defConfig := args.DefaultValues[idx]
	v := reflect.ValueOf(args.Configs[idx]).Elem()
	vptr := v
	// if its a pointer, resolve its value
	if v.Kind() == reflect.Ptr {
		vptr = reflect.Indirect(v)
	}

	for i := 0; i < vptr.NumField(); i++ {
		f := v.Field(i)
		// make sure that this field is defined, and can be changed.
		if !f.IsValid() || !f.CanSet() {
			continue
		}
		envName, haveTagEnv := v.Type().Field(i).Tag.Lookup(EnvTagName)
		if !haveTagEnv {
			if args.Options.OmitNotTagged {
				continue
			}
			prefix := ""
			if args.Configs[idx+1] != "" {
				var ok bool
				prefix, ok = args.Configs[idx+1].(string)
				if !ok {
					panic("ERROR: Invalid configurarion. Expected string")
				}
			}
			envName = changeCase(args.Options.EnvNameCaseType, prefix+v.Type().Field(i).Name)
		}
		envVal, exist := os.LookupEnv(envName)
		if !exist || envVal == "" {
			continue
		}
		err := setValue(f, envVal)
		if err != nil {
			return err
		}
	}
	return nil
}

func setValue(field reflect.Value, value interface{}) error {
	switch field.Kind() {
	case reflect.Int, reflect.Int64:
		valueInt, err := strconv.ParseInt(value.(string), 10, 64)
		if err != nil {
			return fmt.Errorf("invalid integer: %v of type %v", value, reflect.TypeOf(value))
		}
		field.SetInt(int64(valueInt))
	case reflect.String:
		field.SetString(value.(string))
	case reflect.Bool:
		val, err := strconv.ParseBool(value.(string))
		if err != nil {
			panic(fmt.Errorf("invalid boolean value %v", value))
		}
		field.SetBool(val)
	}
	return nil
}
