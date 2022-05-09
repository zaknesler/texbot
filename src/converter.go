// Author: Zak Nesler
// Date: 2022-05-07
//
// This file is a part of TexBot. Its main purpose is to render a LaTeX
// expression using the following CLI commands:
//
//   - latex (for parsing and rendering to DVI file)
//   - dvisvgm (for converting DVI file to SVG file)
//   - inkscape (for converting SVG file to PNG file)
//
// All of the above commands are run in a temporary directory that is removed
// after the expression is rendered and the image file is in memory. These
// commands are all installed and configured in the Docker image.

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"regexp"
)

// RenderResult is the result of a render operation
// Contains a file pointer, a possible error, and a possible LaTeX parse error message
type RenderResult struct {
	File        *os.File
	Err         error
	ParseErrMsg string
}

// Render a LaTeX expression string and return a RenderResult
func Render(in ParsedString) RenderResult {
	// Create temporary directory to operate on LaTeX expression
	dir, err := os.MkdirTemp("", "texbot")
	if err != nil {
		return RenderResult{nil, err, ""}
	}

	log.Printf("Rendering expression '%s'", in.Expr)

	// Remove temp directory when finished
	defer os.RemoveAll(dir)

	// Create file to store LaTeX source
	f, err := os.Create(path.Join(dir, "source.tex"))

	log.Printf("Writing temp source file: %s", f.Name())

	// Configure LaTeX document
	// Includes math libraries, sets the color/sizing, etc.
	f.WriteString(fmt.Sprintf(`
        \documentclass{standalone}
        \usepackage{amsmath}
        \usepackage{amssymb}
        \usepackage{amsfonts}
        \usepackage{xcolor}
        \usepackage{siunitx}
        \usepackage[dvips]{graphicx}
        \usepackage[utf8]{inputenc}
        \thispagestyle{empty}
        \begin{document}
        \color{white}
        $%s$
        \end{document}`, in.Expr))

	f.Close()

	// Convert LaTeX expression to DVI file
	out, err := exec.Command(
		"latex",
		"-no-shell-escape",
		"-interaction=nonstopmode",
		"-halt-on-error",
		"-output-directory="+dir,
		"-output-format=dvi",
		f.Name()).Output()

	// If there was an error with parsing the LaTeX expression,
	// get the error message written to stdout, and conver to basic string
	if err != nil {
		return RenderResult{nil, err, getLatexParseError(string(out))}
	}

	// Convert DVI file to SVG file
	outFile := f.Name() + ".svg"
	cmd := exec.Command(
		"dvisvgm",
		"--no-fonts",
		"--exact-bbox",
		"-Z4",
		"--output="+outFile,
		path.Join(dir, "source.dvi"))

	// Return error in rare case that dvisvgm fails
	if err := cmd.Run(); err != nil {
		return RenderResult{nil, err, ""}
	}

	// Convert SVG file as PNG
	outFile = f.Name() + ".png"
	cmd = exec.Command(
		"inkscape",
		f.Name()+".svg",
		fmt.Sprintf("--export-height=%d", in.Config.Scale*50),
		"--export-png="+outFile)

	// Return error in rare case that inkscape conversion fails
	if err := cmd.Run(); err != nil {
		return RenderResult{nil, err, ""}
	}

	// Lastly, open final image file and return it
	f, err = os.Open(outFile)
	if err != nil {
		return RenderResult{nil, err, ""}
	}

	// Return successful file with no errors
	return RenderResult{f, nil, ""}
}

func getLatexParseError(output string) string {
	// Parse output for LaTeX parsing error
	// Grab the error between known strings
	search := regexp.MustCompile(`\n! ([\w\W]+)\nNo pages`)
	matches := search.FindAllStringSubmatch(output, -1)

	// If no matches, return empty string
	if len(matches) == 0 {
		return ""
	}

	// Return the LaTeX error string
	return matches[0][1]
}
