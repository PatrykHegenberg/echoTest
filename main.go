package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/chasefleming/elem-go"
	"github.com/chasefleming/elem-go/attrs"
	"github.com/chasefleming/elem-go/htmx"
	"github.com/jung-kurt/gofpdf"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Todo model
type Bewertung struct {
	Vorname       string
	Nachname      string
	ID            int
	HvPunkte      float64
	HvProzent     float64
	HvNote        int
	LvPunkte      float64
	LvProzent     float64
	LvNote        int
	GesamtProzent float64
	GesamtNote    int
	Gewertet      bool
}

type MaxPunkte struct {
	HvMax        float64
	LvMax        float64
	HvGewichtung float64
	LvGewichtung float64
}

var (
	bewertungen []Bewertung
	maxPunkte   = MaxPunkte{
		HvMax:        0.00,
		HvGewichtung: 0.00,
		LvMax:        0.00,
		LvGewichtung: 0.00,
	}
)

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/", renderBewertungenRoute)
	e.POST("/toggle/:id", toggleWertungRoute)
	e.POST("/add", addBewertungRoute)
	e.GET("/export", exportBewertungenRoute)
	e.GET("/end", endRoute)

	// Start the server
	//e.Logger.Fatal(e.Start(":3000"))
	go func() {
		e.Logger.Fatal(e.Start(":3000"))
	}()

	// Öffne den Standardbrowser mit der Seite localhost:3000
	openInBrowser("http://localhost:3000")

	// Warte auf ein Signal zum Beenden (zum Beispiel STRG+C)
	select {}

}

func renderBewertungenRoute(c echo.Context) error {
	return c.HTML(http.StatusOK, renderBewertungen(bewertungen))
}

func toggleWertungRoute(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var updatedBewertung Bewertung
	for i, bewertung := range bewertungen {
		if bewertung.ID == id {
			bewertungen[i].Gewertet = !bewertung.Gewertet
			updatedBewertung = bewertungen[i]
			break
		}
	}
	return c.HTML(http.StatusOK, createBewertungNode(updatedBewertung).Render())
}

func addBewertungRoute(c echo.Context) error {
	new := parseBewertungen(c)
	if new.Nachname != "" {
		bewertungen = append(bewertungen, new)
	}
	return c.Redirect(http.StatusSeeOther, "/")
}

func parseBewertungen(c echo.Context) Bewertung {
	newName := validateName(c)
	vorname := c.FormValue("vorname")
	if maxPunkte.HvMax == 0.00 {
		hvMax, _ := strconv.ParseFloat(c.FormValue("hv_max"), 64)
		lvMax, _ := strconv.ParseFloat(c.FormValue("lv_max"), 64)
		hvGewichtung, _ := strconv.ParseFloat(c.FormValue("hv_gewichtung"), 64)
		lvGewichtung, _ := strconv.ParseFloat(c.FormValue("lv_gewichtung"), 64)
		maxPunkte.HvMax = hvMax
		maxPunkte.LvMax = lvMax
		maxPunkte.LvGewichtung = lvGewichtung
		maxPunkte.HvGewichtung = hvGewichtung
	}
	hvPunkte, _ := strconv.ParseFloat(c.FormValue("hv_punkte"), 64)
	lvPunkte, _ := strconv.ParseFloat(c.FormValue("lv_punkte"), 64)
	hvProzent := 100.00 / maxPunkte.HvMax * hvPunkte
	lvProzent := 100.00 / maxPunkte.LvMax * lvPunkte
	hvNote := setNote(hvProzent)
	lvNote := setNote(lvProzent)
	gesamtProzent := hvProzent*maxPunkte.HvGewichtung/100 + lvProzent*maxPunkte.LvGewichtung/100
	gesamtNote := setNote(gesamtProzent)

	// Create a new Bewertung struct
	return Bewertung{
		ID:            len(bewertungen) + 1,
		Vorname:       string(vorname),
		Nachname:      string(newName),
		HvPunkte:      hvPunkte,
		HvProzent:     hvProzent,
		HvNote:        int(hvNote),
		LvPunkte:      lvPunkte,
		LvProzent:     lvProzent,
		LvNote:        int(lvNote),
		GesamtProzent: gesamtProzent,
		GesamtNote:    int(gesamtNote),
		Gewertet:      true,
	}
}

func updateGewertetRoute(bewertung Bewertung) elem.Node {
	checkbox := elem.Input(attrs.Props{
		attrs.Type:    "checkbox",
		attrs.Checked: strconv.FormatBool(bewertung.Gewertet),
		htmx.HXPost:   "/toggle/" + strconv.Itoa(bewertung.ID),
		htmx.HXTarget: "#bewertung-" + strconv.Itoa(bewertung.ID),
	})
	return checkbox
}

