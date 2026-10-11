package utils

import (
	"context"
	"dev/internal/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/net/html"
)

const (
	wikipediaAPI         = "https://en.wikipedia.org/w/api.php"
	wikipediaUserAgent   = "dev.matheo-galuba.com/1.0 (https://dev.matheo-galuba.com/contact)"
	transistorCountPage  = "Transistor count"
	transistorCountTitle = "Transistor_count"
)

// Section heading of the Wikipedia page → type of the components listed in its tables
var transistorCountSections = map[string]string{
	"Microprocessors": "CPU",
	"GPUs":            "GPU",
}

var vendorAliases = map[string]string{
	"Nvidia": "NVIDIA",
}

var (
	wikipediaHTTPClient = &http.Client{Timeout: time.Minute}
	leadingNumber       = regexp.MustCompile(`^[~≈]?\s*(\d[\d,]*(?:\.\d+)?)`)
	slugSeparator       = regexp.MustCompile(`[^a-z0-9]+`)
	parenthesized       = regexp.MustCompile(`\([^)]*\)`)
	unitSuffix          = regexp.MustCompile(`^\s*-?\s*(nm|[µμ]m|mm2|mm²|mm|cm2|cm²|[kKMGT]?bits?|[kKMGT]i?[bB]|W|[kMG]?Hz)(?:[^A-Za-z0-9]|$)`)
)

func SeedMicroprocessors(ctx context.Context) {
	count, err := models.CountMicroprocessors(ctx, "")
	if err != nil {
		log.Println("Microprocessors seed: count failed:", err)
		return
	}
	if count > 0 {
		return
	}
	RefreshMicroprocessors(ctx)
}

// The collection is left alone when it holds chips from another origin, so curated data is never mixed or replaced
func RefreshMicroprocessors(ctx context.Context) {
	total, err := models.CountMicroprocessors(ctx, "")
	if err != nil {
		log.Println("Microprocessors refresh: count failed:", err)
		return
	}
	imported, err := models.CountMicroprocessors(ctx, models.MicroprocessorSourceWikipedia)
	if err != nil {
		log.Println("Microprocessors refresh: count failed:", err)
		return
	}
	if total > imported {
		log.Printf("Microprocessors refresh: skipped, the collection holds %d chips that do not come from Wikipedia", total-imported)
		return
	}

	chips, err := fetchWikipediaMicroprocessors(ctx)
	if err == nil && len(chips) == 0 {
		err = fmt.Errorf("no chips found")
	}
	if err == nil {
		err = models.ReplaceMicroprocessors(ctx, models.MicroprocessorSourceWikipedia, chips)
	}
	if err != nil {
		log.Println("Microprocessors refresh: failed:", err)
		return
	}
	log.Printf("Microprocessors refresh: stored %d chips", len(chips))
}

