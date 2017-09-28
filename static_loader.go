package main

import (
	"regexp"
	"fmt"
	"golang.org/x/net/html/charset"
	"strings"
	"net/http"
	"io/ioutil"
	"strconv"
	"os"
)

const EXTRACTOR_REGEX_GROUP = "lyrics"

type VariableDefinition struct {
	Name    string
	Filters [][]string
	Lookup  map[string]string
}

type Header struct {
	Name  string
	Value string
}

type StaticProviderConf struct {
	Name         string
	Url          string
	Extractor    string
	MaxRedirects int        `yaml:"max-redirects"`
	Variables    []VariableDefinition
	Headers      []Header
	Filters      [][]string
	Validations  [][]string
}

func compileUrl(template string, vars map[string]string) string {
	regex := regexp.MustCompile("{([^}]+)}")
	return regex.ReplaceAllStringFunc(template, func(m string) string {
		rawName := m[1:len(m) - 1]
		name := strings.ToLower(rawName)
		value, found := vars[name]
		if found {
			return value
		} else {
			logLine("Unknown variable %s in URL!", rawName)
			return m
		}
	})
}

func checkRedirect(maxRedirects int) func(*http.Request, []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		logLine("Redirected to: %s", req.URL)
		if len(via) >= maxRedirects {
			return fmt.Errorf("stopped after %d redirects", maxRedirects)
		}
		return nil
	}
}

func processVars(conf StaticProviderConf, values map[string]string, filters map[string]FilterFunc) map[string]string {

	utf8, _ := charset.Lookup("utf-8")
	vars := make(map[string]string)
	for _, variable := range conf.Variables {
		name := strings.ToLower(variable.Name)
		value, found := values[name]
		if !found {
			value = ""
		}
		if variable.Lookup != nil {
			replacement, found := variable.Lookup[value]
			if found {
				vars[name] = replacement
				continue
			}
		}
		vars[name] = filterValue(value, variable.Filters, filters, utf8)
	}
	return vars
}

func buildRequest(conf StaticProviderConf, variables map[string]string) (*http.Request, error) {


	compiledUrl := compileUrl(conf.Url, variables)

	logLine("Compiled URL: %s", compiledUrl)

	req, err := http.NewRequest("GET", compiledUrl, nil)
	if err != nil {
		return nil, err
	}

	for _, header := range conf.Headers {
		req.Header.Add(header.Name, header.Value)
	}
	cookie, cookieFoundInEnv := os.LookupEnv("STATIC_COOKIE_HEADER")
	if cookieFoundInEnv {
		req.Header.Add("Cookie", cookie)
	}

	return req, nil
}

func loadStatically(conf StaticProviderConf, values map[string]string, filters map[string]FilterFunc, validations map[string]ValidationFunc) error {

	extractorRegex, err := regexp.Compile(conf.Extractor)
	if err != nil {
		return err
	}

	extractorGroupIndex := extractGroupIndex(extractorRegex, EXTRACTOR_REGEX_GROUP)
	if extractorGroupIndex < 0 {
		return fmt.Errorf("group %s not defined in extractor regex", EXTRACTOR_REGEX_GROUP)
	}


	for name, value := range values {
		logLine("Variable: %s=%s", name, value)
	}

	client := &http.Client{
		CheckRedirect: checkRedirect(conf.MaxRedirects),
	}

	request, err := buildRequest(conf, processVars(conf, values, filters))
	if err != nil {
		return err
	}

	res, err := client.Do(request)
	if err != nil {
		return err
	}

	if res.StatusCode == 200 {
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			panic(err)
		}
		defer res.Body.Close()

		contentType := res.Header.Get("Content-Type")
		if len(contentType) == 0 {
			contentType = "text/html"
		}

		detectedEncoding, name, certain := charset.DetermineEncoding(bodyBytes, contentType)
		logLine("Detected encoding: %s (Certain: %s)", name, strconv.FormatBool(certain))


		match := extractorRegex.FindStringSubmatch(string(bodyBytes))
		if match == nil {
			return fmt.Errorf("lyrics not found in response")
		}

		rawLyrics := match[extractorGroupIndex]
		lyrics := filterValue(rawLyrics, conf.Filters, filters, detectedEncoding)

		if validateValue(lyrics, conf.Validations, validations, detectedEncoding) {
			fmt.Println(lyrics)
			return nil
		} else {
			return fmt.Errorf("validation failed")
		}
	} else {
		return fmt.Errorf("HTTP Error %d: %s", res.StatusCode, res.Status)
	}

}
