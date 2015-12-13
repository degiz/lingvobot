package main

import (
	"fmt"
	"gopkg.in/xmlpath.v2"
	"log"
	"net/http"
	"regexp"
	"strings"
)

func convertArticle(symbol string) string {
	switch symbol {
	case "m":
		return "der"
	case "f":
		return "die"
	case "n":
		return "das"
	case "0":
		return "die (pl)"
	}
	return ""
}

func getNounInfo(noun string) []string {

	type nounsRegex map[string]string

	var finalResult []string

	var nouns []nounsRegex
	nouns = append(nouns, nounsRegex{
		"name":   "Gender",
		"regexp": "\\|Genus=(\\w)"})
	nouns = append(nouns, nounsRegex{
		"name":   "Nominativ Singular",
		"regexp": "\\|Nominativ Singular(?:.*)=(.*\\b)"})
	nouns = append(nouns, nounsRegex{
		"name":   "Nominativ Plural",
		"regexp": "\\|Nominativ Plural(?:.*)=(.*\\b)"})
	nouns = append(nouns, nounsRegex{
		"name":   "Akkusativ Singular",
		"regexp": "\\|Akkusativ Singular(?:.*)=(.*\\b)"})
	nouns = append(nouns, nounsRegex{
		"name":   "Akkusativ Plural",
		"regexp": "\\|Akkusativ Plural(?:.*)=(.*\\b)"})
	nouns = append(nouns, nounsRegex{
		"name":   "Dativ Singular",
		"regexp": "\\|Dativ Singular(?:.*)=(.*\\b)"})
	nouns = append(nouns, nounsRegex{
		"name":   "Dativ Plural",
		"regexp": "\\|Dativ Plural(?:.*)=(.*\\b)"})
	nouns = append(nouns, nounsRegex{
		"name":   "Genitiv Singular",
		"regexp": "\\|Genitiv Singular(?:.*)=(.*\\b)"})
	nouns = append(nouns, nounsRegex{
		"name":   "Genitiv Plural",
		"regexp": "\\|Genitiv Plural(?:.*)=(.*\\b)"})

	template := "https://de.wiktionary.org/w/api.php?action=query&prop=revisions&rvprop=content&format=xml&"

	templateWiki := "https://de.wiktionary.org/wiki"

	query := template + "titles=" + strings.Title(noun)
	queryWiki := templateWiki + "/" + strings.Title(noun)

	resp, err := http.Get(query)
	if err != nil {
		log.Println("Error processing query:", err)
		return finalResult
	}
	defer resp.Body.Close()

	root, err := xmlpath.Parse(resp.Body)
	if err != nil {
		return finalResult
	}
	path := xmlpath.MustCompile("/api/query/pages/page/revisions/rev")

	wiki, ok := path.String(root)
	if !ok {
		return finalResult
	}

	for _, noun := range nouns {
		var result []string
		regex, _ := regexp.Compile(noun["regexp"])
		submatches := regex.FindAllStringSubmatch(wiki, -1)
		if submatches == nil {
			continue
		}
		for _, submatch := range submatches {
			result = append(result, submatch[1])
		}
		text := fmt.Sprintf("%s: %v", noun["name"], result)
		finalResult = append(finalResult, text)
	}

	finalResult = append(finalResult, queryWiki)

	return finalResult

}