func createBewertungNode(bewertung Bewertung) elem.Node {
	checkbox := elem.Input(attrs.Props{
		attrs.Type:    "checkbox",
		attrs.Checked: strconv.FormatBool(bewertung.Gewertet),
		htmx.HXPost:   "/toggle/" + strconv.Itoa(bewertung.ID),
		htmx.HXTarget: "#bewertung-" + strconv.Itoa(bewertung.ID),
		htmx.HXSwap:   "outerHTML",
	})

	return elem.Tr(attrs.Props{
		attrs.ID: "bewertung-" + strconv.Itoa(bewertung.ID),
	},
		elem.Td(nil, checkbox),
		elem.Td(nil, elem.Text(bewertung.Vorname)),
		elem.Td(nil, elem.Text(bewertung.Nachname)),
		elem.Td(nil, elem.Text(strconv.FormatFloat(bewertung.HvPunkte, 'f', 2, 64))),
		elem.Td(nil, elem.Text(strconv.FormatFloat(bewertung.HvProzent, 'f', 2, 64))),
		elem.Td(nil, elem.Text(strconv.Itoa(bewertung.HvNote))),
		elem.Td(nil, elem.Text(strconv.FormatFloat(bewertung.LvPunkte, 'f', 2, 64))),
		elem.Td(nil, elem.Text(strconv.FormatFloat(bewertung.LvProzent, 'f', 2, 64))),
		elem.Td(nil, elem.Text(strconv.Itoa(bewertung.LvNote))),
		elem.Td(nil, elem.Text(strconv.FormatFloat(bewertung.GesamtProzent, 'f', 2, 64))),
		elem.Td(nil, elem.Text(strconv.Itoa(bewertung.GesamtNote))),
	)
}

