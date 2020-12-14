package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

const (
	defaultsnippet string = "# Commented line\n\n@include \"../snippet.pen\""
	defaultcontent string = "\n\nnewsite"
)

func main() {

	var argv []string  = os.Args[1:]

	if len(argv) == 0 {
		printhelp()
	} else {

		switch argv[0] {
			case "new":
				new()
			case "build":
				buildsite()
			default:
				printhelp()
		}
	}
}

func printhelp() {
	fmt.Println("pen, the do it yourself site generator.")
	fmt.Println("Options:")
	fmt.Println("\tnew: generate a new site with default config.")
	fmt.Println("\tbuild: build the new site with config.")
}

func exitif(message string, err error) {
	if err != nil {
		fmt.Println(message)
		os.Exit(1)
	}

}

func new() {
	fmt.Println("pen is making a new site directory...")
	err := ioutil.WriteFile("./snippet.pen", []byte(defaultsnippet), 0644)
	exitif("Unable to make environment file.", err)
	err = ioutil.WriteFile("./content.pen",[]byte(defaultcontent), 0644)
	exitif("Unable to make content file", err)
	err = ioutil.WriteFile("./styles.css",[]byte("@import \"../styles.css\""), 0644)
	exitif("Unable to make css file", err)

}

type snippetPiece struct {
	toMatch *regexp.Regexp;
	toReplace string;
}

func readImports(line string) string {
	includer, _ := regexp.Compile("(?m)^@include \"(.+)\"$")
	matches := includer.FindAllStringSubmatch(line, -1)
	output := line
	for i:= 0; i < len(matches); i ++ {
		includedFile := matches[i][1]
		fileContents, err := ioutil.ReadFile(includedFile)
		exitif(fmt.Sprintf("Unable to include file: %s",includedFile),err)
		filer, err:= regexp.Compile(fmt.Sprintf("(?m)^@include \"%s\"$",regexp.QuoteMeta(includedFile)))
		output = filer.ReplaceAllLiteralString(output, string(fileContents))
	}
	return output
}

func listmatches(file string) string {
	lister, _ := regexp.Compile("(?m)^@list (.+) / (.+)$")
	matches := lister.FindAllStringSubmatch(file, -1)
	for i:=0; i < len(matches); i++{
		list, err := regexp.Compile("(?m)" + matches[i][1])
		exitif(fmt.Sprintf("Bad list regex: %s", matches[i][1]), err)
		var listmatch string;
		sums := list.FindAllString(file, -1)
		for _, match := range sums {
			listmatch += match + "\n"
		}
		listed := list.ReplaceAllString(listmatch, matches[i][2])
		speclist, _ := regexp.Compile("(?m)^@list " + regexp.QuoteMeta(matches[i][1]) + " / " + regexp.QuoteMeta(matches[i][2]) + "$")
		file = speclist.ReplaceAllString(file, listed)
	}
	return file
}


func buildPage(p string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return nil
	}

	pwd, _ := os.Getwd()
	defer os.Chdir(pwd)
	err = os.Chdir(p)
	exitif(fmt.Sprintf("Unable to read directory %s",p),err)
	content, err := ioutil.ReadFile("content.pen")
	if err != nil {
		return nil
	}

	indexfile, err := os.Create("index.html")
	exitif("Unable to write index.html", err)


	defer indexfile.Close()

	fmt.Println("Building page at " + p)
	var snippets []snippetPiece = checksite("snippet.pen")
	replacedText := readImports(string(content))
	replacedText = listmatches(replacedText)

	for _, j := range snippets {
		replacedText = j.toMatch.ReplaceAllString(replacedText, j.toReplace)
	}

	err = ioutil.WriteFile("index.html", []byte(replacedText), 0644)
	exitif("Unable to write to html file", err)

	fmt.Println("Created site at " + p)
	return nil ;
}


func buildsite() {
	err := filepath.Walk(".",buildPage)
	exitif("Unable to access directory",err)
}

func checksite(snippetfile string) []snippetPiece {
	snippetstyle, _ := regexp.Compile(`(?mU)^([^\s/].+)\n\t(.+)$`)
	fullfile, err := ioutil.ReadFile(snippetfile)
	exitif("Unable to read snippet file.", err)
	fullfileString := string(fullfile)
	fullfileString = readImports(fullfileString)
	var matches [][]string = snippetstyle.FindAllStringSubmatch(fullfileString, -1)
	snippets := make([]snippetPiece, len(matches))
	for i:= 0; i < len(matches); i++ {
		snippets[i].toMatch, err = regexp.Compile("(?m)" + matches[i][1])
		exitif(fmt.Sprintf("Bad Regex: %s", matches[i][0]), err)
		snippets[i].toReplace = matches[i][2]
	}

	return snippets
}
