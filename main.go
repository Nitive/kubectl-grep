package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"reflect"
	"strings"

	"github.com/urfave/cli/v2"
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

func ContainsIgnoreCase(str, substr string) bool {
	return strings.Contains(strings.ToLower(str), strings.ToLower(substr))
}

func parse(data interface{}, path string, opts AppOptions) []ParseResult {
	results := []ParseResult{}

	untypedValue := reflect.ValueOf(data)
	switch untypedValue.Kind() {
	case reflect.Map:
		iter := untypedValue.MapRange()
		for iter.Next() {
			mapKey := iter.Key().Interface().(string)
			mapValuue := iter.Value().Interface()

			if mapKey == "status" && !opts.ShowStatus {
				continue
			}

			innerPath := fmt.Sprintf("%s.%s", path, mapKey)
			result := ParseResult{
				Exact: mapKey == opts.Search,
				Path:  innerPath,
				Data:  mapValuue,
			}

			if opts.ExactMatch {
				if mapKey == opts.Search {
					results = append(results, result)
					continue
				}
			} else {

				if opts.IgnoreCase && ContainsIgnoreCase(mapKey, opts.Search) {
					results = append(results, result)
					continue
				}

				if strings.Contains(mapKey, opts.Search) {
					results = append(results, result)
					continue
				}
			}

			results = append(results, parse(mapValuue, innerPath, opts)...)
		}
	case reflect.Slice:
		for i, v := range untypedValue.Interface().([]interface{}) {
			innerPath := fmt.Sprintf("%s[%d]", path, i)

			name := getName(reflect.ValueOf(v))
			if name != "" {
				innerPath = fmt.Sprintf("%s.%s", path, name)
			}

			metadataName := getMetadataName(reflect.ValueOf(v))
			if metadataName != "" {
				innerPath = fmt.Sprintf("%s.%s", path, metadataName)
			}

			results = append(results, parse(v, innerPath, opts)...)
		}
	}

	return results
}

func getString(v reflect.Value) string {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Interface:
		unwrapedValue, ok := v.Interface().(string)
		if !ok {
			return ""
		}
		return unwrapedValue
	default:
		return ""
	}
}

func getProp(v reflect.Value, prop string) reflect.Value {
	value := (v.MapIndex(reflect.ValueOf(prop)))
	if value.Kind() == reflect.Interface {
		return reflect.ValueOf(value.Interface())
	}

	return value
}

func getName(m reflect.Value) string {
	if m.Kind() == reflect.Map {
		return getString(getProp(m, "name"))
	}

	return ""
}

func getMetadataName(v reflect.Value) string {
	if v.Kind() == reflect.Map {
		metadata := getProp(v, "metadata")
		if metadata.Kind() == reflect.Map {
			return getString(getProp(metadata, "name"))
		}
	}

	return ""
}

type AppOptions struct {
	Search     string
	IgnoreCase bool
	ExactMatch bool
	ShowStatus bool
	PassOutput string
}

func app(stdin []byte, opts AppOptions) Result {
	data := make(map[interface{}]interface{})

	err := yaml.Unmarshal([]byte(stdin), &data)
	if err != nil {
		return errorResult(fmt.Sprintf("Could not parse yaml:\n%s", stdin))
	}

	parsed := parse(data, "", opts)

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
	var opts AppOptions

	cliApp := &cli.App{
		Name:    "kubectl-grep",
		Usage:   "find the right peace in kubectl get -o yaml output",
		Version: "0.1.0",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Aliases:     []string{"i"},
				Name:        "ignore-case",
				Usage:       "Perform case insensitive matching. By default, search is case sensitive",
				Destination: &opts.IgnoreCase,
				EnvVars:     []string{"KUBECTL_GREP_IGNORE_CASE"},
			},
			&cli.BoolFlag{
				Aliases:     []string{"e"},
				Name:        "exact",
				Usage:       "Search for exact key matches. By default, it searches all keys which includes search string",
				Destination: &opts.ExactMatch,
				EnvVars:     []string{"KUBECTL_GREP_EXACT"},
			},
			&cli.BoolFlag{
				Aliases:     []string{"s"},
				Name:        "show-status",
				Usage:       "Include results from status field",
				Destination: &opts.ShowStatus,
				EnvVars:     []string{"KUBECTL_GREP_SHOW_STATUS"},
			},
			&cli.StringFlag{
				Aliases:     []string{"p"},
				Name:        "pass-output",
				Usage:       "Pass output to shell script typically to prettify it. Recommended script: `bat --language yaml --style plain --color always`",
				Destination: &opts.PassOutput,
				EnvVars:     []string{"KUBECTL_GREP_PASS_OUTPUT"},
			},
		},
		Action: func(c *cli.Context) error {
			stdinStr, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				log.Fatalf("error: %v", err)
			}

			searchString := c.Args().Get(0)
			if searchString == "" {
				return cli.Exit("Unexpected empty search string", 1)
			}
			opts.Search = searchString

			result := app(stdinStr, opts)

			if result.Yaml != "" {
				if opts.PassOutput != "" {
					cmd := exec.Command("/bin/sh", "-c", opts.PassOutput)
					cmd.Stdin = bytes.NewBuffer([]byte(result.Yaml))
					output, err := cmd.Output()
					if err != nil {
						return cli.Exit(err, result.ExitCode)
					}
					fmt.Fprintf(os.Stdout, string(output))
				} else {
					fmt.Fprintf(os.Stdout, result.Yaml)
				}
			}

			if result.ExitCode != 0 {
				return cli.Exit(result.Error, result.ExitCode)
			} else if result.Error != "" {
				fmt.Fprintf(os.Stderr, result.Error)
			}
			return nil
		},
	}

	cli.AppHelpTemplate = `
{{- .Name }} — find the right peace in kubectl get -o yaml output

Examples:
	# Show images of pods in kube-system namespace
	kubectl get pods -o yaml -n kube-system | {{ .Name }} image --exact

	# Show kernel version on nodes
	kubectl get node -o yaml | {{ .Name }} ker --show-status

	# Show pod's nodeAffinity
	kubectl get pod my-pod -o yaml | ./kubectl-grep nodeAff

USAGE:
  kubectl get pods -o yaml | {{ .Name }} [options] [search string]

Options:
  {{- range .VisibleFlags }}
  {{ . }}
  {{- end }}

VERSION:
  {{.Version}}
`
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("%s\n", c.App.Version)
	}

	err := cliApp.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

}
