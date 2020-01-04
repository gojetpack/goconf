package goconf

import "github.com/iancoleman/strcase"

type caseType string

const (
	ScreamingSnake = caseType("ANY_KIND_OF_STRING")
	Snake          = caseType("any_kind_of_string")
	Kebab          = caseType("any-kind-of-string")
	ScreamingKebab = caseType("ANY-KIND-OF-STRING")
	Camel          = caseType("AnyKindOfString")
	LowerCamel     = caseType("anyKindOfString")
)

func changeCase(caseType_ caseType, envName string) string {
	m := map[caseType]func(string) string{
		ScreamingSnake: strcase.ToScreamingSnake,
		Snake:          strcase.ToSnake,
		Kebab:          strcase.ToKebab,
		ScreamingKebab: strcase.ToScreamingKebab,
		Camel:          strcase.ToCamel,
		LowerCamel:     strcase.ToLowerCamel,
	}
	f, ok := m[caseType_]
	if !ok {
		return envName
	}
	return f(envName)
}
