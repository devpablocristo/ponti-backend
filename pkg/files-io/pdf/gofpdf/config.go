package pkgpdf

import "fmt"

// ConfigPort define getters y validación (igual al patrón de pkggorm).
type ConfigPort interface {
	GetFilePath() string
	GetOrientation() string // "P" o "L"
	GetPageSize() string    // "A4" o "Letter"
	GetMarginLeft() float64
	GetMarginTop() float64
	GetMarginRight() float64
	GetFontFamily() string
	GetFontSize() float64
	GetTitle() string
	GetColumnWidths() []float64
	Validate() error
}

// Config concreta.
type Config struct {
	filePath     string
	orientation  string
	pageSize     string
	marginLeft   float64
	marginTop    float64
	marginRight  float64
	fontFamily   string
	fontSize     float64
	title        string
	columnWidths []float64
}

// newConfig crea Config (estilo pkggorm).
func newConfig(filePath, orientation, pageSize, fontFamily, title string,
	fontSize, ml, mt, mr float64, widths []float64) *Config {
	return &Config{
		filePath:     filePath,
		orientation:  orientation,
		pageSize:     pageSize,
		marginLeft:   ml,
		marginTop:    mt,
		marginRight:  mr,
		fontFamily:   fontFamily,
		fontSize:     fontSize,
		title:        title,
		columnWidths: widths,
	}
}

// Getters.
func (c *Config) GetFilePath() string        { return c.filePath }
func (c *Config) GetOrientation() string     { return c.orientation }
func (c *Config) GetPageSize() string        { return c.pageSize }
func (c *Config) GetMarginLeft() float64     { return c.marginLeft }
func (c *Config) GetMarginTop() float64      { return c.marginTop }
func (c *Config) GetMarginRight() float64    { return c.marginRight }
func (c *Config) GetFontFamily() string      { return c.fontFamily }
func (c *Config) GetFontSize() float64       { return c.fontSize }
func (c *Config) GetTitle() string           { return c.title }
func (c *Config) GetColumnWidths() []float64 { return c.columnWidths }

// Validate con defaults (igual estilo a pkggorm).
func (c *Config) Validate() error {
	if c.filePath == "" {
		return fmt.Errorf("file path is required")
	}
	if c.orientation == "" {
		c.orientation = "P"
	}
	if c.orientation != "P" && c.orientation != "L" {
		return fmt.Errorf("invalid orientation: %s", c.orientation)
	}
	if c.pageSize == "" {
		c.pageSize = "A4"
	}
	switch c.pageSize {
	case "A4", "Letter":
	default:
		return fmt.Errorf("invalid page size: %s", c.pageSize)
	}
	if c.marginLeft == 0 {
		c.marginLeft = 10
	}
	if c.marginTop == 0 {
		c.marginTop = 10
	}
	if c.marginRight == 0 {
		c.marginRight = 10
	}
	if c.fontFamily == "" {
		c.fontFamily = "Arial"
	}
	if c.fontSize <= 0 {
		c.fontSize = 12
	}
	return nil
}
