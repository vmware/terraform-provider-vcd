package types

// Type: AdminCatalogRecord
// Namespace: http://www.vmware.com/vcloud/v1.5
// https://code.vmware.com/apis/287/vcloud#/doc/doc/types/QueryResultCatalogRecordType.html
// Issue that description partly matches with what is returned
// Description: Represents Catalog record
// Since: 1.5
type AdminCatalogRecord struct {
	HREF                    string    `xml:"href,attr,omitempty"`
	ID                      string    `xml:"id,attr,omitempty"`
	Type                    string    `xml:"type,attr,omitempty"`
	Name                    string    `xml:"name,attr,omitempty"`
	Description             string    `xml:"description,attr,omitempty"`
	IsPublished             bool      `xml:"isPublished,attr,omitempty"`
	IsShared                bool      `xml:"isShared,attr,omitempty"`
	CreationDate            string    `xml:"creationDate,attr,omitempty"`
	OrgName                 string    `xml:"orgName,attr,omitempty"`
	OwnerName               string    `xml:"ownerName,attr,omitempty"`
	NumberOfVAppTemplates   int64     `xml:"numberOfVAppTemplates,attr,omitempty"`
	NumberOfMedia           int64     `xml:"numberOfMedia,attr,omitempty"`
	Owner                   string    `xml:"owner,attr,omitempty"`
	PublishSubscriptionType string    `xml:"publishSubscriptionType,attr,omitempty"`
	Version                 int64     `xml:"version,attr,omitempty"`
	Status                  string    `xml:"status,attr,omitempty"`
	Link                    *Link     `xml:"Link,omitempty"`
	Vdc                     *Metadata `xml:"Metadata,omitempty"`
}

// QueryResultEdgeGatewayRecordsType is a container for query results in records format.
// Type: QueryResultRecordsType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Container for query results in records format.
// Since: 1.5
type QueryResultEdgeGatewayRecordsType struct {
	// Attributes
	HREF     string  `xml:"href,attr,omitempty"`     // The URI of the entity.
	Type     string  `xml:"type,attr,omitempty"`     // The MIME type of the entity.
	Name     string  `xml:"name,attr,omitempty"`     // The name of the entity.
	Page     int     `xml:"page,attr,omitempty"`     // Page of the result set that this container holds. The first page is page number 1.
	PageSize int     `xml:"pageSize,attr,omitempty"` // Page size, as a number of records or references.
	Total    float64 `xml:"total,attr,omitempty"`    // Total number of records or references in the container.
	// Elements
	Link              []*Link                             `xml:"Link,omitempty"`    // A reference to an entity or operation associated with this object.
	EdgeGatewayRecord []*QueryResultEdgeGatewayRecordType `xml:"EdgeGatewayRecord"` // A record representing a EdgeGateway result.
}

