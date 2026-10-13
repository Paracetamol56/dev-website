package chip

import (
	"context"
	"dev/internal/db"
	"dev/internal/testutil"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/net/html"
)

const transistorCountFixture = `
<div>
<h2 id="History">History</h2>
<table class="wikitable"><tr><th>Processor</th></tr><tr><td>Unknown section</td></tr></table>

<h2 id="Microprocessors">Microprocessors</h2>
<table class="infobox"><tr><th>Processor</th></tr><tr><td>Not a wikitable</td></tr></table>
<table class="wikitable sortable"><tbody>
<tr><th>Processor</th><th>Transistor count</th><th>Year</th><th>Designer</th><th>Process<br>(nm)</th><th>Area (mm<sup>2</sup>)</th><th>Transistor density, tr./mm2</th><th>Ref</th></tr>
<tr><td>Intel 4004<sup>[1]</sup></td><td>2,250</td><td>1971</td><td rowspan="2">Intel</td><td>10,000 nm</td><td>12 mm<sup>2</sup></td><td>188</td><td>[1]</td></tr>
<tr><td>Intel 8008</td><td>3,500</td><td>1972</td><td>10,000</td><td>14</td><td>~250</td><td></td></tr>
<tr><th>Processor</th><th>Transistor count</th><th>Year</th><th>Designer</th><th>Process (nm)</th><th>Area (mm2)</th><th>Transistor density, tr./mm2</th><th>Ref</th></tr>
<tr><td>Apple M1</td><td>~16,000,000,000</td><td>2020</td><td>Apple</td><td>5 nm (N5)</td><td>119</td><td>?</td><td>—</td></tr>
<tr><td>Apple M1</td><td>16,000,000,000</td><td>2020</td><td>Apple</td><td>5</td><td>— unknown</td><td></td><td></td></tr>
<tr><td>Too short</td><td>1</td></tr>
<tr><td><style>.mw{color:red}</style>Mystery</td><td>unknown</td><td>n/a</td><td>Nvidia</td><td></td><td></td><td></td><td></td></tr>
</tbody></table>

<h3 id="GPUs">GPUs</h3>
<table class="wikitable"><tr><th>Only a header</th></tr></table>
<table class="wikitable"><tr><th>Unrelated</th><th>Column</th></tr><tr><td>a</td><td>b</td></tr></table>
<table class="wikitable">
<tr><th>Processor</th><th>Transistor count</th><th>Year</th><th>Designer(s)</th><th>Fab(s)</th><th>Process</th><th>Area</th><th>Transistor density (tr./mm2)</th></tr>
<tr><td>GeForce 256</td><td>17,000,000</td><td>1999</td><td>Nvidia</td><td>TSMC</td><td>220 nm</td><td>139 mm²</td><td>122,000</td></tr>
<tr><td>Programmable</td><td>50,000 LUTs</td><td>2001</td><td>Xilinx</td><td>UMC</td><td>150 nm</td><td></td><td></td></tr>
</table>
<table class="wikitable">
<tr><th>Chip name</th><th>Capacity (bits)</th><th>RAM type</th><th>Transistor count (billion)</th><th>Date of introduction</th><th>Manufacturer</th><th>MOS process</th><th>Area</th></tr>
<tr><td>HM5116</td><td>16 Mbit</td><td>DRAM</td><td>0.016</td><td>1990</td><td>Hitachi</td><td>800 nm</td><td>70 mm2</td></tr>
<tr><td></td><td>64 kbit</td><td>SRAM</td><td></td><td>1980</td><td></td><td>3 μm</td><td></td></tr>
<tr><td></td><td></td><td></td><td></td><td>2005</td><td></td><td>65 nm</td><td></td></tr>
<tr><td></td><td></td><td></td><td></td><td></td><td></td><td></td><td></td></tr>
</table>

<h2 id="Memory">Memory</h2>
<table class="wikitable"><tr><th>Chip name</th></tr><tr><td>Section not imported</td></tr></table>
</div>`

func wikipediaResponse(t testing.TB, text string) string {
	t.Helper()
	body, err := json.Marshal(map[string]any{"parse": map[string]any{"text": text, "revid": 1234}})
	require.NoError(t, err)
	return string(body)
}

