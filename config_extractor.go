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
	"strings"
	"unicode"
)

const EnvTagName string = "env"
const DefaultEnvCase = ScreamingSnake
const DefaultCMDArgCase = Snake
const unsetDefaultValue = "true"

type envSource string

const (
	OSEnv   = envSource("OSEnv")
	CMDArgs = envSource("CMDArgs")
)

var defaultSourcePrecedence = []envSource{
	OSEnv,
	CMDArgs,
}

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

	// The default value is Snake. i.e: "any_kind_of_string"
	CMDArgsNameCaseType caseType

	EnvSourcePrecedence []envSource
}

func (e *ExtractorOptions) mergeWithDefault() {
	if e.EnvNameCaseType == caseType("") {
		e.EnvNameCaseType = DefaultEnvCase
	}
	if e.CMDArgsNameCaseType == caseType("") {
		e.CMDArgsNameCaseType = DefaultCMDArgCase
	}
	if e.EnvSourcePrecedence == nil || len(e.EnvSourcePrecedence) == 0 {
		e.EnvSourcePrecedence = defaultSourcePrecedence
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

// Extract system environment variables, parse and save
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

func isValidVarStartName(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

func removeQuotes(val string) string {
	startIdx := 0
	endIdx := len(val)
	if val[0] == '\'' || val[0] == '"' {
		startIdx = 1
	}
	if val[len(val)-1] == '\'' || val[len(val)-1] == '"' {
		if endIdx > 1 {
			endIdx = len(val) - 1
		}
	}
	return val[startIdx:endIdx]
}

func getEnvValFromCMDArgs(envName string, args []string) string {
	if args == nil || len(args) < 2 {
		return ""
	}
	for _, arg := range args {
		if arg[0] == '-' {
			// case "-param=123" or "-active"
			arg = strings.TrimPrefix(arg, "-")
			// case "--param=123" or "--active"
			arg = strings.TrimPrefix(arg, "-")
		}
		if !isValidVarStartName(rune(arg[0])) {
			continue
		}
		// case "active"
		if arg == envName {
			return unsetDefaultValue
		}
		for _, c := range arg {
			if c == '=' {
				spArgs := strings.SplitN(arg, "=", 2)
				if spArgs[0] != envName {
					continue
				}
				if len(spArgs) > 1 {
					return removeQuotes(spArgs[1])
				} else {
					return unsetDefaultValue
				}
			}
		}
	}
	return ""
}

func getEnvValFromOSEnv(envName string) string {
	val, exist := os.LookupEnv(envName)
	if !exist {
		return ""
	}
	return val
}

func getEnvVal(envName string, opts ExtractorOptions) (string, envSource) {
	val := ""
	var finalSource envSource
	for _, source := range opts.EnvSourcePrecedence {
		switch source {
		case OSEnv:
			envName_ := changeCase(opts.EnvNameCaseType, envName)
			val = getEnvValFromOSEnv(envName_)
		case CMDArgs:
			envName_ := changeCase(opts.CMDArgsNameCaseType, envName)
			val = getEnvValFromCMDArgs(envName_, os.Args)
		}
		if val == "" {
			continue
		}
		finalSource = source
		break
	}
	return val, finalSource
}

func getKeyName(v reflect.Value, i int, idx int, args ExtractorArgs) (string, error) {
	envName, haveTagEnv := v.Type().Field(i).Tag.Lookup(EnvTagName)
	prefix := ""
	if args.Configs[idx+1] != "" {
		var ok bool
		prefix, ok = args.Configs[idx+1].(string)
		if !ok {
			panic("ERROR: Invalid configuration. Expected string")
		}
	}
	if !haveTagEnv {
		if args.Options.OmitNotTagged {
			return "", fmt.Errorf("unable to get the value name: %v", v)
		}
		envName = v.Type().Field(i).Name
	}
	return prefix + envName, nil
}

func extractByIdx(idx int, args ExtractorArgs) error {
	//defConfig := args.DefaultValues[idx]
	v := reflect.ValueOf(args.Configs[idx]).Elem()
	vPtr := v
	// if its a pointer, resolve its value
	if v.Kind() == reflect.Ptr {
		vPtr = reflect.Indirect(v)
	}
	for i := 0; i < vPtr.NumField(); i++ {
		f := v.Field(i)
		// make sure that this field is defined, and can be changed.
		if !f.IsValid() || !f.CanSet() {
			continue
		}
		envName, err := getKeyName(v, i, idx, args)
		if err != nil {
			continue
		}
		envVal, _ := getEnvVal(envName, args.Options)
		if envVal == "" {
			continue
		}
		err = setValue(f, envVal)
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
		field.SetInt(valueInt)
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
