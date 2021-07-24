package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"

	"gopkg.in/yaml.v2"
)

type Result struct {
	ExitCode int
	Error    string
	Yaml     string
}

func errorResult(msg string) Result {
	return Result{
		ExitCode: 1,
		Error:    msg,
	}
}

func successResult(yamlStr string) Result {
	return Result{
		ExitCode: 0,
		Yaml:     yamlStr,
	}
}

type ParseResult struct {
	Exact bool
	Path  string
	Data  interface{}
}

func parse(data interface{}, path, search string) []ParseResult {
	result := []ParseResult{}

	val := reflect.ValueOf(data)
	switch val.Kind() {
	case reflect.Map:
		iter := val.MapRange()
		for iter.Next() {
			k := iter.Key().Interface().(string)
			v := iter.Value().Interface()

			innerPath := fmt.Sprintf("%s.%s", path, k)

			if strings.Contains(k, search) {
				result = append(result, ParseResult{
					Exact: k == search,
					Path:  innerPath,
					Data:  v,
				})
				continue
			}
			result = append(result, parse(v, innerPath, search)...)
		}
	case reflect.Slice:
		for i, v := range val.Interface().([]interface{}) {
			innerPath := fmt.Sprintf("%s[%d]", path, i)
			result = append(result, parse(v, innerPath, search)...)
		}
	}

	return result
}

func app(stdin []byte, search string) Result {
	data := make(map[interface{}]interface{})

	err := yaml.Unmarshal([]byte(stdin), &data)
	if err != nil {
		return errorResult(fmt.Sprintf("Could not parse yaml:\n%s", stdin))
	}

	parsed := parse(data, "", search)

	if len(parsed) == 0 {
		return errorResult("Nothing found")
	}

	resultYaml := make(map[interface{}]interface{})

	for _, found := range parsed {
		resultYaml[found.Path] = found.Data
	}

	resultYamlStr, err := yaml.Marshal(&resultYaml)
	if err != nil {
		return errorResult(fmt.Sprintf("Could not stringify yaml:\n%s", data))
	}

	return successResult(string(resultYamlStr))
}

func main() {
	stdinStr, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	result := app(stdinStr, os.Args[1])

	if result.Error != "" {
		fmt.Fprintf(os.Stderr, result.Error)
	}

	if result.Yaml != "" {
		fmt.Fprintf(os.Stdout, result.Yaml)
	}

	os.Exit(result.ExitCode)
}