func parseHTML(t *testing.T, fragment string) *html.Node {
	t.Helper()
	document, err := html.Parse(strings.NewReader(fragment))
	require.NoError(t, err)
	return document
}

func year(value int) *primitive.DateTime {
	date := primitive.NewDateTimeFromTime(time.Date(value, time.January, 1, 0, 0, 0, 0, time.UTC))
	return &date
}

func TestFetchWikipediaMicroprocessors(t *testing.T) {
	internet := testutil.FakeInternet(t)
	internet.JSON("en.wikipedia.org", http.StatusOK, wikipediaResponse(t, transistorCountFixture))

	chips, err := fetchWikipediaMicroprocessors(context.Background())
	require.NoError(t, err)

	requests := internet.Requests("en.wikipedia.org")
	require.Len(t, requests, 1)
	assert.Equal(t, "/w/api.php", requests[0].Path)
	assert.Equal(t, "action=parse&format=json&formatversion=2&page=Transistor+count&prop=text%7Crevid", requests[0].Query)
	assert.Equal(t, wikipediaUserAgent, requests[0].Header.Get("User-Agent"))

	byId := map[string]Microprocessor{}
	ids := []string{}
	for _, chip := range chips {
		byId[chip.Id] = chip
		ids = append(ids, chip.Id)
		assert.Equal(t, "wikipedia", chip.Source)
		assert.Equal(t, "https://en.wikipedia.org/w/index.php?title=Transistor_count&oldid=1234", chip.SourceURL)
	}
	assert.Equal(t, []string{
		"wikipedia:cpu-intel-4004-1971",
		"wikipedia:cpu-intel-8008-1972",
		"wikipedia:cpu-apple-m1-2020",
		"wikipedia:cpu-apple-m1-2020-2",
		"wikipedia:cpu-mystery-0",
		"wikipedia:gpu-geforce-256-1999",
		"wikipedia:gpu-programmable-2001",
		"wikipedia:ram-hm5116-1990",
		"wikipedia:ram-64-kbit-sram-1980",
		"wikipedia:ram-65-nm-2005",
	}, ids)

	assert.Equal(t, Microprocessor{
		Id:          "wikipedia:cpu-intel-4004-1971",
		Name:        "Intel 4004",
		Type:        "CPU",
		Release:     year(1971),
		GateSize:    10000,
		DieSize:     12,
		Transistors: 0.00225,
		Density:     188,
		Vendor:      "Intel",
		Source:      "wikipedia",
		SourceURL:   "https://en.wikipedia.org/w/index.php?title=Transistor_count&oldid=1234",
		Details: map[string]any{
			"processor":         "Intel 4004",
			"transistorCount":   "2,250",
			"year":              "1971",
			"designer":          "Intel",
			"process":           map[string]any{"value": 10000.0, "unit": "nm"},
			"area":              map[string]any{"value": 12.0, "unit": "mm²"},
			"transistorDensity": map[string]any{"value": 188.0, "unit": "tr/mm²"},
		},
	}, byId["wikipedia:cpu-intel-4004-1971"])

	intel8008 := byId["wikipedia:cpu-intel-8008-1972"]
	assert.Equal(t, "Intel", intel8008.Vendor, "rowspan")
	assert.Equal(t, 250.0, intel8008.Density)
	assert.Equal(t, map[string]any{"value": 10000.0, "unit": "nm"}, intel8008.Details["process"], "unit of the header")
	assert.Equal(t, map[string]any{"value": 250.0, "unit": "tr/mm²", "text": "~250"}, intel8008.Details["transistorDensity"])

	m1 := byId["wikipedia:cpu-apple-m1-2020"]
	assert.Equal(t, 16000.0, m1.Transistors)
	assert.Equal(t, 5.0, m1.GateSize)
	assert.Equal(t, map[string]any{"value": 5.0, "unit": "nm", "text": "5 nm (N5)"}, m1.Details["process"])
	assert.Zero(t, m1.Density, "? is not a value")
	assert.NotContains(t, m1.Details, "transistorDensity")
	assert.NotContains(t, m1.Details, "ref")
	assert.Zero(t, byId["wikipedia:cpu-apple-m1-2020-2"].DieSize, "— is not a value")

	mystery := byId["wikipedia:cpu-mystery-0"]
	assert.Equal(t, "Mystery", mystery.Name, "style elements are dropped")
	assert.Nil(t, mystery.Release)
	assert.Zero(t, mystery.Transistors)
	assert.Equal(t, "NVIDIA", mystery.Vendor)
	assert.Equal(t, "Nvidia", mystery.Details["designer"])

	geforce := byId["wikipedia:gpu-geforce-256-1999"]
	assert.Equal(t, "GPU", geforce.Type)
	assert.Equal(t, "NVIDIA", geforce.Vendor)
	assert.Equal(t, "TSMC", geforce.Manufacturer)
	assert.Equal(t, 17.0, geforce.Transistors)
	assert.Equal(t, 220.0, geforce.GateSize)
	assert.Equal(t, 139.0, geforce.DieSize)
	assert.Equal(t, 122000.0, geforce.Density)
	assert.Equal(t, "TSMC", geforce.Details["fab"])
	assert.Zero(t, byId["wikipedia:gpu-programmable-2001"].Transistors, "LUTs are not transistors")

	ram := byId["wikipedia:ram-hm5116-1990"]
	assert.Equal(t, "RAM", ram.Type, "type of the table's type column")
	assert.Equal(t, 16.0, ram.Transistors, "counted in billions")
	assert.Equal(t, "Hitachi", ram.Vendor)
	assert.Equal(t, "Hitachi", ram.Manufacturer)
	assert.Equal(t, 800.0, ram.GateSize)
	assert.Equal(t, map[string]any{
		"chipName":           "HM5116",
		"capacity":           map[string]any{"value": 16.0, "unit": "Mbit"},
		"ramType":            "DRAM",
		"transistorCount":    map[string]any{"value": 0.016, "unit": "billion"},
		"dateOfIntroduction": "1990",
		"mosProcess":         map[string]any{"value": 800.0, "unit": "nm"},
		"area":               map[string]any{"value": 70.0, "unit": "mm²"},
	}, ram.Details)

	assert.Equal(t, "64 kbit SRAM", byId["wikipedia:ram-64-kbit-sram-1980"].Name, "named after capacity and type")
	assert.Equal(t, map[string]any{"value": 3.0, "unit": "µm"}, byId["wikipedia:ram-64-kbit-sram-1980"].Details["mosProcess"])
	assert.Equal(t, "65 nm", byId["wikipedia:ram-65-nm-2005"].Name, "named after the process")
}

