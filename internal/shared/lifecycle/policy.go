package lifecycle

type RelationPolicy string

const (
	CascadeArchive        RelationPolicy = "CASCADE_ARCHIVE"
	BlockIfActiveChildren RelationPolicy = "BLOCK_IF_ACTIVE_CHILDREN"
	NoAction              RelationPolicy = "NO_ACTION"
	RelationOnlyDelete    RelationPolicy = "RELATION_ONLY_DELETE"
	AppendOnlyNoDelete    RelationPolicy = "APPEND_ONLY_NO_DELETE"
)

// ParentRef declares a parent FK that the entity depends on. Used by
// RestoreRequiresActiveParent enforcement and by the generic CascadeArchive
// helper to find parent->child relationships.
type ParentRef struct {
	Table  string // parent table
	Column string // FK column on this entity pointing to parent.id
	Label  string // English label for error messages ("project", "field", ...)
}

// CascadeChild declares a child relationship for cascade archive. When the
// parent entity archives, the child rows scoped by `ScopeColumn = parent.id`
// are archived with the same Cause.
type CascadeChild struct {
	Table       string // child table or pivot table
	ScopeColumn string // FK column on child pointing to parent.id
}

type Policy struct {
	ResourceName                string
	TableName                   string
	SupportsCreate              bool
	SupportsRead                bool
	SupportsUpdate              bool
	SupportsArchive             bool
	SupportsRestore             bool
	SupportsHardDelete          bool
	HardDeleteRequiresArchived  bool
	RestoreRequiresActiveParent bool
	ArchivePolicy               RelationPolicy
	RestorePolicy               RelationPolicy
	DeletePolicy                RelationPolicy

	// Parents lists the FK refs to validate when this entity is created,
	// updated, or restored (RestoreRequiresActiveParent). Centralizes what
	// `assertXReferencesActive` calls per repo.
	Parents []ParentRef

	// CascadeTables are pivots / child tables that share the parent's
	// `deleted_at` via the same `Cause`. Used by CascadeArchive/Restore.
	CascadeTables []CascadeChild

	// ChildEntities are entities with their own Policies that should be
	// cascaded recursively when this entity is archived. Each child's own
	// CascadeTables/ChildEntities run within the same Cause.
	ChildEntities []CascadeChild
}

