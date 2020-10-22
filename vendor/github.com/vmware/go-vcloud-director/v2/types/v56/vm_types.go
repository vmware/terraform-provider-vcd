package types

import "encoding/xml"

// VM represents a virtual machine
// Type: VmType
// Namespace: http://www.vmware.com/vcloud/v1.5
// Description: Represents a virtual machine.
// Since: 0.9
type VM struct {
	// Attributes
	XMLName xml.Name `xml:"Vm"`
	Ovf     string   `xml:"xmlns:ovf,attr,omitempty"`
	Xsi     string   `xml:"xmlns:xsi,attr,omitempty"`
	Xmlns   string   `xml:"xmlns,attr,omitempty"`

	HREF                    string `xml:"href,attr,omitempty"`                    // The URI of the entity.
	Type                    string `xml:"type,attr,omitempty"`                    // The MIME type of the entity.
	ID                      string `xml:"id,attr,omitempty"`                      // The entity identifier, expressed in URN format. The value of this attribute uniquely identifies the entity, persists for the life of the entity, and is never reused
	OperationKey            string `xml:"operationKey,attr,omitempty"`            // Optional unique identifier to support idempotent semantics for create and delete operations.
	Name                    string `xml:"name,attr"`                              // The name of the entity.
	Status                  int    `xml:"status,attr,omitempty"`                  // Creation status of the resource entity.
	Deployed                bool   `xml:"deployed,attr,omitempty"`                // True if the virtual machine is deployed.
	NeedsCustomization      bool   `xml:"needsCustomization,attr,omitempty"`      // True if this virtual machine needs customization.
	NestedHypervisorEnabled bool   `xml:"nestedHypervisorEnabled,attr,omitempty"` // True if hardware-assisted CPU virtualization capabilities in the host should be exposed to the guest operating system.
	// Elements
	Link        LinkList         `xml:"Link,omitempty"`        // A reference to an entity or operation associated with this object.
	Description string           `xml:"Description,omitempty"` // Optional description.
	Tasks       *TasksInProgress `xml:"Tasks,omitempty"`       // A list of queued, running, or recently completed tasks associated with this entity.
	Files       *FilesList       `xml:"FilesList,omitempty"`   // Represents a list of files to be transferred (uploaded or downloaded). Each File in the list is part of the ResourceEntity.
	VAppParent  *Reference       `xml:"VAppParent,omitempty"`  // Reserved. Unimplemented.
	// TODO: OVF Sections to be implemented
	// Section OVF_Section `xml:"Section,omitempty"
	DateCreated string `xml:"DateCreated,omitempty"` // Creation date/time of the vApp.

	// Section ovf:VirtualHardwareSection
	VirtualHardwareSection *VirtualHardwareSection `xml:"VirtualHardwareSection,omitempty"`

	// FIXME: Upstream bug? Missing NetworkConnectionSection
	NetworkConnectionSection *NetworkConnectionSection `xml:"NetworkConnectionSection,omitempty"`

	VAppScopedLocalID string `xml:"VAppScopedLocalId,omitempty"` // A unique identifier for the virtual machine in the scope of the vApp.

	Snapshots *SnapshotSection `xml:"SnapshotSection,omitempty"`

	// TODO: OVF Sections to be implemented
	// Environment OVF_Environment `xml:"Environment,omitempty"

	VmSpecSection *VmSpecSection `xml:"VmSpecSection,omitempty"`

	// GuestCustomizationSection contains settings for VM customization like admin password, SID
	// changes, domain join configuration, etc
	GuestCustomizationSection *GuestCustomizationSection `xml:"GuestCustomizationSection,omitempty"`

	VMCapabilities *VmCapabilities `xml:"VmCapabilities,omitempty"` // Allows you to specify certain capabilities of this virtual machine.
	StorageProfile *Reference      `xml:"StorageProfile,omitempty"` // A reference to a storage profile to be used for this object. The specified storage profile must exist in the organization vDC that contains the object. If not specified, the default storage profile for the vDC is used.
	ProductSection *ProductSection `xml:"ProductSection,omitempty"`
	ComputePolicy  *ComputePolicy  `xml:"ComputePolicy,omitempty"` // accessible only from version API 33.0
	Media          *Reference      `xml:"Media,omitempty"`         // Reference to the media object to insert in a new VM.
}