func TestFetchWikipediaMicroprocessorsFailures(t *testing.T) {
	ctx := context.Background()
	internet := testutil.FakeInternet(t)

	_, err := fetchWikipediaMicroprocessors(ctx)
	assert.ErrorContains(t, err, "unexpected request")

	internet.JSON("en.wikipedia.org", http.StatusServiceUnavailable, ``)
	_, err = fetchWikipediaMicroprocessors(ctx)
	assert.ErrorContains(t, err, "503")

	internet.JSON("en.wikipedia.org", http.StatusOK, `<html>`)
	_, err = fetchWikipediaMicroprocessors(ctx)
	assert.Error(t, err)

	_, err = fetchWikipediaMicroprocessors(nil)
	assert.Error(t, err, "nil context")
}

func storedChips(t *testing.T) []Microprocessor {
	t.Helper()
	chips := []Microprocessor{}
	cursor, err := db.GetDB().Collection("microprocessors").Find(context.Background(), bson.M{})
	require.NoError(t, err)
	require.NoError(t, cursor.All(context.Background(), &chips))
	return chips
}

func TestRefreshMicroprocessors(t *testing.T) {
	testutil.Clean(t)
	ctx := context.Background()
	internet := testutil.FakeInternet(t)
	internet.JSON("en.wikipedia.org", http.StatusOK, wikipediaResponse(t, transistorCountFixture))

	RefreshMicroprocessors(ctx)
	assert.Len(t, storedChips(t), 10)

	internet.JSON("en.wikipedia.org", http.StatusOK, wikipediaResponse(t, `
		<h2 id="Microprocessors"></h2>
		<table class="wikitable"><tr><th>Processor</th><th>Year</th></tr><tr><td>Z80</td><td>1976</td></tr></table>`))
	RefreshMicroprocessors(ctx)
	chips := storedChips(t)
	require.Len(t, chips, 1, "chips that left the page are removed")
	assert.Equal(t, "wikipedia:cpu-z80-1976", chips[0].Id)
	assert.Equal(t, "Z80", chips[0].Details["processor"])

	internet.JSON("en.wikipedia.org", http.StatusOK, wikipediaResponse(t, `<p>Page blanked</p>`))
	RefreshMicroprocessors(ctx)
	assert.Len(t, storedChips(t), 1, "an empty page does not empty the collection")

	internet.JSON("en.wikipedia.org", http.StatusInternalServerError, ``)
	RefreshMicroprocessors(ctx)
	assert.Len(t, storedChips(t), 1)
}

