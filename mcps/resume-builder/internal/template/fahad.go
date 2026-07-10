package template

import (
	"strings"

	"github.com/LordFarquaadtheCreator/resume-builder/internal/resume"
	"github.com/go-pdf/fpdf"
)

// FahadTemplate renders a resume matching Fahad's LaTeX template layout.
// Letter paper, ~13mm margins, Times serif, darkgray body text.
type FahadTemplate struct{}

const (
	fahadPageWidth    = 215.9 // Letter width in mm
	fahadPageHeight   = 279.4 // Letter height in mm
	fahadMargin       = 13.0
	fahadBottomMargin = 13.0
	fahadContentWidth = fahadPageWidth - 2*fahadMargin // ~189.9mm
)

func (t *FahadTemplate) Name() string { return "fahad" }

func (t *FahadTemplate) Quotas() resume.Quotas {
	return resume.DefaultQuotas
}

func (t *FahadTemplate) PageHeight() float64 {
	return fahadPageHeight
}

func (t *FahadTemplate) BottomMargin() float64 {
	return fahadBottomMargin
}

// darkgray matches LaTeX RGB(38,38,38)
const darkGrayR = 38
const darkGrayG = 38
const darkGrayB = 38

func (t *FahadTemplate) scale(base float64, scale float64) float64 {
	return base * scale
}

func (t *FahadTemplate) Render(pdf *fpdf.Fpdf, data resume.ResumeData, fontScale float64) float64 {
	tr := pdf.UnicodeTranslatorFromDescriptor("")

	pdf.SetAutoPageBreak(false, 0)
	pdf.AddPage()

	leftX := fahadMargin
	rightX := fahadPageWidth - fahadMargin

	// ---------- Header ----------
	pdf.SetY(fahadMargin)
	pdf.SetFont("Times", "B", t.scale(24, fontScale))
	pdf.SetTextColor(0, 0, 0)
	if data.Name != "" {
		pdf.CellFormat(0, 10, tr(data.Name), "", 1, "C", false, 0, "")
	}
	pdf.Ln(1)

	// Contact line: location | email | linkedin | github | website
	pdf.SetFont("Times", "", t.scale(9, fontScale))
	var parts []string
	if data.Contact.Location != "" {
		parts = append(parts, tr(data.Contact.Location))
	}
	if data.Contact.Email != "" {
		parts = append(parts, tr(data.Contact.Email))
	}
	if link, ok := data.Contact.Links["linkedin"]; ok && link != "" {
		parts = append(parts, tr(link))
	}
	if link, ok := data.Contact.Links["github"]; ok && link != "" {
		parts = append(parts, tr(link))
	}
	if link, ok := data.Contact.Links["website"]; ok && link != "" {
		parts = append(parts, tr(link))
	}
	if len(parts) > 0 {
		contactLine := strings.Join(parts, " | ")
		pdf.CellFormat(0, 5, contactLine, "", 1, "C", false, 0, "")
	}
	pdf.Ln(2)

	// ---------- Sections ----------
	// Order: Education → Skills → Experience → Projects
	if len(data.Education) > 0 {
		t.renderSectionHeader(pdf, "Education", fontScale, tr)
		for _, edu := range data.Education {
			t.renderEducationEntry(pdf, edu, fontScale, tr, leftX, rightX)
		}
		pdf.Ln(2)
	}

	if len(data.Skills) > 0 {
		t.renderSectionHeader(pdf, "Technical Skills", fontScale, tr)
		for _, skill := range data.Skills {
			t.renderSkillEntry(pdf, skill, fontScale, tr, leftX, rightX)
		}
		pdf.Ln(2)
	}

	if len(data.Experiences) > 0 {
		t.renderSectionHeader(pdf, "Professional Experience", fontScale, tr)
		for _, exp := range data.Experiences {
			t.renderExperienceEntry(pdf, exp, fontScale, tr, leftX, rightX)
		}
		pdf.Ln(2)
	}

	if len(data.Projects) > 0 {
		t.renderSectionHeader(pdf, "Projects / Open Source", fontScale, tr)
		for _, proj := range data.Projects {
			t.renderProjectEntry(pdf, proj, fontScale, tr, leftX, rightX)
		}
	}

	return pdf.GetY()
}