func fetchWikipediaMicroprocessors(ctx context.Context) ([]models.Microprocessor, error) {
	query := url.Values{}
	query.Set("action", "parse")
	query.Set("page", transistorCountPage)
	query.Set("prop", "text|revid")
	query.Set("format", "json")
	query.Set("formatversion", "2")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, wikipediaAPI+"?"+query.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", wikipediaUserAgent)
	resp, err := wikipediaHTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", wikipediaAPI, resp.Status)
	}

	var page struct {
		Parse struct {
			Text  string `json:"text"`
			RevId int    `json:"revid"`
		} `json:"parse"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&page); err != nil {
		return nil, err
	}
	document, err := html.Parse(strings.NewReader(page.Parse.Text))
	if err != nil {
		return nil, err
	}
	sourceURL := fmt.Sprintf("https://en.wikipedia.org/w/index.php?title=%s&oldid=%d", transistorCountTitle, page.Parse.RevId)

	chips := []models.Microprocessor{}
	ids := map[string]bool{}
	for _, table := range findSectionTables(document) {
		rows := expandTable(table.node)
		if len(rows) < 2 {
			continue
		}
		headers := rows[0]
		keys := make([]string, len(headers))
		units := make([]string, len(headers))
		for i, header := range headers {
			keys[i] = detailKey(header)
			units[i] = headerUnit(header)
		}
		column := func(prefixes ...string) int {
			for _, prefix := range prefixes {
				for i, header := range headers {
					rest, found := strings.CutPrefix(strings.ToLower(header), prefix)
					if found && (rest == "" || !unicode.IsLetter([]rune(rest)[0])) {
						return i
					}
				}
			}
			return -1
		}
		nameColumn := column("processor", "fpga", "chip name", "computer", "function", "device name", "node name", "name")
		if nameColumn == -1 {
			continue
		}
		transistorsColumn := column("transistor count", "fgmos transistor count", "number of mosfets")
		yearColumn := column("year", "date of introduction", "production year")
		vendorColumn := column("designer", "manufacturer")
		manufacturerColumn := column("fab", "manufacturer")
		processColumn := column("process", "mos process")
		areaColumn := column("area")
		densityColumn := column("transistor density")
		typeColumn := column("ram type", "flash type", "rom type")
		capacityColumn := column("capacity")

		chipType := table.chipType
		if typeColumn != -1 {
			chipType = strings.Fields(headers[typeColumn])[0]
		}
		cell := func(row []string, index int) string {
			if index == -1 {
				return ""
			}
			return row[index]
		}
		number := func(row []string, index int) float64 {
			value, _ := parseLeadingNumber(cell(row, index))
			return value
		}

		for _, row := range rows[1:] {
			if len(row) != len(headers) || row[nameColumn] == headers[nameColumn] {
				continue
			}
			name := row[nameColumn]
			if name == "" {
				name = strings.TrimSpace(cell(row, capacityColumn) + " " + cell(row, typeColumn))
			}
			if name == "" {
				name = cell(row, processColumn)
			}
			if name == "" {
				continue
			}

			transistors := number(row, transistorsColumn) / 1e6
			if transistorsColumn != -1 && strings.Contains(strings.ToLower(headers[transistorsColumn]), "billion") {
				transistors *= 1e9
			}
			if strings.Contains(cell(row, transistorsColumn), "LUT") {
				transistors = 0
			}
			vendor := cell(row, vendorColumn)
			if alias, found := vendorAliases[vendor]; found {
				vendor = alias
			}

			year := int(number(row, yearColumn))
			var release *primitive.DateTime
			if year > 0 {
				date := primitive.NewDateTimeFromTime(time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC))
				release = &date
			}

			slug := strings.Trim(slugSeparator.ReplaceAllString(strings.ToLower(chipType+" "+name), "-"), "-")
			id := fmt.Sprintf("%s:%s-%d", models.MicroprocessorSourceWikipedia, slug, year)
			for base, n := id, 2; ids[id]; n++ {
				id = fmt.Sprintf("%s-%d", base, n)
			}
			ids[id] = true

			chip := models.Microprocessor{
				Id:           id,
				Name:         name,
				Type:         chipType,
				Release:      release,
				GateSize:     number(row, processColumn),
				DieSize:      number(row, areaColumn),
				Transistors:  transistors,
				Density:      number(row, densityColumn),
				Vendor:       vendor,
				Manufacturer: cell(row, manufacturerColumn),
				Source:       models.MicroprocessorSourceWikipedia,
				SourceURL:    sourceURL,
			}
			for i, value := range row {
				if keys[i] == "" || keys[i] == "ref" || value == "" {
					continue
				}
				if measured, ok := measure(value, units[i]); ok && i != nameColumn && !strings.HasSuffix(keys[i], "Type") {
					chip.SetDetail(keys[i], measured)
				} else {
					chip.SetDetail(keys[i], value)
				}
			}
			chips = append(chips, chip)
		}
	}
	return chips, nil
}

// "Transistor density (tr./mm2)" → "transistorDensity"
func detailKey(header string) string {
	header, _, _ = strings.Cut(header, ",")
	header = parenthesized.ReplaceAllString(header, " ")
	words := strings.FieldsFunc(strings.ToLower(header), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	for i, word := range words {
		if i > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, "")
}

var unitNames = strings.NewReplacer("μ", "µ", "mm2", "mm²", "cm2", "cm²", "tr./", "tr/", "transistors/", "tr/")

// "Process ( nm )" and "Transistor density, tr./mm2" carry the unit of their bare numbers
func headerUnit(header string) string {
	unit := ""
	if match := parenthesized.FindString(header); match != "" {
		unit = strings.Trim(match, "() ")
	} else if _, after, found := strings.Cut(header, ","); found {
		unit = strings.TrimSpace(after)
	}
	switch unit {
	case "nm", "mm2", "bits", "billion", "tr./mm2", "transistors/mm2":
		return unitNames.Replace(unit)
	}
	return ""
}

// measure turns "8,000 nm" into {value: 8000, unit: "nm"}. The original text is kept
// alongside when it says more than that, as in "5 nm (CCD) 6 nm (IOD)".
func measure(text string, defaultUnit string) (map[string]any, bool) {
	number := leadingNumber.FindStringSubmatchIndex(text)
	if number == nil {
		return nil, false
	}
	value, err := strconv.ParseFloat(strings.ReplaceAll(text[number[2]:number[3]], ",", ""), 64)
	if err != nil {
		return nil, false
	}

	unit, rest := defaultUnit, text[number[1]:]
	if match := unitSuffix.FindStringSubmatchIndex(rest); match != nil {
		unit, rest = unitNames.Replace(rest[match[2]:match[3]]), rest[match[3]:]
	}
	if unit == "" {
		return nil, false
	}

	measured := map[string]any{"value": value, "unit": unit}
	if strings.TrimSpace(rest) != "" || number[0] != number[2] {
		measured["text"] = text
	}
	return measured, true
}

func parseLeadingNumber(text string) (float64, bool) {
	match := leadingNumber.FindStringSubmatch(text)
	if match == nil {
		return 0, false
	}
	value, err := strconv.ParseFloat(strings.ReplaceAll(match[1], ",", ""), 64)
	return value, err == nil
}

func htmlAttribute(node *html.Node, name string) string {
	for _, attribute := range node.Attr {
		if attribute.Key == name {
			return attribute.Val
		}
	}
	return ""
}

type sectionTable struct {
	chipType string
	node     *html.Node
}

func findSectionTables(document *html.Node) []sectionTable {
	tables := []sectionTable{}
	section := ""
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			switch node.Data {
			case "h2", "h3", "h4":
				section = htmlAttribute(node, "id")
			case "table":
				if strings.Contains(htmlAttribute(node, "class"), "wikitable") {
					if chipType, known := transistorCountSections[section]; known {
						tables = append(tables, sectionTable{chipType: chipType, node: node})
					}
				}
				return
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(document)
	return tables
}

func rawText(node *html.Node) string {
	if node.Type == html.TextNode {
		return node.Data
	}
	text := ""
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		text += rawText(child)
	}
	return text
}

func cellText(node *html.Node) string {
	var builder strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.TextNode {
			builder.WriteString(node.Data)
			return
		}
		if node.Type == html.ElementNode && node.Data == "style" {
			return
		}
		// Footnote markers are dropped, exponents such as the 2 of mm2 are kept
		if node.Type == html.ElementNode && node.Data == "sup" && strings.Contains(rawText(node), "[") {
			return
		}
		if node.Type == html.ElementNode && node.Data == "br" {
			builder.WriteString(" ")
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	text := strings.Join(strings.Fields(builder.String()), " ")
	if text == "?" || strings.HasPrefix(text, "—") {
		return ""
	}
	return text
}

func expandTable(table *html.Node) [][]string {
	type span struct {
		text string
		rows int
	}
	rows := [][]string{}
	spans := map[int]*span{}

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "tr" {
			row := []string{}
			fillSpans := func() {
				for spans[len(row)] != nil {
					pending := spans[len(row)]
					if pending.rows--; pending.rows == 0 {
						delete(spans, len(row))
					}
					row = append(row, pending.text)
				}
			}
			for cell := node.FirstChild; cell != nil; cell = cell.NextSibling {
				if cell.Type != html.ElementNode || (cell.Data != "td" && cell.Data != "th") {
					continue
				}
				fillSpans()
				text := cellText(cell)
				colspan, _ := strconv.Atoi(htmlAttribute(cell, "colspan"))
				rowspan, _ := strconv.Atoi(htmlAttribute(cell, "rowspan"))
				for i := 0; i < max(colspan, 1); i++ {
					if rowspan > 1 {
						spans[len(row)] = &span{text: text, rows: rowspan - 1}
					}
					row = append(row, text)
				}
			}
			fillSpans()
			rows = append(rows, row)
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(table)
	return rows
}