func TestRefreshMicroprocessorsLeavesCuratedDataAlone(t *testing.T) {
	testutil.Clean(t)
	ctx := context.Background()
	internet := testutil.FakeInternet(t)
	internet.JSON("en.wikipedia.org", http.StatusOK, wikipediaResponse(t, transistorCountFixture))
	require.NoError(t, ReplaceMicroprocessors(ctx, "curated", []Microprocessor{{Id: "curated:1", Name: "Hand written", Source: "curated"}}))

	RefreshMicroprocessors(ctx)
	SeedMicroprocessors(ctx)

	assert.Len(t, storedChips(t), 1)
	assert.Empty(t, internet.Requests("en.wikipedia.org"))
}

func TestSeedMicroprocessors(t *testing.T) {
	testutil.Clean(t)
	ctx := context.Background()
	internet := testutil.FakeInternet(t)
	internet.JSON("en.wikipedia.org", http.StatusOK, wikipediaResponse(t, transistorCountFixture))

	SeedMicroprocessors(ctx)
	assert.Len(t, storedChips(t), 10)
	SeedMicroprocessors(ctx)
	assert.Len(t, internet.Requests("en.wikipedia.org"), 1, "a filled collection is not seeded again")

	testutil.BreakDB(t)
	SeedMicroprocessors(ctx)
	RefreshMicroprocessors(ctx)
	assert.Len(t, internet.Requests("en.wikipedia.org"), 1)
}

func TestDetailKey(t *testing.T) {
	assert.Equal(t, "transistorDensity", detailKey("Transistor density (tr./mm2)"))
	assert.Equal(t, "transistorDensity", detailKey("Transistor density, tr./mm2"))
	assert.Equal(t, "process", detailKey("Process ( nm )"))
	assert.Equal(t, "dateOfIntroduction", detailKey("Date of introduction"))
	assert.Equal(t, "ref", detailKey("Ref."))
	assert.Equal(t, "", detailKey("(?)"))
}

func TestHeaderUnit(t *testing.T) {
	assert.Equal(t, "nm", headerUnit("Process ( nm )"))
	assert.Equal(t, "mm²", headerUnit("Area (mm2)"))
	assert.Equal(t, "tr/mm²", headerUnit("Transistor density, tr./mm2"))
	assert.Equal(t, "tr/mm²", headerUnit("Transistor density (transistors/mm2)"))
	assert.Equal(t, "bits", headerUnit("Capacity (bits)"))
	assert.Equal(t, "billion", headerUnit("Transistor count (billion)"))
	assert.Equal(t, "", headerUnit("Designer(s)"))
	assert.Equal(t, "", headerUnit("Year"))
	assert.Equal(t, "", headerUnit("Name, also known as"))
}

