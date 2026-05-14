package lifecycle

type RelationPolicy string

const (
	CascadeArchive        RelationPolicy = "CASCADE_ARCHIVE"
	BlockIfActiveChildren RelationPolicy = "BLOCK_IF_ACTIVE_CHILDREN"
	NoAction              RelationPolicy = "NO_ACTION"
	RelationOnlyDelete    RelationPolicy = "RELATION_ONLY_DELETE"
	AppendOnlyNoDelete    RelationPolicy = "APPEND_ONLY_NO_DELETE"
)

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
}

var Policies = map[string]Policy{
	"actors":              {ResourceName: "actors", TableName: "actors", SupportsCreate: true, SupportsRead: true, SupportsUpdate: true, SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true, HardDeleteRequiresArchived: true, ArchivePolicy: NoAction, RestorePolicy: NoAction, DeletePolicy: BlockIfActiveChildren},
	"customers":           {ResourceName: "customers", TableName: "customers", SupportsCreate: true, SupportsRead: true, SupportsUpdate: true, SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true, HardDeleteRequiresArchived: true, ArchivePolicy: CascadeArchive, RestorePolicy: CascadeArchive, DeletePolicy: BlockIfActiveChildren},
	"projects":            {ResourceName: "projects", TableName: "projects", SupportsCreate: true, SupportsRead: true, SupportsUpdate: true, SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true, HardDeleteRequiresArchived: true, RestoreRequiresActiveParent: true, ArchivePolicy: CascadeArchive, RestorePolicy: CascadeArchive, DeletePolicy: BlockIfActiveChildren},
	"fields":              {ResourceName: "fields", TableName: "fields", SupportsCreate: true, SupportsRead: true, SupportsUpdate: true, SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true, HardDeleteRequiresArchived: true, RestoreRequiresActiveParent: true, ArchivePolicy: CascadeArchive, RestorePolicy: CascadeArchive, DeletePolicy: BlockIfActiveChildren},
	"lots":                {ResourceName: "lots", TableName: "lots", SupportsCreate: true, SupportsRead: true, SupportsUpdate: true, SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true, HardDeleteRequiresArchived: true, RestoreRequiresActiveParent: true, ArchivePolicy: NoAction, RestorePolicy: NoAction, DeletePolicy: BlockIfActiveChildren},
	"campaigns":           {ResourceName: "campaigns", TableName: "campaigns", SupportsCreate: true, SupportsRead: true, SupportsUpdate: true, SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true, HardDeleteRequiresArchived: true, ArchivePolicy: NoAction, RestorePolicy: NoAction, DeletePolicy: BlockIfActiveChildren},
	"workorders":          {ResourceName: "work-orders", TableName: "workorders", SupportsCreate: true, SupportsRead: true, SupportsUpdate: true, SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true, HardDeleteRequiresArchived: true, RestoreRequiresActiveParent: true, ArchivePolicy: CascadeArchive, RestorePolicy: CascadeArchive, DeletePolicy: BlockIfActiveChildren},
	"labors":              {ResourceName: "labors", TableName: "labors", SupportsCreate: true, SupportsRead: true, SupportsUpdate: true, SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true, HardDeleteRequiresArchived: true, RestoreRequiresActiveParent: true, ArchivePolicy: NoAction, RestorePolicy: NoAction, DeletePolicy: BlockIfActiveChildren},
	"supplies":            {ResourceName: "supplies", TableName: "supplies", SupportsCreate: true, SupportsRead: true, SupportsUpdate: true, SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true, HardDeleteRequiresArchived: true, RestoreRequiresActiveParent: true, ArchivePolicy: NoAction, RestorePolicy: NoAction, DeletePolicy: BlockIfActiveChildren},
	"supply_movements":    {ResourceName: "supply-movements", TableName: "supply_movements", SupportsCreate: true, SupportsRead: true, SupportsUpdate: true, SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true, HardDeleteRequiresArchived: true, RestoreRequiresActiveParent: true, ArchivePolicy: NoAction, RestorePolicy: NoAction, DeletePolicy: BlockIfActiveChildren},
	"managers":            {ResourceName: "managers", TableName: "managers", SupportsCreate: true, SupportsRead: true, SupportsUpdate: true, SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true, HardDeleteRequiresArchived: true, ArchivePolicy: NoAction, RestorePolicy: NoAction, DeletePolicy: BlockIfActiveChildren},
	"investors":           {ResourceName: "investors", TableName: "investors", SupportsCreate: true, SupportsRead: true, SupportsUpdate: true, SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true, HardDeleteRequiresArchived: true, ArchivePolicy: NoAction, RestorePolicy: NoAction, DeletePolicy: BlockIfActiveChildren},
	"providers":           {ResourceName: "providers", TableName: "providers", SupportsCreate: true, SupportsRead: true, SupportsUpdate: true, SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true, HardDeleteRequiresArchived: true, ArchivePolicy: NoAction, RestorePolicy: NoAction, DeletePolicy: BlockIfActiveChildren},
	"categories":          {ResourceName: "categories", TableName: "categories", SupportsCreate: true, SupportsRead: true, SupportsUpdate: true, SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true, HardDeleteRequiresArchived: true, ArchivePolicy: NoAction, RestorePolicy: NoAction, DeletePolicy: BlockIfActiveChildren},
	"crops":               {ResourceName: "crops", TableName: "crops", SupportsCreate: true, SupportsRead: true, SupportsUpdate: true, SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true, HardDeleteRequiresArchived: true, ArchivePolicy: NoAction, RestorePolicy: NoAction, DeletePolicy: BlockIfActiveChildren},
	"class_types":         {ResourceName: "types", TableName: "types", SupportsCreate: true, SupportsRead: true, SupportsUpdate: true, SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true, HardDeleteRequiresArchived: true, ArchivePolicy: NoAction, RestorePolicy: NoAction, DeletePolicy: BlockIfActiveChildren},
	"lease_types":         {ResourceName: "lease-types", TableName: "lease_types", SupportsCreate: true, SupportsRead: true, SupportsUpdate: true, SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true, HardDeleteRequiresArchived: true, ArchivePolicy: NoAction, RestorePolicy: NoAction, DeletePolicy: BlockIfActiveChildren},
	"business_parameters": {ResourceName: "business-parameters", TableName: "business_parameters", SupportsCreate: true, SupportsRead: true, SupportsUpdate: true, SupportsArchive: true, SupportsRestore: true, SupportsHardDelete: true, HardDeleteRequiresArchived: true, ArchivePolicy: NoAction, RestorePolicy: NoAction, DeletePolicy: NoAction},
}