func renderBewertungen(bewertungen []Bewertung) string {
	inputPunkte := elem.Div(nil)
	if maxPunkte.HvGewichtung == 0.00 {
		inputPunkte = elem.Div(attrs.Props{attrs.Class: "tile is-ancestor"},
			elem.Div(attrs.Props{attrs.Class: "tile field is-parent"},
				elem.Input(attrs.Props{
					attrs.Class:       "input is-child",
					attrs.Type:        "text",
					attrs.Name:        "hv_max",
					attrs.Placeholder: "HV-Max-Punkte",
				},
				),
			),
			elem.Div(attrs.Props{attrs.Class: "tile field is-parent"},
				elem.Input(attrs.Props{
					attrs.Class:       "input is-child",
					attrs.Type:        "text",
					attrs.Name:        "hv_gewichtung",
					attrs.Placeholder: "HV-Gewichtung in %",
				},
				),
			),
			elem.Div(attrs.Props{attrs.Class: "tile field is-parent"},
				elem.Input(attrs.Props{
					attrs.Class:       "input is-child",
					attrs.Type:        "text",
					attrs.Name:        "lv_max",
					attrs.Placeholder: "LV-Max-Punkte",
				},
				),
			),
			elem.Div(attrs.Props{attrs.Class: "tile field is-parent"},
				elem.Input(attrs.Props{
					attrs.Class:       "input is-child",
					attrs.Type:        "text",
					attrs.Name:        "lv_gewichtung",
					attrs.Placeholder: "LV-Gewichtung in %",
				},
				),
			),
		)
	} else {
		inputPunkte = elem.Div(attrs.Props{attrs.Class: "tile is-ancestor"},
			elem.Div(attrs.Props{attrs.Class: "tile field is-parent"},
				elem.Input(attrs.Props{
					attrs.Class:       "input is-child",
					attrs.Type:        "text",
					attrs.Name:        "hv_max",
					attrs.Placeholder: "HV-Max-Punkte",
					attrs.Value:       fmt.Sprintf("%.2f", maxPunkte.HvMax),
				},
				),
			),
			elem.Div(attrs.Props{attrs.Class: "tile field is-parent"},
				elem.Input(attrs.Props{
					attrs.Class:       "input is-child",
					attrs.Type:        "text",
					attrs.Name:        "hv_gewichtung",
					attrs.Placeholder: "HV-Gewichtung in %",
					attrs.Value:       fmt.Sprintf("%.2f", maxPunkte.HvGewichtung),
				},
				),
			),
			elem.Div(attrs.Props{attrs.Class: "tile field is-parent"},
				elem.Input(attrs.Props{
					attrs.Class:       "input is-child",
					attrs.Type:        "text",
					attrs.Name:        "lv_max",
					attrs.Placeholder: "LV-Max-Punkte",
					attrs.Value:       fmt.Sprintf("%.2f", maxPunkte.LvMax),
				},
				),
			),
			elem.Div(attrs.Props{attrs.Class: "tile field is-parent"},
				elem.Input(attrs.Props{
					attrs.Class:       "input is-child",
					attrs.Type:        "text",
					attrs.Name:        "lv_gewichtung",
					attrs.Placeholder: "LV-Gewichtung in %",
					attrs.Value:       fmt.Sprintf("%.2f", maxPunkte.LvGewichtung),
				},
				),
			),
		)
	}

	headContent := elem.Head(nil,
		elem.Meta(attrs.Props{attrs.Charset: "UTF-8", attrs.Name: "viewport", attrs.Content: "width=device-width, initial-scale=1.0"}),
		elem.Script(attrs.Props{attrs.Src: "https://unpkg.com/htmx.org"}),
		elem.Link(attrs.Props{attrs.Rel: "stylesheet", attrs.Href: "https://cdn.jsdelivr.net/npm/bulma@0.9.4/css/bulma.min.css"}),
	)

	headerContent := elem.Header(attrs.Props{
		attrs.Class:     "navbar",
		attrs.Role:      "navigation",
		attrs.AriaLabel: "main navigation",
	},
		elem.Div(attrs.Props{
			attrs.ID:    "navbarBasicExample",
			attrs.Class: "navbar-menu",
		},
			elem.Div(attrs.Props{
				attrs.Class: "navbar-start",
			},
				elem.A(attrs.Props{
					attrs.Class: "navbar-item",
				}, elem.Text("Home"),
				),
			),
			elem.Div(attrs.Props{
				attrs.Class: "navbar-end",
			},
				elem.Span(attrs.Props{
					attrs.Class: "navbar-item",
				},
					elem.Button(attrs.Props{
						attrs.Class:    "button is-primary",
						htmx.HXTrigger: "click",
						htmx.HXGet:     "/end",
					}, elem.Text("Beenden"),
					),
				),
			),
		),
	)

	bodyContent := elem.Div(attrs.Props{attrs.Class: "container is-widescreen"},
		elem.Div(attrs.Props{attrs.Class: "card tile is-vertical is-ancestor"},
			elem.Header(attrs.Props{attrs.Class: "card-header"},
				elem.P(attrs.Props{attrs.Class: "card-header-title"}, elem.Text("Englischarbeit"))),
			elem.Div(attrs.Props{attrs.Class: "card-content"},
				elem.Div(attrs.Props{attrs.Class: "content tile is-parent is-vertical gap"},
					elem.H1(attrs.Props{attrs.Class: "tilte"}, elem.Text("Bewertungen")),
					elem.Form(attrs.Props{attrs.Method: "post", attrs.Action: "/add"}, inputPunkte,
						elem.Div(attrs.Props{attrs.Class: "tile is-ancestor"},
							elem.Div(attrs.Props{attrs.Class: "tile field is-parent"},
								elem.Input(attrs.Props{
									attrs.Type:        "text",
									attrs.Name:        "vorname",
									attrs.Class:       "input is-child",
									attrs.Placeholder: "Vorname",
								},
								),
							),
							elem.Div(attrs.Props{attrs.Class: "tile field is-parent"},
								elem.Input(attrs.Props{
									attrs.Type:        "text",
									attrs.Name:        "nachname",
									attrs.Class:       "input is-child",
									attrs.Placeholder: "Nachname",
								},
								),
							),
							elem.Div(attrs.Props{attrs.Class: "tile field is-parent"},
								elem.Input(attrs.Props{
									attrs.Type:        "text",
									attrs.Name:        "hv_punkte",
									attrs.Class:       "input is-child",
									attrs.Placeholder: "HV-Punkte",
								},
								),
							),
							elem.Div(attrs.Props{attrs.Class: "tile field is-parent"},
								elem.Input(attrs.Props{
									attrs.Type:        "text",
									attrs.Name:        "lv_punkte",
									attrs.Class:       "input is-child",
									attrs.Placeholder: "LV-Punkte",
								},
								),
							),
							elem.Div(attrs.Props{attrs.Class: "tile field is-parent"},
								elem.Button(
									attrs.Props{
										attrs.Type:  "submit",
										attrs.Class: "button tile is-child",
									},
									elem.Text("Add"),
								),
							),
						),
					),
					elem.Div(attrs.Props{attrs.Class: "table-container"},
						elem.Table(attrs.Props{attrs.Class: "table is-hoverable"},
							elem.THead(nil,
								elem.Tr(nil,
									elem.Th(nil, elem.Text("Gewertet")),
									elem.Th(nil, elem.Text("Vorname")),
									elem.Th(nil, elem.Text("Nachname")),
									elem.Th(nil, elem.Text("HV-Punkte")),
									elem.Th(nil, elem.Text("HV-Prozent")),
									elem.Th(nil, elem.Text("HV-Note")),
									elem.Th(nil, elem.Text("LV-Punkte")),
									elem.Th(nil, elem.Text("LV-Prozent")),
									elem.Th(nil, elem.Text("LV-Note")),
									elem.Th(nil, elem.Text("Gesamt-Prozent")),
									elem.Th(nil, elem.Text("Gesamt-Note")),
								),
							),
							elem.TBody(nil,
								elem.TransformEach(bewertungen, createBewertungNode)...),
						),
					),
					elem.Div(nil,
						elem.Button(attrs.Props{
							htmx.HXTrigger: "click",
							htmx.HXGet:     "/export",
							attrs.Class:    "button",
						},
							elem.Text("export"),
						),
					),
				),
			),
		),
	)
	footerContent := elem.Footer(attrs.Props{
		attrs.Class: "footer",
	},
		elem.Div(attrs.Props{
			attrs.Class: "content has-text-centered",
		},
			elem.P(nil, elem.Text("&copy; 2023 Alle Rechte vorbehalten.")),
		),
	)

	tbodyContent := elem.TBody(nil)
	htmlContent := elem.Html(nil, elem.Raw("<!DOCTYPE html>"), headContent, headerContent, bodyContent, tbodyContent, footerContent)

	return htmlContent.Render()
}