func TestMeasure(t *testing.T) {
	cases := []struct {
		text        string
		defaultUnit string
		expected    map[string]any
	}{
		{"8,000 nm", "", map[string]any{"value": 8000.0, "unit": "nm"}},
		{"8000", "nm", map[string]any{"value": 8000.0, "unit": "nm"}},
		{"1.5 μm", "nm", map[string]any{"value": 1.5, "unit": "µm"}},
		{"12 mm2", "", map[string]any{"value": 12.0, "unit": "mm²"}},
		{"3 cm²", "", map[string]any{"value": 3.0, "unit": "cm²"}},
		{"5 nm (CCD) 6 nm (IOD)", "", map[string]any{"value": 5.0, "unit": "nm", "text": "5 nm (CCD) 6 nm (IOD)"}},
		{"~250", "tr/mm²", map[string]any{"value": 250.0, "unit": "tr/mm²", "text": "~250"}},
		{"64-bit", "", map[string]any{"value": 64.0, "unit": "bit"}},
		{"4 GHz", "", map[string]any{"value": 4.0, "unit": "GHz"}},
		{"95 W", "", map[string]any{"value": 95.0, "unit": "W"}},
		{"16 GiB", "", map[string]any{"value": 16.0, "unit": "GiB"}},
		{"2 nmos", "nm", map[string]any{"value": 2.0, "unit": "nm", "text": "2 nmos"}},
		{"1971", "", nil},
		{"Intel 4004", "nm", nil},
		{"", "nm", nil},
	}
	for _, c := range cases {
		measured, ok := measure(c.text, c.defaultUnit)
		assert.Equal(t, c.expected != nil, ok, c.text)
		assert.Equal(t, c.expected, measured, c.text)
	}
}

func TestParseLeadingNumber(t *testing.T) {
	for text, expected := range map[string]float64{
		"2,250":           2250,
		"~16,000,000,000": 16e9,
		"≈ 3.5 billion":   3.5,
		"1971 (est.)":     1971,
		"0.016":           0.016,
	} {
		value, ok := parseLeadingNumber(text)
		assert.True(t, ok, text)
		assert.Equal(t, expected, value, text)
	}
	for _, text := range []string{"", "unknown", "-5", ".5", "about 3"} {
		_, ok := parseLeadingNumber(text)
		assert.False(t, ok, text)
	}
}

func TestExpandTable(t *testing.T) {
	document := parseHTML(t, `<table>
		<tr><th colspan="2">Wide</th><th rowspan="3">Tall</th><th>Last</th></tr>
		<tr><td rowspan="2">a</td><td>b</td><td>c</td></tr>
		<tr><td>d</td><td>e<br>f</td></tr>
		<tr><td>g</td>text<td>h<sup>2</sup></td><td>i<sup>[note 1]</sup></td><td rowspan="2">j</td></tr>
		<tr></tr>
	</table>`)

	assert.Equal(t, [][]string{
		{"Wide", "Wide", "Tall", "Last"},
		{"a", "b", "Tall", "c"},
		{"a", "d", "Tall", "e f"},
		{"g", "h2", "i", "j"},
		{},
	}, expandTable(document))
}

func TestFindSectionTables(t *testing.T) {
	document := parseHTML(t, `
		<table class="wikitable" id="before"></table>
		<h2 id="Microprocessors"></h2>
		<div><table class="wikitable" id="cpu"><tr><td><table class="wikitable" id="nested"></table></td></tr></table></div>
		<table id="plain"></table>
		<h4 id="GPUs"></h4>
		<table class="sortable wikitable" id="gpu"></table>
		<h3 id="Other"></h3>
		<table class="wikitable" id="other"></table>`)

	found := map[string]string{}
	for _, table := range findSectionTables(document) {
		found[htmlAttribute(table.node, "id")] = table.chipType
	}
	assert.Equal(t, map[string]string{"cpu": "CPU", "gpu": "GPU"}, found)
	assert.Equal(t, "", htmlAttribute(document, "id"))
}

func TestScheduleMicroprocessorRefreshSeedsAtStartup(t *testing.T) {
	testutil.Clean(t)
	testutil.FakeInternet(t).JSON("en.wikipedia.org", http.StatusOK, wikipediaResponse(t, transistorCountFixture))

	ScheduleMicroprocessorRefresh()

	assert.Eventually(t, func() bool {
		stored, _ := CountMicroprocessors(context.Background(), "")
		return stored == 10
	}, 5*time.Second, 20*time.Millisecond)
}