// VmSpecSection from VM struct
type VmSpecSection struct {
	Modified          *bool             `xml:"Modified,attr,omitempty"`
	Info              string            `xml:"ovf:Info"`
	OsType            string            `xml:"OsType,omitempty"`            // The type of the OS. This parameter may be omitted when using the VmSpec to update the contents of an existing VM.
	NumCpus           *int              `xml:"NumCpus,omitempty"`           // Number of CPUs. This parameter may be omitted when using the VmSpec to update the contents of an existing VM.
	NumCoresPerSocket *int              `xml:"NumCoresPerSocket,omitempty"` // Number of cores among which to distribute CPUs in this virtual machine. This parameter may be omitted when using the VmSpec to update the contents of an existing VM.
	CpuResourceMhz    *CpuResourceMhz   `xml:"CpuResourceMhz,omitempty"`    // CPU compute resources. This parameter may be omitted when using the VmSpec to update the contents of an existing VM.
	MemoryResourceMb  *MemoryResourceMb `xml:"MemoryResourceMb"`            // Memory compute resources. This parameter may be omitted when using the VmSpec to update the contents of an existing VM.
	MediaSection      *MediaSection     `xml:"MediaSection,omitempty"`      // The media devices of this VM.
	DiskSection       *DiskSection      `xml:"DiskSection,omitempty"`       // virtual disks of this VM.
	HardwareVersion   *HardwareVersion  `xml:"HardwareVersion"`             // vSphere name of Virtual Hardware Version of this VM. Example: vmx-13 - This parameter may be omitted when using the VmSpec to update the contents of an existing VM.
	VmToolsVersion    string            `xml:"VmToolsVersion,omitempty"`    // VMware tools version of this VM.
	VirtualCpuType    string            `xml:"VirtualCpuType,omitempty"`    // The capabilities settings for this VM. This parameter may be omitted when using the VmSpec to update the contents of an existing VM.
	TimeSyncWithHost  *bool             `xml:"TimeSyncWithHost,omitempty"`  // Synchronize the VM's time with the host.
}

// RecomposeVAppParamsForEmptyVm represents a vApp structure which allows to create VM.
type RecomposeVAppParamsForEmptyVm struct {
	XMLName          xml.Name    `xml:"RecomposeVAppParams"`
	XmlnsVcloud      string      `xml:"xmlns,attr"`
	XmlnsOvf         string      `xml:"xmlns:ovf,attr"`
	CreateItem       *CreateItem `xml:"CreateItem,omitempty"`
	AllEULAsAccepted bool        `xml:"AllEULAsAccepted,omitempty"`
}

// CreateItem represents structure to create VM, part of RecomposeVAppParams structure.
type CreateItem struct {
	Name                      string                     `xml:"name,attr,omitempty"`
	Description               string                     `xml:"Description,omitempty"`
	GuestCustomizationSection *GuestCustomizationSection `xml:"GuestCustomizationSection,omitempty"`
	NetworkConnectionSection  *NetworkConnectionSection  `xml:"NetworkConnectionSection,omitempty"`
	VmSpecSection             *VmSpecSection             `xml:"VmSpecSection,omitempty"`
	StorageProfile            *Reference                 `xml:"StorageProfile,omitempty"`
	ComputePolicy             *ComputePolicy             `xml:"ComputePolicy,omitempty"` // accessible only from version API 33.0
	BootImage                 *Media                     `xml:"Media,omitempty"`         // boot image as vApp template. Href, Id and name needed.
}

// ComputePolicy represents structure to manage VM compute polices, part of RecomposeVAppParams structure.
type ComputePolicy struct {
	HREF                   string     `xml:"href,attr,omitempty"`
	Type                   string     `xml:"type,attr,omitempty"`
	Link                   *Link      `xml:"Link,omitempty"`                   // A reference to an entity or operation associated with this object.
	VmPlacementPolicy      *Reference `xml:"VmPlacementPolicy,omitempty"`      // VdcComputePolicy that defines VM's placement on a host through various affinity constraints.
	VmPlacementPolicyFinal *bool      `xml:"VmPlacementPolicyFinal,omitempty"` // True indicates that the placement policy cannot be removed from a VM that is instantiated with it. This value defaults to false.
	VmSizingPolicy         *Reference `xml:"VmSizingPolicy,omitempty"`         // VdcComputePolicy that defines VM's sizing and resource allocation.
	VmSizingPolicyFinal    *bool      `xml:"VmSizingPolicyFinal,omitempty"`    // True indicates that the sizing policy cannot be removed from a VM that is instantiated with it. This value defaults to false.
}