func setNote(prozent float64) float64 {
	switch {
	case prozent <= 22:
		return 6.00
	case prozent <= 49:
		return 5.00
	case prozent <= 64:
		return 4.00
	case prozent <= 79:
		return 3.00
	case prozent <= 94:
		return 2.00
	default:
		return 1.00
	}
}

func checkGewichtung(lv, hv float64) bool {
	if hv/100+lv/100 > 1 {
		return false
	} else if hv/100+lv/100 < 1 {
		return false
	}
	return true
}

// Die Funktion zum Öffnen des Standardbrowsers
func openInBrowser(url string) {
	var cmd *exec.Cmd

	switch os := runtime.GOOS; os {
	case "darwin":
		// macOS
		cmd = exec.Command("open", url)
	case "windows":
		// Windows
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		// Linux und andere
		cmd = exec.Command("xdg-open", url)
	}

	err := cmd.Start()
	if err != nil {
		fmt.Println("Fehler beim Öffnen des Browsers:", err)
	}
}

func validateName(c echo.Context) string {
	newNachname := c.FormValue("nachname")
	newVorname := c.FormValue("vorname")
	for _, bewertung := range bewertungen {
		fmt.Printf("Name: %v", bewertung.Nachname)
		if bewertung.Nachname == newNachname && bewertung.Vorname == newVorname {
			return ""
		}
	}
	return newNachname
}

func exportBewertungenRoute(c echo.Context) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Add table headers
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(30, 10, "Vorname", "1", 0, "", false, 0, "")
	pdf.CellFormat(30, 10, "Nachname", "1", 0, "", false, 0, "")
	pdf.CellFormat(30, 10, "HV-Punkte", "1", 0, "", false, 0, "")
	pdf.CellFormat(30, 10, "LV-Punkte", "1", 0, "", false, 0, "")
	pdf.CellFormat(30, 10, "Gesamtnote", "1", 0, "", false, 0, "")
	pdf.Ln(-1)

	// Add table rows
	pdf.SetFont("Arial", "", 11)
	for _, bewertung := range bewertungen {
		pdf.CellFormat(30, 10, bewertung.Vorname, "1", 0, "", false, 0, "")
		pdf.CellFormat(30, 10, bewertung.Nachname, "1", 0, "", false, 0, "")
		pdf.CellFormat(30, 10, strconv.FormatFloat(bewertung.HvPunkte, 'f', 2, 64), "1", 0, "", false, 0, "")
		pdf.CellFormat(30, 10, strconv.FormatFloat(bewertung.LvPunkte, 'f', 2, 64), "1", 0, "", false, 0, "")
		pdf.CellFormat(30, 10, strconv.FormatInt(int64(bewertung.GesamtNote), 10), "1", 0, "", false, 0, "")
		pdf.Ln(-1)
	}

	// Save PDF file
	err := pdf.OutputFileAndClose("bewertungen.pdf")
	if err != nil {
		fmt.Println("Fehler beim Exportieren der Bewertungen:", err)
		return err
	}

	c.Response().Header().Set("Content-Disposition", "attachment; filename=bewertung.pdf")
	c.Response().Header().Set("Content-Type", "application/pdf")
	return c.Attachment("bewertung.pdf", "bewertung.pdf")
}

func endRoute(c echo.Context) error {
	os.Exit(0)
	return c.HTML(http.StatusOK, "Tschüss")
}