type QueryResultRecordsType struct {
	// Attributes
	HREF     string  `xml:"href,attr,omitempty"`     // The URI of the entity.
	Type     string  `xml:"type,attr,omitempty"`     // The MIME type of the entity.
	Name     string  `xml:"name,attr,omitempty"`     // The name of the entity.
	Page     int     `xml:"page,attr,omitempty"`     // Page of the result set that this container holds. The first page is page number 1.
	PageSize int     `xml:"pageSize,attr,omitempty"` // Page size, as a number of records or references.
	Total    float64 `xml:"total,attr,omitempty"`    // Total number of records or references in the container.
	// Elements
	Link                            []*Link                                           `xml:"Link,omitempty"`                  // A reference to an entity or operation associated with this object.
	EdgeGatewayRecord               []*QueryResultEdgeGatewayRecordType               `xml:"EdgeGatewayRecord"`               // A record representing a EdgeGateway result.
	VMRecord                        []*QueryResultVMRecordType                        `xml:"VMRecord"`                        // A record representing a VM result.
	AdminVMRecord                   []*QueryResultVMRecordType                        `xml:"AdminVMRecord"`                   // A record representing a Admin VM result.
	VAppRecord                      []*QueryResultVAppRecordType                      `xml:"VAppRecord"`                      // A record representing a VApp result.
	OrgVdcStorageProfileRecord      []*QueryResultOrgVdcStorageProfileRecordType      `xml:"OrgVdcStorageProfileRecord"`      // A record representing storage profiles
	MediaRecord                     []*MediaRecordType                                `xml:"MediaRecord"`                     // A record representing media
	AdminMediaRecord                []*MediaRecordType                                `xml:"AdminMediaRecord"`                // A record representing Admin media
	VMWProviderVdcRecord            []*QueryResultVMWProviderVdcRecordType            `xml:"VMWProviderVdcRecord"`            // A record representing a Provider VDC result.
	ProviderVdcStorageProfileRecord []*QueryResultProviderVdcStorageProfileRecordType `xml:"ProviderVdcStorageProfileRecord"` // A record representing a Provider VDC storage profile result
	NetworkPoolRecord               []*QueryResultNetworkPoolRecordType               `xml:"NetworkPoolRecord"`               // A record representing a network pool
	DiskRecord                      []*DiskRecordType                                 `xml:"DiskRecord"`                      // A record representing a independent Disk.
	AdminDiskRecord                 []*DiskRecordType                                 `xml:"AdminDiskRecord"`                 // A record representing a independent Disk.
	VirtualCenterRecord             []*QueryResultVirtualCenterRecordType             `xml:"VirtualCenterRecord"`             // A record representing a vSphere server
	PortGroupRecord                 []*PortGroupRecordType                            `xml:"PortgroupRecord"`                 // A record representing a port group
	OrgVdcNetworkRecord             []*QueryResultOrgVdcNetworkRecordType             `xml:"OrgVdcNetworkRecord"`             // A record representing a org VDC network
	AdminCatalogRecord              []*AdminCatalogRecord                             `xml:"AdminCatalogRecord"`              // A record representing a catalog
}

// QueryResultEdgeGatewayRecordType represents an edge gateway record as query result.
type QueryResultEdgeGatewayRecordType struct {
	// Attributes
	HREF                string `xml:"href,attr,omitempty"`                // The URI of the entity.
	Type                string `xml:"type,attr,omitempty"`                // The MIME type of the entity.
	Name                string `xml:"name,attr,omitempty"`                // EdgeGateway name.
	Vdc                 string `xml:"vdc,attr,omitempty"`                 // VDC Reference or ID
	NumberOfExtNetworks int    `xml:"numberOfExtNetworks,attr,omitempty"` // Number of external networks connected to the edgeGateway.	Yes	Yes
	NumberOfOrgNetworks int    `xml:"numberOfOrgNetworks,attr,omitempty"` // Number of org VDC networks connected to the edgeGateway	Yes	Yes
	IsBusy              bool   `xml:"isBusy,attr"`                        // True if this Edge Gateway is busy.	Yes	Yes
	GatewayStatus       string `xml:"gatewayStatus,attr,omitempty"`       //
	HaStatus            string `xml:"haStatus,attr,omitempty"`            // High Availability Status of the edgeGateway	Yes	Yes
}