func (t *FahadTemplate) renderSectionHeader(pdf *fpdf.Fpdf, title string, fontScale float64, tr func(string) string) {
	pdf.SetFont("Times", "B", t.scale(13, fontScale))
	pdf.SetTextColor(0, 0, 0)
	pdf.SetX(fahadMargin)
	pdf.CellFormat(0, 6, tr(title), "", 1, "L", false, 0, "")
	// horizontal rule
	y := pdf.GetY()
	pdf.SetDrawColor(0, 0, 0)
	pdf.Line(fahadMargin, y, fahadPageWidth-fahadMargin, y)
	pdf.Ln(2)
}

func (t *FahadTemplate) renderEducationEntry(pdf *fpdf.Fpdf, edu resume.Education, fontScale float64, tr func(string) string, leftX, rightX float64) {
	pdf.SetFont("Times", "B", t.scale(10, fontScale))
	pdf.SetTextColor(0, 0, 0)
	pdf.SetX(leftX)

	// Left: Institution | Degree (italic)
	leftText := ""
	if edu.Institution != "" {
		leftText = tr(edu.Institution)
	}
	if edu.Degree != "" {
		leftText += " | " + tr(edu.Degree)
	}

	// Right: Location | EndDate
	rightText := ""
	if edu.Location != "" {
		rightText = tr(edu.Location)
	}
	if edu.End != "" {
		if rightText != "" {
			rightText += " | "
		}
		rightText += tr(edu.End)
	}

	t.renderTwoColumnRow(pdf, leftText, rightText, fontScale, leftX, rightX)
	pdf.Ln(1)
}

func (t *FahadTemplate) renderSkillEntry(pdf *fpdf.Fpdf, skill resume.SkillGroup, fontScale float64, tr func(string) string, leftX, rightX float64) {
	pdf.SetX(leftX)
	pdf.SetFont("Times", "B", t.scale(10, fontScale))
	pdf.SetTextColor(0, 0, 0)

	categoryText := ""
	if skill.Category != "" {
		categoryText = tr(skill.Category) + ": "
	}

	valuesText := ""
	if skill.Values != "" {
		valuesText = tr(skill.Values)
	}

	// Bold category + darkgray values
	pdf.SetFont("Times", "B", t.scale(10, fontScale))
	pdf.SetTextColor(0, 0, 0)
	catW := pdf.GetStringWidth(categoryText)
	pdf.CellFormat(catW, 5, categoryText, "", 0, "L", false, 0, "")

	pdf.SetFont("Times", "", t.scale(10, fontScale))
	pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
	pdf.CellFormat(0, 5, valuesText, "", 1, "L", false, 0, "")
	pdf.Ln(0.5)
}

func (t *FahadTemplate) renderExperienceEntry(pdf *fpdf.Fpdf, exp resume.Experience, fontScale float64, tr func(string) string, leftX, rightX float64) {
	pdf.SetX(leftX)
	pdf.SetFont("Times", "B", t.scale(10, fontScale))
	pdf.SetTextColor(0, 0, 0)

	// Left: Company | Role (italic)
	leftText := ""
	if exp.Company != "" {
		leftText = tr(exp.Company)
	}
	if exp.Role != "" {
		leftText += " | " + tr(exp.Role)
	}

	// Right: Location | EndDate -- StartDate
	rightText := ""
	if exp.Location != "" {
		rightText = tr(exp.Location)
	}
	dateRange := ""
	if exp.End != "" {
		dateRange = tr(exp.End)
	}
	if exp.Start != "" && exp.Start != exp.End {
		if dateRange != "" {
			dateRange += " " + tr("\u2014") + " "
		}
		dateRange += tr(exp.Start)
	}
	if dateRange != "" {
		if rightText != "" {
			rightText += " | "
		}
		rightText += dateRange
	}

	t.renderTwoColumnRow(pdf, leftText, rightText, fontScale, leftX, rightX)
	pdf.Ln(0.5)

	// Bullets
	for _, bullet := range exp.Bullets {
		if bullet == "" {
			continue
		}
		t.renderBullet(pdf, bullet, fontScale, tr, leftX)
	}
	pdf.Ln(1)
}