var Policies = map[string]Policy{
	"actors": {
		ResourceName: "actors", TableName: "actors",
		SupportsCreate: true, SupportsRead: true, SupportsUpdate: true,
		SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true,
		HardDeleteRequiresArchived: true,
		ArchivePolicy:              NoAction, RestorePolicy: NoAction, DeletePolicy: BlockIfActiveChildren,
		// Archive of an actor cascades to legacy_actor_map so the mapping
		// follows the actor's lifecycle (see G5 in the audit).
		CascadeTables: []CascadeChild{{Table: "legacy_actor_map", ScopeColumn: "actor_id"}},
	},
	"customers": {
		ResourceName: "customers", TableName: "customers",
		SupportsCreate: true, SupportsRead: true, SupportsUpdate: true,
		SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true,
		HardDeleteRequiresArchived: true,
		ArchivePolicy:              CascadeArchive, RestorePolicy: CascadeArchive, DeletePolicy: BlockIfActiveChildren,
		Parents: []ParentRef{
			{Table: "actors", Column: "actor_id", Label: "actor"},
		},
		ChildEntities: []CascadeChild{{Table: "projects", ScopeColumn: "customer_id"}},
	},
	"projects": {
		ResourceName: "projects", TableName: "projects",
		SupportsCreate: true, SupportsRead: true, SupportsUpdate: true,
		SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true,
		HardDeleteRequiresArchived:  true,
		RestoreRequiresActiveParent: true,
		ArchivePolicy:               CascadeArchive, RestorePolicy: CascadeArchive, DeletePolicy: BlockIfActiveChildren,
		Parents: []ParentRef{
			{Table: "customers", Column: "customer_id", Label: "customer"},
			{Table: "campaigns", Column: "campaign_id", Label: "campaign"},
		},
		CascadeTables: []CascadeChild{
			{Table: "project_managers", ScopeColumn: "project_id"},
			{Table: "project_investors", ScopeColumn: "project_id"},
			{Table: "admin_cost_investors", ScopeColumn: "project_id"},
			{Table: "project_dollar_values", ScopeColumn: "project_id"},
			{Table: "crop_commercializations", ScopeColumn: "project_id"},
		},
		ChildEntities: []CascadeChild{
			{Table: "fields", ScopeColumn: "project_id"},
			{Table: "workorders", ScopeColumn: "project_id"},
			{Table: "work_order_drafts", ScopeColumn: "project_id"},
			{Table: "labors", ScopeColumn: "project_id"},
			{Table: "supplies", ScopeColumn: "project_id"},
			{Table: "supply_movements", ScopeColumn: "project_id"},
			{Table: "stocks", ScopeColumn: "project_id"},
		},
	},
	"fields": {
		ResourceName: "fields", TableName: "fields",
		SupportsCreate: true, SupportsRead: true, SupportsUpdate: true,
		SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true,
		HardDeleteRequiresArchived:  true,
		RestoreRequiresActiveParent: true,
		ArchivePolicy:               CascadeArchive, RestorePolicy: CascadeArchive, DeletePolicy: BlockIfActiveChildren,
		Parents: []ParentRef{
			{Table: "projects", Column: "project_id", Label: "project"},
		},
		CascadeTables: []CascadeChild{{Table: "field_investors", ScopeColumn: "field_id"}},
		ChildEntities: []CascadeChild{{Table: "lots", ScopeColumn: "field_id"}},
	},
	"lots": {
		ResourceName: "lots", TableName: "lots",
		SupportsCreate: true, SupportsRead: true, SupportsUpdate: true,
		SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true,
		HardDeleteRequiresArchived:  true,
		RestoreRequiresActiveParent: true,
		ArchivePolicy:               NoAction, RestorePolicy: NoAction, DeletePolicy: BlockIfActiveChildren,
		Parents: []ParentRef{
			{Table: "fields", Column: "field_id", Label: "field"},
		},
		CascadeTables: []CascadeChild{{Table: "lot_dates", ScopeColumn: "lot_id"}},
	},
	"campaigns": {
		ResourceName: "campaigns", TableName: "campaigns",
		SupportsCreate: true, SupportsRead: true, SupportsUpdate: true,
		SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true,
		HardDeleteRequiresArchived: true,
		ArchivePolicy:              NoAction, RestorePolicy: NoAction, DeletePolicy: BlockIfActiveChildren,
	},
	"workorders": {
		ResourceName: "work-orders", TableName: "workorders",
		SupportsCreate: true, SupportsRead: true, SupportsUpdate: true,
		SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true,
		HardDeleteRequiresArchived:  true,
		RestoreRequiresActiveParent: true,
		ArchivePolicy:               CascadeArchive, RestorePolicy: CascadeArchive, DeletePolicy: BlockIfActiveChildren,
		Parents: []ParentRef{
			{Table: "projects", Column: "project_id", Label: "project"},
			{Table: "fields", Column: "field_id", Label: "field"},
			{Table: "lots", Column: "lot_id", Label: "lot"},
		},
		CascadeTables: []CascadeChild{
			{Table: "workorder_items", ScopeColumn: "workorder_id"},
			{Table: "workorder_investor_splits", ScopeColumn: "workorder_id"},
		},
	},
	"work_order_drafts": {
		ResourceName: "work-order-drafts", TableName: "work_order_drafts",
		SupportsCreate: true, SupportsRead: true, SupportsUpdate: true,
		SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true,
		HardDeleteRequiresArchived:  true,
		RestoreRequiresActiveParent: true,
		ArchivePolicy:               CascadeArchive, RestorePolicy: CascadeArchive, DeletePolicy: BlockIfActiveChildren,
		Parents: []ParentRef{
			{Table: "projects", Column: "project_id", Label: "project"},
		},
		CascadeTables: []CascadeChild{
			{Table: "work_order_draft_items", ScopeColumn: "draft_id"},
			{Table: "work_order_draft_investor_splits", ScopeColumn: "draft_id"},
		},
	},
	"labors": {
		ResourceName: "labors", TableName: "labors",
		SupportsCreate: true, SupportsRead: true, SupportsUpdate: true,
		SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true,
		HardDeleteRequiresArchived:  true,
		RestoreRequiresActiveParent: true,
		ArchivePolicy:               NoAction, RestorePolicy: NoAction, DeletePolicy: BlockIfActiveChildren,
		Parents: []ParentRef{
			{Table: "projects", Column: "project_id", Label: "project"},
		},
	},
	"supplies": {
		ResourceName: "supplies", TableName: "supplies",
		SupportsCreate: true, SupportsRead: true, SupportsUpdate: true,
		SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true,
		HardDeleteRequiresArchived:  true,
		RestoreRequiresActiveParent: true,
		ArchivePolicy:               NoAction, RestorePolicy: NoAction, DeletePolicy: BlockIfActiveChildren,
		Parents: []ParentRef{
			{Table: "projects", Column: "project_id", Label: "project"},
		},
	},
	"supply_movements": {
		ResourceName: "supply-movements", TableName: "supply_movements",
		SupportsCreate: true, SupportsRead: true, SupportsUpdate: true,
		SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true,
		HardDeleteRequiresArchived:  true,
		RestoreRequiresActiveParent: true,
		ArchivePolicy:               NoAction, RestorePolicy: NoAction, DeletePolicy: BlockIfActiveChildren,
		Parents: []ParentRef{
			{Table: "projects", Column: "project_id", Label: "project"},
			{Table: "supplies", Column: "supply_id", Label: "supply"},
		},
	},
	"managers": {
		ResourceName: "managers", TableName: "managers",
		SupportsCreate: true, SupportsRead: true, SupportsUpdate: true,
		SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true,
		HardDeleteRequiresArchived: true,
		ArchivePolicy:              NoAction, RestorePolicy: NoAction, DeletePolicy: BlockIfActiveChildren,
		Parents: []ParentRef{
			{Table: "actors", Column: "actor_id", Label: "actor"},
		},
	},
	"investors": {
		ResourceName: "investors", TableName: "investors",
		SupportsCreate: true, SupportsRead: true, SupportsUpdate: true,
		SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true,
		HardDeleteRequiresArchived: true,
		ArchivePolicy:              NoAction, RestorePolicy: NoAction, DeletePolicy: BlockIfActiveChildren,
		Parents: []ParentRef{
			{Table: "actors", Column: "actor_id", Label: "actor"},
		},
	},
	"providers": {
		ResourceName: "providers", TableName: "providers",
		SupportsCreate: true, SupportsRead: true, SupportsUpdate: true,
		SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true,
		HardDeleteRequiresArchived: true,
		ArchivePolicy:              NoAction, RestorePolicy: NoAction, DeletePolicy: BlockIfActiveChildren,
	},
	"categories": {
		ResourceName: "categories", TableName: "categories",
		SupportsCreate: true, SupportsRead: true, SupportsUpdate: true,
		SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true,
		HardDeleteRequiresArchived: true,
		ArchivePolicy:              NoAction, RestorePolicy: NoAction, DeletePolicy: BlockIfActiveChildren,
	},
	"crops": {
		ResourceName: "crops", TableName: "crops",
		SupportsCreate: true, SupportsRead: true, SupportsUpdate: true,
		SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true,
		HardDeleteRequiresArchived: true,
		ArchivePolicy:              NoAction, RestorePolicy: NoAction, DeletePolicy: BlockIfActiveChildren,
	},
	"class_types": {
		ResourceName: "types", TableName: "types",
		SupportsCreate: true, SupportsRead: true, SupportsUpdate: true,
		SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true,
		HardDeleteRequiresArchived: true,
		ArchivePolicy:              NoAction, RestorePolicy: NoAction, DeletePolicy: BlockIfActiveChildren,
	},
	"lease_types": {
		ResourceName: "lease-types", TableName: "lease_types",
		SupportsCreate: true, SupportsRead: true, SupportsUpdate: true,
		SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true,
		HardDeleteRequiresArchived: true,
		ArchivePolicy:              NoAction, RestorePolicy: NoAction, DeletePolicy: BlockIfActiveChildren,
	},
	"business_parameters": {
		ResourceName: "business-parameters", TableName: "business_parameters",
		SupportsCreate: true, SupportsRead: true, SupportsUpdate: true,
		SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true,
		HardDeleteRequiresArchived: true,
		ArchivePolicy:              NoAction, RestorePolicy: NoAction, DeletePolicy: NoAction,
	},
	"crop_commercializations": {
		ResourceName: "commercializations", TableName: "crop_commercializations",
		SupportsCreate: true, SupportsRead: true, SupportsUpdate: true,
		SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true,
		HardDeleteRequiresArchived:  true,
		RestoreRequiresActiveParent: true,
		ArchivePolicy:               NoAction, RestorePolicy: NoAction, DeletePolicy: BlockIfActiveChildren,
		Parents: []ParentRef{
			{Table: "projects", Column: "project_id", Label: "project"},
			{Table: "crops", Column: "crop_id", Label: "crop"},
		},
	},
}
