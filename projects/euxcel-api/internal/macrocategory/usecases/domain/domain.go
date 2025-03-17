package domain

// MacroCategory represents a macro category (e.g., "Agrochemicals", "Fertilizers", "Seeds", "Operations").
type MacroCategory struct {
	ID   int64  // Primary key (numeric)
	Name string // Name of the macro category
}