// QueryResultVMRecordType represents a VM record as query result.
type QueryResultVMRecordType struct {
	// Attributes
	HREF                 string    `xml:"href,attr,omitempty"` // The URI of the entity.
	ID                   string    `xml:"id,attr,omitempty"`
	Name                 string    `xml:"name,attr,omitempty"`          // VM name.
	Type                 string    `xml:"type,attr,omitempty"`          // Contains the type of the resource.
	ContainerName        string    `xml:"containerName,attr,omitempty"` // The name of the vApp or vApp template that contains this VM.
	ContainerID          string    `xml:"container,attr,omitempty"`     // The ID of the vApp or vApp template that contains this VM.
	OwnerName            string    `xml:"ownerName,attr,omitempty"`
	Owner                string    `xml:"owner,attr,omitempty"`
	VdcHREF              string    `xml:"vdc,attr,omitempty"`
	VAppTemplate         bool      `xml:"isVAppTemplate,attr,omitempty"`
	Deleted              bool      `xml:"isDeleted,attr,omitempty"`
	GuestOS              string    `xml:"guestOs,attr,omitempty"`
	Cpus                 int       `xml:"numberOfCpus,attr,omitempty"`
	MemoryMB             int       `xml:"memoryMB,attr,omitempty"`
	Status               string    `xml:"status,attr,omitempty"`
	NetworkName          string    `xml:"networkName,attr,omitempty"`
	NetworkHref          string    `xml:"network,attr,omitempty"`
	IpAddress            string    `xml:"ipAddress,attr,omitempty"` // If configured, the IP Address of the VM on the primary network, otherwise empty.
	Busy                 bool      `xml:"isBusy,attr,omitempty"`
	Deployed             bool      `xml:"isDeployed,attr,omitempty"` // True if the virtual machine is deployed.
	Published            bool      `xml:"isPublished,attr,omitempty"`
	CatalogName          string    `xml:"catalogName,attr,omitempty"`
	HardwareVersion      int       `xml:"hardwareVersion,attr,omitempty"`
	VmToolsStatus        string    `xml:"vmToolsStatus,attr,omitempty"`
	MaintenanceMode      bool      `xml:"isInMaintenanceMode,attr,omitempty"`
	AutoNature           bool      `xml:"isAutoNature,attr,omitempty"` //  	True if the parent vApp is a managed vApp
	StorageProfileName   string    `xml:"storageProfileName,attr,omitempty"`
	GcStatus             string    `xml:"gcStatus,attr,omitempty"` // GC status of this VM.
	AutoUndeployDate     string    `xml:"autoUndeployDate,attr,omitempty"`
	AutoDeleteDate       string    `xml:"autoDeleteDate,attr,omitempty"`
	AutoUndeployNotified bool      `xml:"isAutoUndeployNotified,attr,omitempty"`
	AutoDeleteNotified   bool      `xml:"isAutoDeleteNotified,attr,omitempty"`
	Link                 []*Link   `xml:"Link,omitempty"`
	MetaData             *Metadata `xml:"Metadata,omitempty"`
}

// QueryResultVAppRecordType represents a VM record as query result.
type QueryResultVAppRecordType struct {
	// Attributes
	HREF                    string `xml:"href,attr,omitempty"`         // The URI of the entity.
	Name                    string `xml:"name,attr"`                   // The name of the entity.
	CreationDate            string `xml:"creationDate,attr,omitempty"` // Creation date/time of the vApp.
	Busy                    bool   `xml:"isBusy,attr,omitempty"`
	Deployed                bool   `xml:"isDeployed,attr,omitempty"` // True if the vApp is deployed.
	Enabled                 bool   `xml:"isEnabled,attr,omitempty"`
	Expired                 bool   `xml:"isExpired,attr,omitempty"`
	MaintenanceMode         bool   `xml:"isInMaintenanceMode,attr,omitempty"`
	Public                  bool   `xml:"isPublic,attr,omitempty"`
	OwnerName               string `xml:"ownerName,attr,omitempty"`
	Status                  string `xml:"status,attr,omitempty"`
	VdcHREF                 string `xml:"vdc,attr,omitempty"`
	VdcName                 string `xml:"vdcName,attr,omitempty"`
	NumberOfVMs             int    `xml:"numberOfVMs,attr,omitempty"`
	NumberOfCPUs            int    `xml:"numberOfCpus,attr,omitempty"`
	CpuAllocationMhz        int    `xml:"cpuAllocationMhz,attr,omitempty"`
	CpuAllocationInMhz      int    `xml:"cpuAllocationInMhz,attr,omitempty"`
	StorageKB               int    `xml:"storageKB,attr,omitempty"`
	MemoryAllocationMB      int    `xml:"memoryAllocationMB,attr,omitempty"`
	AutoDeleteNotified      bool   `xml:"isAutoDeleteNotified,attr,omitempty"`
	AutoUndeployNotified    bool   `xml:"isAutoUndeployNotified,attr,omitempty"`
	VdcEnabled              bool   `xml:"isVdcEnabled,attr,omitempty"`
	HonorBootOrder          bool   `xml:"honorBookOrder,attr,omitempty"`
	HighestSupportedVersion int    `xml:"pvdcHighestSupportedHardwareVersion,attr,omitempty"`
	LowestHardwareVersion   int    `xml:"lowestHardwareVersionInVApp,attr,omitempty"`
	TaskHREF                string `xml:"task,attr,omitempty"`
	TaskStatusName          string `xml:"taskStatusName,attr,omitempty"`
	TaskStatus              string `xml:"TaskStatus,attr,omitempty"`
	TaskDetails             string `xml:"taskDetails,attr,omitempty"`
}