func (t *FahadTemplate) renderProjectEntry(pdf *fpdf.Fpdf, proj resume.Project, fontScale float64, tr func(string) string, leftX, rightX float64) {
	pdf.SetX(leftX)
	pdf.SetFont("Times", "B", t.scale(10, fontScale))
	pdf.SetTextColor(0, 0, 0)

	// Left: Name | Tech (italic)
	leftText := ""
	if proj.Name != "" {
		leftText = tr(proj.Name)
	}
	if proj.Tech != "" {
		leftText += " | " + tr(proj.Tech)
	}

	// Right: Date
	rightText := ""
	if proj.Date != "" {
		rightText = tr(proj.Date)
	}

	t.renderTwoColumnRow(pdf, leftText, rightText, fontScale, leftX, rightX)
	pdf.Ln(0.5)

	// Bullets
	for _, bullet := range proj.Bullets {
		if bullet == "" {
			continue
		}
		t.renderBullet(pdf, bullet, fontScale, tr, leftX)
	}
	pdf.Ln(1)
}

func (t *FahadTemplate) renderBullet(pdf *fpdf.Fpdf, text string, fontScale float64, tr func(string) string, leftX float64) {
	pdf.SetFont("Times", "", t.scale(9.5, fontScale))
	pdf.SetTextColor(darkGrayR, darkGrayG, darkGrayB)
	pdf.SetX(leftX + 3)
	// small bullet marker
	pdf.CellFormat(3, 4, tr("\u2022"), "", 0, "L", false, 0, "")
	// text wraps within remaining width
	textW := fahadContentWidth - 6
	pdf.MultiCell(textW, 4.5, tr(text), "", "L", false)
	pdf.Ln(0.2)
}

// renderTwoColumnRow draws left text (bold) and right text (italic) on one line
// using a two-column layout. Left is left-aligned, right is right-aligned.
func (t *FahadTemplate) renderTwoColumnRow(pdf *fpdf.Fpdf, leftText, rightText string, fontScale float64, leftX, rightX float64) {
	if leftText == "" && rightText == "" {
		return
	}

	// Calculate widths
	pdf.SetFont("Times", "B", t.scale(10, fontScale))
	leftW := 0.0
	if leftText != "" {
		leftW = pdf.GetStringWidth(leftText)
	}

	pdf.SetFont("Times", "I", t.scale(9, fontScale))
	rightW := 0.0
	if rightText != "" {
		rightW = pdf.GetStringWidth(rightText)
	}

	// gap between columns
	gap := 10.0
	availW := fahadContentWidth

	// If both fit, render side by side
	if leftW+rightW+gap <= availW {
		y := pdf.GetY()
		// Left
		pdf.SetFont("Times", "B", t.scale(10, fontScale))
		pdf.SetXY(leftX, y)
		if leftText != "" {
			pdf.CellFormat(leftW, 5, leftText, "", 0, "L", false, 0, "")
		}
		// Right
		pdf.SetFont("Times", "I", t.scale(9, fontScale))
		pdf.SetXY(rightX-rightW, y)
		if rightText != "" {
			pdf.CellFormat(rightW, 5, rightText, "", 1, "L", false, 0, "")
		}
	} else {
		// Stack: left on first line, right on second
		pdf.SetFont("Times", "B", t.scale(10, fontScale))
		pdf.SetX(leftX)
		if leftText != "" {
			pdf.CellFormat(0, 5, leftText, "", 1, "L", false, 0, "")
		}
		pdf.SetFont("Times", "I", t.scale(9, fontScale))
		pdf.SetX(leftX)
		if rightText != "" {
			pdf.CellFormat(0, 4, rightText, "", 1, "R", false, 0, "")
		}
	}
}
