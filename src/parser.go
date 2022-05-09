// Author: Zak Nesler
// Date: 2022-05-07
//
// This file is a part of TexBot. Its job is to parse a string and return an
// object that contains the expression as well as the configuration options
// that will be used to render it (scale, etc).

package main

import (
	"regexp"
	"strconv"
	"strings"
)

// Placeholders for regular expressions used to match LaTeX expressions
var SEARCH_EXPR *regexp.Regexp
var SEARCH_EXPR_CONFIG *regexp.Regexp

// ParsedString is a struct that contains a parsed and cleaned LaTeX expression
// sans any surrounding whitespace or formatting symbols.
type ParsedString struct {
	Expr     string
	Config   Config
	HasMatch bool
}

// Config object for ParsedString
type Config struct {
	Scale int
}

// Initialize the regular expressions
func init() {
	// Search string for surrounding $$ symbols and an optional scale argument
	// surrounded by square brackets. (e.g. $$ \frac{1}{2} $$[3])
	SEARCH_EXPR = regexp.MustCompile(`\$\$\s*([^$]+)\s*\$\$(\[(\d+)\])?`)
}

// Takes a raw string and parses it for a single LaTeX expression
func ParseString(input string) ParsedString {
	// Match the input string against the regular expression
	match := SEARCH_EXPR.FindStringSubmatch(input)
	parsed := ParsedString{}

	// If there is no match, return an empty ParsedString with HasMatch set to false
	if len(match) == 0 {
		return ParsedString{"", Config{}, false}
	}

	// Trim the expression (the first match)
	parsed.Expr = strings.Trim(match[1], " ")

	// If after trimming there is no expression, it was empty, so return empty ParsedString
	if len(parsed.Expr) == 0 {
		return ParsedString{"", Config{}, false}
	}

	// If there is a scale argument, set it in the Config object, otherwise set it to 4
	if len(match[3]) != 0 {
		parsed.Config.Scale, _ = strconv.Atoi(match[3])
	} else {
		parsed.Config.Scale = 4
	}

	// Set the HasMatch flag to true and return the ParsedString
	parsed.HasMatch = true
	return parsed
}
