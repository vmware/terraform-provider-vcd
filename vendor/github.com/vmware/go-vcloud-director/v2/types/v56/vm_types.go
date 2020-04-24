package types

import "encoding/xml"

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

type RecomposeVAppParamsForEmptyVm struct {
	XMLName          xml.Name    `xml:"RecomposeVAppParams"`
	XmlnsVcloud      string      `xml:"xmlns,attr"`
	XmlnsOvf         string      `xml:"xmlns:ovf,attr"`
	CreateItem       *CreateItem `xml:"CreateItem,omitempty"`
	AllEULAsAccepted bool        `xml:"AllEULAsAccepted,omitempty"`
}

type CreateItem struct {
	//XMLName                   xml.Name                   `xml:"CreateItem"`
	Name                      string                     `xml:"name,attr,omitempty"`
	Description               string                     `xml:"Description,omitempty"`
	GuestCustomizationSection *GuestCustomizationSection `xml:"GuestCustomizationSection,omitempty"`
	NetworkConnectionSection  *NetworkConnectionSection  `xml:"NetworkConnectionSection,omitempty"`
	VmSpecSection             *VmSpecSection             `xml:"VmSpecSection,omitempty"`
	BootImage                 *Media                     `xml:"Media,omitempty"` // boot image as vApp template. Href, Id and name needed.
}