// QueryResultOrgVdcStorageProfileRecordType represents a storage
// profile as query result.
type QueryResultOrgVdcStorageProfileRecordType struct {
	// Attributes
	HREF                    string `xml:"href,attr,omitempty"` // The URI of the entity.
	Name                    string `xml:"name,attr,omitempty"` // Storage Profile name.
	VdcHREF                 string `xml:"vdc,attr,omitempty"`
	VdcName                 string `xml:"vdcName,attr,omitempty"`
	IsDefaultStorageProfile bool   `xml:"isDefaultStorageProfile,attr,omitempty"`
	IsEnabled               bool   `xml:"isEnabled,attr,omitempty"`
	IsVdcBusy               bool   `xml:"isVdcBusy,attr,omitempty"`
	NumberOfConditions      int    `xml:"numberOfConditions,attr,omitempty"`
	StorageUsedMB           int    `xml:"storageUsedMB,attr,omitempty"`
	StorageLimitMB          int    `xml:"storageLimitMB,attr,omitempty"`
}

// QueryResultVMWProviderVdcRecordType represents a Provider VDC as query result.
type QueryResultVMWProviderVdcRecordType struct {
	// Attributes
	HREF                    string `xml:"href,attr,omitempty"` // The URI of the entity.
	Name                    string `xml:"name,attr,omitempty"` // Provider VDC name.
	Status                  string `xml:"status,attr,omitempty"`
	IsBusy                  bool   `xml:"isBusy,attr,omitempty"`
	IsDeleted               bool   `xml:"isDeleted,attr,omitempty"`
	IsEnabled               bool   `xml:"isEnabled,attr,omitempty"`
	CpuAllocationMhz        int    `xml:"cpuAllocationMhz,attr,omitempty"`
	CpuLimitMhz             int    `xml:"cpuLimitMhz,attr,omitempty"`
	CpuUsedMhz              int    `xml:"cpuUsedMhz,attr,omitempty"`
	NumberOfDatastores      int    `xml:"numberOfDatastores,attr,omitempty"`
	NumberOfStorageProfiles int    `xml:"numberOfStorageProfiles,attr,omitempty"`
	NumberOfVdcs            int    `xml:"numberOfVdcs,attr,omitempty"`
	MemoryAllocationMB      int64  `xml:"memoryAllocationMB,attr,omitempty"`
	MemoryLimitMB           int64  `xml:"memoryLimitMB,attr,omitempty"`
	MemoryUsedMB            int64  `xml:"memoryUsedMB,attr,omitempty"`
	StorageAllocationMB     int64  `xml:"storageAllocationMB,attr,omitempty"`
	StorageLimitMB          int64  `xml:"storageLimitMB,attr,omitempty"`
	StorageUsedMB           int64  `xml:"storageUsedMB,attr,omitempty"`
	CpuOverheadMhz          int64  `xml:"cpuOverheadMhz,attr,omitempty"`
	StorageOverheadMB       int64  `xml:"storageOverheadMB,attr,omitempty"`
	MemoryOverheadMB        int64  `xml:"memoryOverheadMB,attr,omitempty"`
}

// QueryResultProviderVdcStorageProfileRecordType represents a Provider VDC storage profile as query result.
type QueryResultProviderVdcStorageProfileRecordType struct {
	// Attributes
	HREF                 string `xml:"href,attr,omitempty"` // The URI of the entity.
	Name                 string `xml:"name,attr,omitempty"` // Provider VDC Storage Profile name.
	ProviderVdcHREF      string `xml:"providerVdc,attr,omitempty"`
	VcHREF               string `xml:"vc,attr,omitempty"`
	StorageProfileMoref  string `xml:"storageProfileMoref,attr,omitempty"`
	IsEnabled            bool   `xml:"isEnabled,attr,omitempty"`
	StorageProvisionedMB int64  `xml:"storageProvisionedMB,attr,omitempty"`
	StorageRequestedMB   int64  `xml:"storageRequestedMB,attr,omitempty"`
	StorageTotalMB       int64  `xml:"storageTotalMB,attr,omitempty"`
	StorageUsedMB        int64  `xml:"storageUsedMB,attr,omitempty"`
	NumberOfConditions   int    `xml:"numberOfConditions,attr,omitempty"`
}

