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

func submatchTextWithRegex(text string, inRegex *regexp.Regexp) []string {
	var result []string
	submatches := inRegex.FindAllStringSubmatch(text, -1)
	if submatches == nil {
		return result
	}
	for _, submatch := range submatches {
		result = append(result, submatch[1])
	}
	return result
}

func initWikiRegexps() *WikiRegexps {

	var result WikiRegexps
	type nounsRegex map[string]string

	// We keep it separated in case the format changes
	nounInfoRegexps := []nounsRegex{
		nounsRegex{
			"name":   "[Gn]",
			"regexp": "\\|Genus=(\\w)",
		},
		nounsRegex{
			"name":   "[NS]",
			"regexp": "\\|Nominativ Singular(?:.*)=(.*\\b)",
		},
		nounsRegex{
			"name":   "[NP]",
			"regexp": "\\|Nominativ Plural(?:.*)=(.*\\b)",
		},
		nounsRegex{
			"name":   "[AS]",
			"regexp": "\\|Akkusativ Singular(?:.*)=(.*\\b)",
		},
		nounsRegex{
			"name":   "[AP]",
			"regexp": "\\|Akkusativ Plural(?:.*)=(.*\\b)",
		},
		nounsRegex{
			"name":   "[DS]",
			"regexp": "\\|Dativ Singular(?:.*)=(.*\\b)",
		},
		nounsRegex{
			"name":   "[DP]",
			"regexp": "\\|Dativ Plural(?:.*)=(.*\\b)",
		},
		nounsRegex{
			"name":   "[GS]",
			"regexp": "\\|Genitiv Singular(?:.*)=(.*\\b)",
		},
		nounsRegex{
			"name":   "[GP]",
			"regexp": "\\|Genitiv Plural(?:.*)=(.*\\b)",
		},
	}

	for _, regex := range nounInfoRegexps {
		compiledRegex, _ := regexp.Compile(regex["regexp"])
		result.nounInfo = append(result.nounInfo, Regexp{
			name: regex["name"], value: compiledRegex})
	}

	return &result
}

func getNounInfo(noun string, wikiRegexps *WikiRegexps) []string {

	var result []string

	template := "https://de.wiktionary.org/w/api.php?action=query&prop=revisions&rvprop=content&format=xml&"

	templateWiki := "https://de.wiktionary.org/wiki"

	query := template + "titles=" + strings.Title(noun)
	queryWiki := templateWiki + "/" + strings.Title(noun)

	resp, err := http.Get(query)
	if err != nil {
		log.Println("Error processing query:", err)
		return result
	}
	defer resp.Body.Close()

	root, err := xmlpath.Parse(resp.Body)
	if err != nil {
		return result
	}
	path := xmlpath.MustCompile("/api/query/pages/page/revisions/rev")

	wiki, ok := path.String(root)
	if !ok {
		return result
	}

	for _, regex := range wikiRegexps.nounInfo {
		match := submatchTextWithRegex(wiki, regex.value)
		text := fmt.Sprintf("%s: %s", regex.name, strings.Join(match[:], ", "))
		result = append(result, text)
	}

	result = append(result, queryWiki)

	return result

}
