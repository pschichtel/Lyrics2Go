package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"gopkg.in/yaml.v2"
	"regexp"
	"strings"
	"net/url"
	"html"
)

const LOADER_STATIC = "static"

type LoaderConf struct {
	Loader string
}

func logLine(msg string, args ...interface{}) {
	os.Stderr.WriteString(fmt.Sprintf(msg + "\n", args...))
}

func extractGroupIndex(regex *regexp.Regexp, name string) int {
	needle := strings.ToLower(name)

	for i, name := range regex.SubexpNames() {
		if strings.ToLower(name) == needle {
			return i
		}
	}

	return -1
}

func parseArguments() (string, map[string]string, error) {
	if len(os.Args) < 3 {
		return "", nil, fmt.Errorf("usage: %s provider (name=value)+", os.Args[0])
	}

	providerFile := os.Args[1]

	variableValues := make(map[string]string)
	for _, arg := range os.Args[2:] {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) > 1 {
			name := strings.ToLower(parts[0])
			value := parts[1]
			variableValues[name] = value
		}
	}

	return providerFile, variableValues, nil

}

func main() {

	file, variables, err := parseArguments()
	if err != nil {
		logLine(err.Error())
		os.Exit(1)
		return
	}

	filters := map[string]FilterFunc{
		"lowercase":     simple(strings.ToLower),
		"uppercase":     simple(strings.ToUpper),
		"trim":          simple(strings.TrimSpace),
		"urlencode":     simple(url.PathEscape),
		"entity_decode": simple(html.UnescapeString),
		"strip_html":     	   strip_pattern("(?is)<[a-z][a-z0-9]*(\\s+[a-z-]+(=(\"[^\"]*\"|'[^']*'|[^\\s\"'/>]+))?)*\\s*/?>|</[a-z][a-z0-9]*>"),
		"strip_html_comments": strip_pattern("(?s)<!--.*?-->"),
		"strip_links":         strip_pattern("(?i)https?://.*?(\\s|$)"),
		"strip_nonascii":      strip_pattern("(?i)[^a-z0-9]+"),
		"br2nl":   to_newline("(?i)<br\\s*/?>", 1),
		"p2break": to_newline("(?is)<p[^/>]*/?>(\\s*</p>)?", 2),
		"regex":   replace_regex,
		"replace": replace_string,
		"strip_diacritics": diacritics2ascii,
		"umlauts2ascii":    umlauts2ascii,
		"clean_spaces": cleanup_whitespace(),
		"utf8_encode": utf8_encode,
	}

	validations := map[string]ValidationFunc{
		"matches": matches,
		"contains": contains,
		"contains_ci": containsIgnoreCase,
	}

	logLine("Reading config file: %s", file)
	fileBytes, err := ioutil.ReadFile(file)
	if err != nil {
		logLine("Failed to read file %s: %s", file, err)
	} else {
		loaderConf := LoaderConf{Loader:LOADER_STATIC}
		err := yaml.Unmarshal(fileBytes, &loaderConf)
		if err != nil {
			logLine("Failed to parse config: %s", err)
		} else {
			switch loaderConf.Loader {
			case "static":
				providerConf := StaticProviderConf{
					MaxRedirects:10,
				}
				yaml.Unmarshal(fileBytes, &providerConf)
				logLine("Loaded config: %s", providerConf.Name)
				logLine("Running... ")
				err := loadStatically(providerConf, variables, filters, validations)
				if err != nil {
					logLine("Error: %s", err)
					os.Exit(1)
				} else {
					logLine("Success!")
				}
			}
		}
	}

}
