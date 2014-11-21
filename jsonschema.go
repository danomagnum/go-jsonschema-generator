package jsonschema

import (
	"encoding/json"
	"reflect"
	"strings"
)

type JSONSchema struct {
	Type       string                 `json:"type"`
	Items      *JSONSchemaItems       `json:"items,omitempty"`
	Properties map[string]*JSONSchema `json:"properties,omitempty"`
	Required   []string               `json:"required,omitempty"`
}

type JSONSchemaItems struct {
	Type string `json:"type,omitempty"`
}

func (j *JSONSchema) Marshal() ([]byte, error) {
	return json.MarshalIndent(j, "", "  ")
}

func (j *JSONSchema) String() string {
	json, _ := j.Marshal()
	return string(json)
}

func (j *JSONSchema) Load(variable interface{}) {
	value := reflect.ValueOf(variable).Elem()
	j.doLoad(value, tagOptions(""))
}

func (j *JSONSchema) doLoad(value reflect.Value, opts tagOptions) {
	kind := value.Kind()

	j.Type = getTypeFromMapping(kind)
	switch kind {
	case reflect.Slice:
		j.doLoadFromSlice(value)
	case reflect.Map:
		j.doLoadFromMap(value)
	case reflect.Struct:
		j.doLoadFromStruct(value)
	}
}

func (j *JSONSchema) doLoadFromSlice(valueObject reflect.Value) {
	k := valueObject.Type().Elem().Kind()
	if k == reflect.Uint8 {
		j.Type = "string"
	} else {
		j.Items = &JSONSchemaItems{Type: getTypeFromMapping(k)}
	}
}

func (j *JSONSchema) doLoadFromMap(valueObject reflect.Value) {
	k := valueObject.Type().Elem().Kind()
	j.Properties = make(map[string]*JSONSchema, 0)
	j.Properties[".*"] = &JSONSchema{Type: getTypeFromMapping(k)}
}

func (j *JSONSchema) doLoadFromStruct(valueObject reflect.Value) {
	j.Type = "object"
	j.Properties = make(map[string]*JSONSchema, 0)
	typeObject := valueObject.Type()

	count := valueObject.NumField()
	for i := 0; i < count; i++ {
		field := typeObject.Field(i)
		value := valueObject.Field(i)

		tag := field.Tag.Get("json")
		name, opts := parseTag(tag)
		if name == "" {
			name = field.Name
		}

		j.Properties[name] = &JSONSchema{}
		j.Properties[name].doLoad(value, opts)

		if !opts.Contains("omitempty") {
			j.Required = append(j.Required, name)
		}
	}
}

var mapping = map[reflect.Kind]string{
	reflect.Bool:    "bool",
	reflect.Int:     "integer",
	reflect.Int8:    "integer",
	reflect.Int16:   "integer",
	reflect.Int32:   "integer",
	reflect.Int64:   "integer",
	reflect.Uint:    "integer",
	reflect.Uint8:   "integer",
	reflect.Uint16:  "integer",
	reflect.Uint32:  "integer",
	reflect.Uint64:  "integer",
	reflect.Float32: "number",
	reflect.Float64: "number",
	reflect.String:  "string",
	reflect.Slice:   "array",
	reflect.Struct:  "object",
	reflect.Map:     "object",
}

func getTypeFromMapping(k reflect.Kind) string {
	if t, ok := mapping[k]; ok {
		return t
	}

	return ""
}

type tagOptions string

func parseTag(tag string) (string, tagOptions) {
	if idx := strings.Index(tag, ","); idx != -1 {
		return tag[:idx], tagOptions(tag[idx+1:])
	}
	return tag, tagOptions("")
}

func (o tagOptions) Contains(optionName string) bool {
	if len(o) == 0 {
		return false
	}
	s := string(o)
	for s != "" {
		var next string
		i := strings.Index(s, ",")
		if i >= 0 {
			s, next = s[:i], s[i+1:]
		}
		if s == optionName {
			return true
		}
		s = next
	}
	return false
}