// QueryResultNetworkPoolRecordType represents a network pool as query result.
type QueryResultNetworkPoolRecordType struct {
	// Attributes
	HREF            string `xml:"href,attr,omitempty"` // The URI of the entity.
	Name            string `xml:"name,attr,omitempty"` // Network pool name.
	IsBusy          bool   `xml:"isBusy,attr,omitempty"`
	NetworkPoolType int    `xml:"networkPoolType,attr,omitempty"`
}

// Type: QueryResultVirtualCenterRecordType
// Namespace: http://www.vmware.com/vcloud/v1.5
// https://vdc-repo.vmware.com/vmwb-repository/dcr-public/7a028e78-bd37-4a6a-8298-9c26c7eeb9aa/09142237-dd46-4dee-8326-e07212fb63a8/doc/doc/types/QueryResultVirtualCenterRecordType.html
// Description: Type for a single virtualCenter query result in records format.
// Since: 1.5
type QueryResultVirtualCenterRecordType struct {
	HREF          string `xml:"href,attr,omitempty"`
	Name          string `xml:"name,attr,omitempty"`
	IsBusy        bool   `xml:"isBusy,attr,omitempty"`
	IsEnabled     bool   `xml:"isEnabled,attr,omitempty"`
	IsSupported   bool   `xml:"isSupported,attr,omitempty"`
	ListenerState string `xml:"listenerState,attr,omitempty"`
	Status        string `xml:"stats,attr,omitempty"`
	Url           string `xml:"url,attr,omitempty"`
	UserName      string `xml:"userName,attr,omitempty"`
	VcVersion     string `xml:"vcVersion,attr,omitempty"`
	UUID          string `xml:"uuid,attr,omitempty"`
	VsmIP         string `xml:"vsmIP,attr,omitempty"`
}

// Type: MediaRecord
// Namespace: http://www.vmware.com/vcloud/v1.5
// https://code.vmware.com/apis/287/vcloud#/doc/doc/operations/GET-MediasFromQuery.html
// Issue that description partly matches with what is returned
// Description: Represents Media record
// Since: 1.5
type MediaRecordType struct {
	HREF               string `xml:"href,attr,omitempty"`
	ID                 string `xml:"id,attr,omitempty"`
	Type               string `xml:"type,attr,omitempty"`
	OwnerName          string `xml:"ownerName,attr,omitempty"`
	CatalogName        string `xml:"catalogName,attr,omitempty"`
	IsPublished        bool   `xml:"isPublished,attr,omitempty"`
	Name               string `xml:"name,attr"`
	Vdc                string `xml:"vdc,attr,omitempty"`
	VdcName            string `xml:"vdcName,attr,omitempty"`
	Org                string `xml:"org,attr,omitempty"`
	CreationDate       string `xml:"creationDate,attr,omitempty"`
	IsBusy             bool   `xml:"isBusy,attr,omitempty"`
	StorageB           int64  `xml:"storageB,attr,omitempty"`
	Owner              string `xml:"owner,attr,omitempty"`
	Catalog            string `xml:"catalog,attr,omitempty"`
	CatalogItem        string `xml:"catalogItem,attr,omitempty"`
	Status             string `xml:"status,attr,omitempty"`
	StorageProfileName string `xml:"storageProfileName,attr,omitempty"`
	Version            int64  `xml:"version,attr,omitempty"`
	LastSuccessfulSync string `xml:"lastSuccessfulSync,attr,omitempty"`
	TaskStatusName     string `xml:"taskStatusName,attr,omitempty"`
	IsInCatalog        bool   `xml:"isInCatalog,attr,omitempty"`
	Task               string `xml:"task,attr,omitempty"`
	IsIso              bool   `xml:"isIso,attr,omitempty"`
	IsVdcEnabled       bool   `xml:"isVdcEnabled,attr,omitempty"`
	TaskStatus         string `xml:"taskStatus,attr,omitempty"`
	TaskDetails        string `xml:"taskDetails,attr,omitempty"`
}

