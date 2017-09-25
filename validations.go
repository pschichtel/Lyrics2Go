package main

import (
	"golang.org/x/text/encoding"
	"strings"
	"regexp"
)

type ValidationFunc func(string, []string, encoding.Encoding) bool

func parseValidationName(name string) (string, bool) {
	const NOT_PREFIX = "not "
	const EXC_PREFIX = "!"
	lower := strings.ToLower(name)

	if strings.Index(lower, NOT_PREFIX) == 0 {
		return lower[len(NOT_PREFIX):], true
	} else if strings.Index(lower, EXC_PREFIX) == 0 {
		return lower[len(EXC_PREFIX):], true
	} else {
		return lower, false
	}

}

func validateValue(value string, validationList [][]string, validationFuncs map[string]ValidationFunc, e encoding.Encoding) bool {

	for _, validation := range validationList {
		if len(validation) > 0 {
			name, inverted := parseValidationName(validation[0])
			args := validation[1:]

			validationFunc, found := validationFuncs[strings.ToLower(name)]
			if found {
				result := validationFunc(value, args, e)
				if inverted && result || !inverted && !result {
					logLine("Lyrics failed a validation!")
					return false
				}
			} else {
				logLine("Unknown filter: %s", name)
			}
		}
	}

	return true
}

func contains(in string, args []string, _ encoding.Encoding) bool {
	if len(args) < 1 {
		logLine("contains validator is missing the string to test!")
		return false
	}
	return strings.Contains(in, args[0])
}

func containsIgnoreCase(in string, args []string, _ encoding.Encoding) bool {
	if len(args) < 1 {
		logLine("contains validator is missing the string to test!")
		return false
	}
	return strings.Contains(strings.ToLower(in), strings.ToLower(args[0]))
}

func matches(in string, args []string, _ encoding.Encoding) bool {
	if len(args) < 1 {
		logLine("matches validator is missing the regex to match!")
		return false
	}
	compiled, err := regexp.Compile(args[0])
	if err != nil {
		logLine("matches validator failed to compile the regex: %s", err)
		return false
	}
	return compiled.MatchString(in)
}