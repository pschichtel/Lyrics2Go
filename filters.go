package main

import (
	"regexp"
	"strings"
	"fmt"
	"unicode"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"golang.org/x/text/runes"
	"golang.org/x/text/encoding"
)

type FilterFunc func(string, []string, encoding.Encoding) string

func filterValue(in string, filterList [][]string, filterFuncs map[string]FilterFunc, e encoding.Encoding) string {
	out := in
	//logLine("Start: %s", out)
	for _, filter := range filterList {
		if len(filter) > 0 {
			name := filter[0]
			args := filter[1:]

			filterFunc, found := filterFuncs[strings.ToLower(name)]
			if found {
				out = filterFunc(out, args, e)
				//logLine("Now: %s", out)
			} else {
				logLine("Unknown filter: %s", name)
			}
		}
	}

	return out
}

func normalizeArg(in string) string {
	return strings.ToLower(strings.TrimSpace(in))
}

func isArg(in string, arg string) bool {
	return normalizeArg(in) == arg
}

func simple(f func(string) string) FilterFunc {
	return func(in string, _ []string, _ encoding.Encoding) string {
		return f(in)
	}
}

func strip_pattern(pattern string) FilterFunc {
	compiled := regexp.MustCompile(pattern)
	return func(in string, args []string, _ encoding.Encoding) string {
		replacement := ""
		if len(args) > 0 {
			replacement = args[0]
		}
		if len(args) > 1 && len(replacement) > 0 && isArg(args[1], "duplicate") {
			return compiled.ReplaceAllStringFunc(in, func(m string) string {
				return strings.Repeat(replacement, len(m))
			})
		} else {
			return compiled.ReplaceAllString(in, replacement)
		}
	}
}

func to_newline(pattern string, n int) FilterFunc {
	compiled := regexp.MustCompile(pattern)
	return func(in string, _ []string, _ encoding.Encoding) string {
		return compiled.ReplaceAllString(in, strings.Repeat("\n", n))
	}
}

func replace_regex(in string, args []string, _ encoding.Encoding) string {
	if len(args) < 1 {
		fmt.Println("regex filter requires at least one argument! (the regex to match)")
		return in
	}
	replacement := ""
	if len(args) > 1 {
		replacement = args[1]
	}
	regex := args[0]
	compiled, err := regexp.Compile(regex)
	if err != nil {
		fmt.Printf("Failed to compile the regex: %s\n", regex)
		fmt.Printf("Error: %s\n", err)
	}

	return compiled.ReplaceAllString(in, replacement)
}

func replace_string(in string, args []string, _ encoding.Encoding) string {
	if len(args) < 1 {
		fmt.Println("replace filter requires at least one argument! (the string to search)")
		return in
	}
	replacement := ""
	if len(args) > 1 {
		replacement = args[1]
	}

	return strings.Replace(in, args[0], replacement, -1)
}

func diacritics2ascii(in string, _ []string, _ encoding.Encoding) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, err := transform.String(t, in)
	if err != nil {
		fmt.Printf("Failed to remove diacritics: %s\n", err)
		return in
	}
	return result
}

func umlauts2ascii(in string, _ []string, _ encoding.Encoding) string {
	out := in

	out = strings.Replace(out, "ä", "ae", -1)
	out = strings.Replace(out, "ö", "oe", -1)
	out = strings.Replace(out, "ü", "ue", -1)
	out = strings.Replace(out, "ß", "ss", -1)

	return out
}

func cleanup_whitespace() FilterFunc {

	cleanupNewlineRegex := regexp.MustCompile("\n{3,}")
	cleanupSpacesRegex := regexp.MustCompile(" {2,}")

	return func(in string, _ []string, _ encoding.Encoding) string {
		out := in

		out = strings.Replace(out, "\t", " ", -1)
		out = strings.Replace(out, "\v", " ", -1)
		out = strings.Replace(out, "\r\r", "\n", -1)
		out = strings.Replace(out, "\r", "\n", -1)

		out = cleanupNewlineRegex.ReplaceAllString(out, "\n\n")
		out = cleanupSpacesRegex.ReplaceAllString(out, " ")

		return out
	}
}

func utf8_encode(in string, _ []string, e encoding.Encoding) string {
	decoded, err := e.NewDecoder().String(in)
	if err != nil {
		fmt.Printf("Failed to decode string to UTF-8!\n")
		return in
	}
	return decoded

}