// Represents an independent disk record
// Reference: vCloud API 27.0 - DiskType
// https://code.vmware.com/apis/287/vcloud#/doc/doc/types/QueryResultDiskRecordType.html
type DiskRecordType struct {
	Xmlns              string  `xml:"xmlns,attr,omitempty"`
	HREF               string  `xml:"href,attr,omitempty"`
	Id                 string  `xml:"id,attr,omitempty"`
	Type               string  `xml:"type,attr,omitempty"`
	Name               string  `xml:"name,attr,omitempty"`
	Vdc                string  `xml:"vdc,attr,omitempty"`
	SizeB              int64   `xml:"sizeB,attr,omitempty"`
	DataStore          string  `xml:"dataStore,attr,omitempty"`
	DataStoreName      string  `xml:"datastoreName,attr,omitempty"`
	OwnerName          string  `xml:"ownerName,attr,omitempty"`
	VdcName            string  `xml:"vdcName,attr,omitempty"`
	Task               string  `xml:"task,attr,omitempty"`
	StorageProfile     string  `xml:"storageProfile,attr,omitempty"`
	StorageProfileName string  `xml:"storageProfileName,attr,omitempty"`
	Status             string  `xml:"status,attr,omitempty"`
	BusType            string  `xml:"busType,attr,omitempty"`
	BusSubType         string  `xml:"busSubType,attr,omitempty"`
	BusTypeDesc        string  `xml:"busTypeDesc,attr,omitempty"`
	IsAttached         bool    `xml:"isAttached,attr,omitempty"`
	Description        string  `xml:"description,attr,omitempty"`
	Link               []*Link `xml:"Link,omitempty"`
}

// Represents port group
// Reference: vCloud API 27.0 - Port group type
// https://code.vmware.com/apis/72/doc/doc/types/QueryResultPortgroupRecordType.html
type PortGroupRecordType struct {
	Xmlns         string  `xml:"xmlns,attr,omitempty"`
	HREF          string  `xml:"href,attr,omitempty"`
	Id            string  `xml:"id,attr,omitempty"`
	Type          string  `xml:"type,attr,omitempty"`
	MoRef         string  `xml:"moref,attr,omitempty"`
	Name          string  `xml:"name,attr,omitempty"`
	PortgroupType string  `xml:"portgroupType,attr,omitempty"`
	Vc            string  `xml:"vc,attr,omitempty"`
	VcName        string  `xml:"vcName,attr,omitempty"`
	IsVCEnabled   bool    `xml:"isVCEnabled,attr,omitempty"`
	Network       string  `xml:"network,attr,omitempty"`
	NetworkName   string  `xml:"networkName,attr,omitempty"`
	ScopeType     int     `xml:"scopeType,attr,omitempty"` // Scope of network using the portgroup(1=Global, 2=Organization, 3=vApp)
	Link          []*Link `xml:"Link,omitempty"`
}

// Represents org VDC Network
// Reference: vCloud API 27.0 - Org VDC Network
// https://code.vmware.com/apis/72/doc/doc/types/QueryResultOrgVdcNetworkRecordType.html
type QueryResultOrgVdcNetworkRecordType struct {
	Xmlns              string  `xml:"xmlns,attr,omitempty"`
	HREF               string  `xml:"href,attr,omitempty"`
	Id                 string  `xml:"id,attr,omitempty"`
	Type               string  `xml:"type,attr,omitempty"`
	Name               string  `xml:"name,attr,omitempty"`
	DefaultGateway     string  `xml:"defaultGateway,attr,omitempty"`
	Netmask            string  `xml:"netmask,attr,omitempty"`
	Dns1               string  `xml:"dns1,attr,omitempty"`
	Dns2               string  `xml:"dns2,attr,omitempty"`
	DnsSuffix          string  `xml:"dnsSuffix,attr,omitempty"`
	LinkType           int     `xml:"linkType,attr,omitempty"` // 0 = direct, 1 = routed, 2 = isolated
	ConnectedTo        string  `xml:"connectedTo,attr,omitempty"`
	Vdc                string  `xml:"vdc,attr,omitempty"`
	IsBusy             bool    `xml:"isBusy,attr,omitempty"`
	IsShared           bool    `xml:"isShared,attr,omitempty"`
	VdcName            string  `xml:"vdcName,attr,omitempty"`
	IsIpScopeInherited bool    `xml:"isIpScopeInherited,attr,omitempty"`
	Link               []*Link `xml:"Link,omitempty"`
}
