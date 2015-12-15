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

	verbInfoRegexps := []nounsRegex{
		nounsRegex{
			"name":   "[Präsens ich]",
			"regexp": "\\|Präsens_ich=(\\S*\\b)",
		},
		nounsRegex{
			"name":   "[Präsens du]",
			"regexp": "\\|Präsens_du=(\\S*\\b)",
		},
		nounsRegex{
			"name":   "[Präsens er,sie,es]",
			"regexp": "\\|Präsens_er, sie, es=(\\S*\\b)",
		},
		nounsRegex{
			"name":   "[Präteritum ich]",
			"regexp": "\\|Präteritum_ich=(\\S*\\b)",
		},
		nounsRegex{
			"name":   "[Partizip II]",
			"regexp": "\\|Partizip II=(\\S*\\b)",
		},
		nounsRegex{
			"name":   "[Konjunktiv II ich]",
			"regexp": "\\|Konjunktiv II_ich=(\\S*\\b)",
		},
		nounsRegex{
			"name":   "[Imperativ Singular]",
			"regexp": "\\|Imperativ Singular=(\\S*\\b)",
		},
		nounsRegex{
			"name":   "[Imperativ Singular*]",
			"regexp": "\\|Imperativ Singular(?:\\*?)=(\\S*\\b)",
		},
		nounsRegex{
			"name":   "[Imperativ Plural]",
			"regexp": "\\|Imperativ Plural=(\\S*\\b)",
		},
	}

	for _, regex := range verbInfoRegexps {
		compiledRegex, _ := regexp.Compile(regex["regexp"])
		result.verbInfo = append(result.verbInfo, Regexp{
			name: regex["name"], value: compiledRegex})
	}

	return &result
}

func makeWikiRequest(word string) string {
	result := ""
	template := "https://de.wiktionary.org/w/api.php?action=query&prop=revisions&rvprop=content&format=xml&"

	query := template + "titles=" + word

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

	result, _ = path.String(root)
	return result
}

func getNounInfo(noun string, wikiRegexps *WikiRegexps) []string {

	var result []string

	templateWiki := "https://de.wiktionary.org/wiki"
	queryWiki := templateWiki + "/" + strings.Title(noun)

	wiki := makeWikiRequest(strings.Title(noun))
	if wiki == "" {
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

func getVerbInfo(verb string, wikiRegexps *WikiRegexps) []string {

	var result []string

	templateWiki := "https://de.wiktionary.org/wiki"
	queryWiki := templateWiki + "/" + strings.ToLower(verb)

	wiki := makeWikiRequest(strings.ToLower(verb))
	if wiki == "" {
		return result
	}

	for _, regex := range wikiRegexps.verbInfo {
		match := submatchTextWithRegex(wiki, regex.value)
		text := fmt.Sprintf("%s: %s", regex.name, strings.Join(match[:], ", "))
		result = append(result, text)
	}

	result = append(result, queryWiki)

	return result

}
