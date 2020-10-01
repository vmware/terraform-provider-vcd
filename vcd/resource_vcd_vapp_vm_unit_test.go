// +build unit ALL

package vcd

import (
	"encoding/xml"
	"testing"

	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// Test_deprecatedReadNetworks was introduced to fix a bug where sometimes testing suite would capture wrong IP because
// vCD API returned unordered NIC list
func Test_deprecatedReadNetworks(t *testing.T) {
	type args struct {
		vmText string
	}
	tests := []struct {
		name    string
		args    args
		ip      string
		mac     string
		wantErr bool
	}{
		{
			name: "ReverseNicOrder",
			args: args{vmText: testGetVmTextWithCustomNetworkSection(networkSectionReversedOrder)},
			// Expect IP and MAC to be for primary network
			ip:      "0.0.0.0",
			mac:     "00:50:56:29:01:45",
			wantErr: false,
		},
		{
			name: "CorrectNicOrder",
			args: args{vmText: testGetVmTextWithCustomNetworkSection(networkSectionCorrectOrder)},
			// Expect IP and MAC to be for primary network
			ip:      "0.0.0.0",
			mac:     "00:50:56:29:01:45",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var vm types.VM
			err := xml.Unmarshal([]byte(tt.args.vmText), &vm)
			if err != nil {
				t.Errorf("error unmarshaling: %s", err)
			}
			fullVm := govcd.VM{
				VM: &vm,
			}

			gotIp, gotMac, err := deprecatedReadNetworks(fullVm)
			if (err != nil) != tt.wantErr {
				t.Errorf("deprecatedReadNetworks() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotIp != tt.ip {
				t.Errorf("deprecatedReadNetworks() ip = %v, want %v", gotIp, tt.ip)
			}
			if gotMac != tt.mac {
				t.Errorf("deprecatedReadNetworks() mac = %v, want %v", gotMac, tt.mac)
			}
		})
	}
}

// networkSectionReversedOrder holds a NetworkConnectionSection with reversed order of NICs. vCD API does return this in
// unordered fashion very rarely and this used to confuse IP/mac reporting.
const networkSectionReversedOrder = `
<NetworkConnectionSection href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/networkConnectionSection/" type="application/vnd.vmware.vcloud.networkConnectionSection+xml" ovf:required="false">
        <ovf:Info>Specifies the available VM network connections</ovf:Info>
        <PrimaryNetworkConnectionIndex>0</PrimaryNetworkConnectionIndex>
        <NetworkConnection needsCustomization="true" network="none">
            <NetworkConnectionIndex>1</NetworkConnectionIndex>
            <IsConnected>false</IsConnected>
			<IpAddress>1.1.1.1</IpAddress>
            <MACAddress>00:50:56:29:01:46</MACAddress>
            <IpAddressAllocationMode>NONE</IpAddressAllocationMode>
            <NetworkAdapterType>VMXNET3</NetworkAdapterType>
        </NetworkConnection>
        <NetworkConnection needsCustomization="false" network="multinic-net">
            <NetworkConnectionIndex>0</NetworkConnectionIndex>
            <IsConnected>true</IsConnected>
            <IpAddress>0.0.0.0</IpAddress>
            <MACAddress>00:50:56:29:01:45</MACAddress>
            <IpAddressAllocationMode>DHCP</IpAddressAllocationMode>
            <NetworkAdapterType>VMXNET3</NetworkAdapterType>
        </NetworkConnection>
        <Link rel="edit" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/networkConnectionSection/" type="application/vnd.vmware.vcloud.networkConnectionSection+xml"/>
        <Link rel="edit" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/networkConnectionSection/" type="application/vnd.vmware.vcloud.networkConnectionSection+json"/>
    </NetworkConnectionSection>`

const networkSectionCorrectOrder = `
<NetworkConnectionSection href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/networkConnectionSection/" type="application/vnd.vmware.vcloud.networkConnectionSection+xml" ovf:required="false">
        <ovf:Info>Specifies the available VM network connections</ovf:Info>
        <PrimaryNetworkConnectionIndex>0</PrimaryNetworkConnectionIndex>
		<NetworkConnection needsCustomization="false" network="multinic-net">
            <NetworkConnectionIndex>0</NetworkConnectionIndex>
            <IsConnected>true</IsConnected>
            <IpAddress>0.0.0.0</IpAddress>
            <MACAddress>00:50:56:29:01:45</MACAddress>
            <IpAddressAllocationMode>DHCP</IpAddressAllocationMode>
            <NetworkAdapterType>VMXNET3</NetworkAdapterType>
        </NetworkConnection>
        <NetworkConnection needsCustomization="true" network="none">
            <NetworkConnectionIndex>1</NetworkConnectionIndex>
            <IsConnected>false</IsConnected>
			<IpAddress>1.1.1.1</IpAddress>
            <MACAddress>00:50:56:29:01:46</MACAddress>
            <IpAddressAllocationMode>NONE</IpAddressAllocationMode>
            <NetworkAdapterType>VMXNET3</NetworkAdapterType>
        </NetworkConnection>
        <Link rel="edit" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/networkConnectionSection/" type="application/vnd.vmware.vcloud.networkConnectionSection+xml"/>
        <Link rel="edit" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/networkConnectionSection/" type="application/vnd.vmware.vcloud.networkConnectionSection+json"/>
    </NetworkConnectionSection>`

// testGetVmTextWithCustomNetworkSection helps to generate whole VM XML with custom `networkSection` to help with unit
// testing
func testGetVmTextWithCustomNetworkSection(networkSection string) string {
	return `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Vm xmlns="http://www.vmware.com/vcloud/v1.5" 
    xmlns:ovf="http://schemas.dmtf.org/ovf/envelope/1" 
    xmlns:vssd="http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_VirtualSystemSettingData" 
    xmlns:common="http://schemas.dmtf.org/wbem/wscim/1/common" 
    xmlns:rasd="http://schemas.dmtf.org/wbem/wscim/1/cim-schema/2/CIM_ResourceAllocationSettingData" 
    xmlns:vmw="http://www.vmware.com/schema/ovf" 
    xmlns:ovfenv="http://schemas.dmtf.org/ovf/environment/1" 
    xmlns:vmext="http://www.vmware.com/vcloud/extension/v1.5" 
    xmlns:ns9="http://www.vmware.com/vcloud/versions" needsCustomization="true" nestedHypervisorEnabled="false" deployed="false" status="8" name="TestAccVcdVAppVmDhcpWaitVM" id="urn:vcloud:vm:676e755e-766e-48e3-92bd-fabdb322bed9" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9" type="application/vnd.vmware.vcloud.vm+xml">
    <Link rel="power:powerOn" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/power/action/powerOn"/>
    <Link rel="deploy" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/action/deploy" type="application/vnd.vmware.vcloud.deployVAppParams+xml"/>
    <Link rel="deploy" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/action/deploy" type="application/vnd.vmware.vcloud.deployVAppParams+json"/>
    <Link rel="edit" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9" type="application/vnd.vmware.vcloud.vm+xml"/>
    <Link rel="edit" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9" type="application/vnd.vmware.vcloud.vm+json"/>
    <Link rel="remove" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9"/>
    <Link rel="down" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/metadata" type="application/vnd.vmware.vcloud.metadata+xml"/>
    <Link rel="down" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/metadata" type="application/vnd.vmware.vcloud.metadata+json"/>
    <Link rel="down" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/productSections/" type="application/vnd.vmware.vcloud.productSections+xml"/>
    <Link rel="down" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/productSections/" type="application/vnd.vmware.vcloud.productSections+json"/>
    <Link rel="screen:thumbnail" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/screen"/>
    <Link rel="media:insertMedia" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/media/action/insertMedia" type="application/vnd.vmware.vcloud.mediaInsertOrEjectParams+xml"/>
    <Link rel="media:insertMedia" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/media/action/insertMedia" type="application/vnd.vmware.vcloud.mediaInsertOrEjectParams+json"/>
    <Link rel="media:ejectMedia" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/media/action/ejectMedia" type="application/vnd.vmware.vcloud.mediaInsertOrEjectParams+xml"/>
    <Link rel="media:ejectMedia" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/media/action/ejectMedia" type="application/vnd.vmware.vcloud.mediaInsertOrEjectParams+json"/>
    <Link rel="disk:attach" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/disk/action/attach" type="application/vnd.vmware.vcloud.diskAttachOrDetachParams+xml"/>
    <Link rel="disk:attach" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/disk/action/attach" type="application/vnd.vmware.vcloud.diskAttachOrDetachParams+json"/>
    <Link rel="disk:detach" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/disk/action/detach" type="application/vnd.vmware.vcloud.diskAttachOrDetachParams+xml"/>
    <Link rel="disk:detach" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/disk/action/detach" type="application/vnd.vmware.vcloud.diskAttachOrDetachParams+json"/>
    <Link rel="upgrade" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/action/upgradeHardwareVersion"/>
    <Link rel="enable" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/action/enableNestedHypervisor"/>
    <Link rel="customizeAtNextPowerOn" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/action/customizeAtNextPowerOn"/>
    <Link rel="snapshot:create" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/action/createSnapshot" type="application/vnd.vmware.vcloud.createSnapshotParams+xml"/>
    <Link rel="snapshot:create" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/action/createSnapshot" type="application/vnd.vmware.vcloud.createSnapshotParams+json"/>
    <Link rel="reconfigureVm" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/action/reconfigureVm" name="TestAccVcdVAppVmDhcpWaitVM" type="application/vnd.vmware.vcloud.vm+xml"/>
    <Link rel="reconfigureVm" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/action/reconfigureVm" name="TestAccVcdVAppVmDhcpWaitVM" type="application/vnd.vmware.vcloud.vm+json"/>
    <Link rel="up" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vapp-eddd8e06-ab9f-4b5a-8244-fea2b44d7bfe" type="application/vnd.vmware.vcloud.vApp+xml"/>
    <Link rel="up" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vapp-eddd8e06-ab9f-4b5a-8244-fea2b44d7bfe" type="application/vnd.vmware.vcloud.vApp+json"/>
    <Description>This OVA provides a minimal installed profile of PhotonOS.
   Default password for root user is changeme. However user will be prompted to change the password during first login.
    </Description>
    <VmSpecSection>
        <ovf:Info>Virtual hardware requirements (simplified)</ovf:Info>
        <OsType>other3xLinux64Guest</OsType>
        <NumCpus>1</NumCpus>
        <NumCoresPerSocket>1</NumCoresPerSocket>
        <CpuResourceMhz>
            <Configured>1</Configured>
        </CpuResourceMhz>
        <MemoryResourceMb>
            <Configured>512</Configured>
        </MemoryResourceMb>
        <MediaSection>
            <MediaSettings>
                <DeviceId>3002</DeviceId>
                <MediaType>ISO</MediaType>
                <MediaState>DISCONNECTED</MediaState>
                <UnitNumber>0</UnitNumber>
                <BusNumber>1</BusNumber>
                <AdapterType>1</AdapterType>
            </MediaSettings>
            <MediaSettings>
                <DeviceId>8000</DeviceId>
                <MediaType>FLOPPY</MediaType>
                <MediaState>DISCONNECTED</MediaState>
                <UnitNumber>0</UnitNumber>
                <BusNumber>0</BusNumber>
            </MediaSettings>
        </MediaSection>
        <DiskSection>
            <DiskSettings>
                <DiskId>2000</DiskId>
                <SizeMb>16384</SizeMb>
                <UnitNumber>0</UnitNumber>
                <BusNumber>0</BusNumber>
                <AdapterType>5</AdapterType>
                <ThinProvisioned>true</ThinProvisioned>
                <StorageProfile href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vdcStorageProfile/f5c4c2e8-2efc-4cbe-a0ca-39b809595c91" id="urn:vcloud:vdcstorageProfile:f5c4c2e8-2efc-4cbe-a0ca-39b809595c91" name="*" type="application/vnd.vmware.vcloud.vdcStorageProfile+xml"/>
                <overrideVmDefault>false</overrideVmDefault>
                <iops>0</iops>
                <VirtualQuantityUnit>byte</VirtualQuantityUnit>
            </DiskSettings>
        </DiskSection>
        <HardwareVersion href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vdc/16c69e4c-dd0d-45ab-9126-a647c35393ff/hwv/vmx-11" type="application/vnd.vmware.vcloud.virtualHardwareVersion+xml">vmx-11</HardwareVersion>
        <VmToolsVersion>2147483647</VmToolsVersion>
        <VirtualCpuType>VM64</VirtualCpuType>
        <TimeSyncWithHost>false</TimeSyncWithHost>
    </VmSpecSection>
    <ovf:VirtualHardwareSection xmlns:ns10="http://www.vmware.com/vcloud/v1.5" ovf:transport="" ns10:type="application/vnd.vmware.vcloud.virtualHardwareSection+xml" ns10:href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/">
        <ovf:Info>Virtual hardware requirements</ovf:Info>
        <ovf:System>
            <vssd:AutomaticRecoveryAction xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <vssd:AutomaticShutdownAction xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <vssd:AutomaticStartupAction xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <vssd:AutomaticStartupActionDelay xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <vssd:AutomaticStartupActionSequenceNumber xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <vssd:Caption xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <vssd:ChangeableType xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <vssd:ConfigurationDataRoot xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <vssd:ConfigurationFile xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <vssd:ConfigurationID xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <vssd:ConfigurationName xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <vssd:CreationTime xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <vssd:Description xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <vssd:ElementName>Virtual Hardware Family</vssd:ElementName>
            <vssd:Generation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <vssd:InstanceID>0</vssd:InstanceID>
            <vssd:LogDataRoot xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <vssd:RecoveryFile xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <vssd:SnapshotDataRoot xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <vssd:SuspendDataRoot xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <vssd:SwapFileDataRoot xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <vssd:VirtualSystemIdentifier>TestAccVcdVAppVmDhcpWaitVM</vssd:VirtualSystemIdentifier>
            <vssd:VirtualSystemType>vmx-11</vssd:VirtualSystemType>
        </ovf:System>
        <ovf:Item>
            <rasd:Address>00:50:56:29:01:46</rasd:Address>
            <rasd:AddressOnParent>1</rasd:AddressOnParent>
            <rasd:AllocationUnits xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:AutomaticAllocation>false</rasd:AutomaticAllocation>
            <rasd:AutomaticDeallocation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Caption xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ChangeableType xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ConfigurationName xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Connection ns10:ipAddressingMode="NONE" ns10:primaryNetworkConnection="false">none</rasd:Connection>
            <rasd:ConsumerVisibility xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Description>Vmxnet3 ethernet adapter on "none"</rasd:Description>
            <rasd:ElementName>Network adapter 1</rasd:ElementName>
            <rasd:Generation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:InstanceID>1</rasd:InstanceID>
            <rasd:Limit xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:MappingBehavior xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:OtherResourceType xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Parent xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:PoolID xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Reservation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ResourceSubType>VMXNET3</rasd:ResourceSubType>
            <rasd:ResourceType>10</rasd:ResourceType>
            <rasd:VirtualQuantity xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:VirtualQuantityUnits xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Weight xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
        </ovf:Item>
        <ovf:Item>
            <rasd:Address>00:50:56:29:01:45</rasd:Address>
            <rasd:AddressOnParent>0</rasd:AddressOnParent>
            <rasd:AllocationUnits xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:AutomaticAllocation>true</rasd:AutomaticAllocation>
            <rasd:AutomaticDeallocation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Caption xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ChangeableType xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ConfigurationName xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Connection ns10:ipAddressingMode="DHCP" ns10:primaryNetworkConnection="true">multinic-net</rasd:Connection>
            <rasd:ConsumerVisibility xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Description>Vmxnet3 ethernet adapter on "multinic-net"</rasd:Description>
            <rasd:ElementName>Network adapter 0</rasd:ElementName>
            <rasd:Generation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:InstanceID>2</rasd:InstanceID>
            <rasd:Limit xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:MappingBehavior xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:OtherResourceType xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Parent xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:PoolID xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Reservation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ResourceSubType>VMXNET3</rasd:ResourceSubType>
            <rasd:ResourceType>10</rasd:ResourceType>
            <rasd:VirtualQuantity xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:VirtualQuantityUnits xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Weight xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
        </ovf:Item>
        <ovf:Item>
            <rasd:Address>0</rasd:Address>
            <rasd:AddressOnParent xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:AllocationUnits xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:AutomaticAllocation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:AutomaticDeallocation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Caption xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ChangeableType xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ConfigurationName xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ConsumerVisibility xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Description>SCSI Controller</rasd:Description>
            <rasd:ElementName>SCSI Controller 0</rasd:ElementName>
            <rasd:Generation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:InstanceID>3</rasd:InstanceID>
            <rasd:Limit xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:MappingBehavior xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:OtherResourceType xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Parent xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:PoolID xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Reservation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ResourceSubType>VirtualSCSI</rasd:ResourceSubType>
            <rasd:ResourceType>6</rasd:ResourceType>
            <rasd:VirtualQuantity xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:VirtualQuantityUnits xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Weight xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
        </ovf:Item>
        <ovf:Item>
            <rasd:Address xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:AddressOnParent>0</rasd:AddressOnParent>
            <rasd:AllocationUnits xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:AutomaticAllocation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:AutomaticDeallocation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Caption xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ChangeableType xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ConfigurationName xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ConsumerVisibility xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Description>Hard disk</rasd:Description>
            <rasd:ElementName>Hard disk 1</rasd:ElementName>
            <rasd:Generation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:HostResource ns10:storageProfileHref="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vdcStorageProfile/f5c4c2e8-2efc-4cbe-a0ca-39b809595c91" ns10:busType="6" ns10:busSubType="VirtualSCSI" ns10:capacity="16384" ns10:iops="0" ns10:storageProfileOverrideVmDefault="false"></rasd:HostResource>
            <rasd:InstanceID>2000</rasd:InstanceID>
            <rasd:Limit xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:MappingBehavior xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:OtherResourceType xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Parent>3</rasd:Parent>
            <rasd:PoolID xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Reservation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ResourceSubType xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ResourceType>17</rasd:ResourceType>
            <rasd:VirtualQuantity>17179869184</rasd:VirtualQuantity>
            <rasd:VirtualQuantityUnits>byte</rasd:VirtualQuantityUnits>
            <rasd:Weight xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
        </ovf:Item>
        <ovf:Item>
            <rasd:Address>1</rasd:Address>
            <rasd:AddressOnParent xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:AllocationUnits xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:AutomaticAllocation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:AutomaticDeallocation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Caption xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ChangeableType xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ConfigurationName xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ConsumerVisibility xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Description>IDE Controller</rasd:Description>
            <rasd:ElementName>IDE Controller 1</rasd:ElementName>
            <rasd:Generation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:InstanceID>4</rasd:InstanceID>
            <rasd:Limit xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:MappingBehavior xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:OtherResourceType xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Parent xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:PoolID xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Reservation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ResourceSubType xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ResourceType>5</rasd:ResourceType>
            <rasd:VirtualQuantity xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:VirtualQuantityUnits xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Weight xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
        </ovf:Item>
        <ovf:Item>
            <rasd:Address xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:AddressOnParent>0</rasd:AddressOnParent>
            <rasd:AllocationUnits xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:AutomaticAllocation>false</rasd:AutomaticAllocation>
            <rasd:AutomaticDeallocation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Caption xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ChangeableType xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ConfigurationName xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ConsumerVisibility xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Description>CD/DVD Drive</rasd:Description>
            <rasd:ElementName>CD/DVD Drive 1</rasd:ElementName>
            <rasd:Generation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:HostResource></rasd:HostResource>
            <rasd:InstanceID>3002</rasd:InstanceID>
            <rasd:Limit xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:MappingBehavior xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:OtherResourceType xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Parent>4</rasd:Parent>
            <rasd:PoolID xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Reservation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ResourceSubType xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ResourceType>15</rasd:ResourceType>
            <rasd:VirtualQuantity xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:VirtualQuantityUnits xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Weight xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
        </ovf:Item>
        <ovf:Item>
            <rasd:Address xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:AddressOnParent>0</rasd:AddressOnParent>
            <rasd:AllocationUnits xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:AutomaticAllocation>false</rasd:AutomaticAllocation>
            <rasd:AutomaticDeallocation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Caption xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ChangeableType xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ConfigurationName xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ConsumerVisibility xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Description>Floppy Drive</rasd:Description>
            <rasd:ElementName>Floppy Drive 1</rasd:ElementName>
            <rasd:Generation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:HostResource></rasd:HostResource>
            <rasd:InstanceID>8000</rasd:InstanceID>
            <rasd:Limit xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:MappingBehavior xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:OtherResourceType xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Parent xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:PoolID xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Reservation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ResourceSubType xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ResourceType>14</rasd:ResourceType>
            <rasd:VirtualQuantity xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:VirtualQuantityUnits xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Weight xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
        </ovf:Item>
        <ovf:Item ns10:type="application/vnd.vmware.vcloud.rasdItem+xml" ns10:href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/cpu">
            <rasd:Address xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:AddressOnParent xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:AllocationUnits>hertz * 10^6</rasd:AllocationUnits>
            <rasd:AutomaticAllocation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:AutomaticDeallocation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Caption xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ChangeableType xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ConfigurationName xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ConsumerVisibility xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Description>Number of Virtual CPUs</rasd:Description>
            <rasd:ElementName>1 virtual CPU(s)</rasd:ElementName>
            <rasd:Generation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:InstanceID>5</rasd:InstanceID>
            <rasd:Limit xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:MappingBehavior xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:OtherResourceType xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Parent xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:PoolID xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Reservation>0</rasd:Reservation>
            <rasd:ResourceSubType xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ResourceType>3</rasd:ResourceType>
            <rasd:VirtualQuantity>1</rasd:VirtualQuantity>
            <rasd:VirtualQuantityUnits xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Weight>0</rasd:Weight>
            <vmw:CoresPerSocket ovf:required="false">1</vmw:CoresPerSocket>
            <Link rel="edit" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/cpu" type="application/vnd.vmware.vcloud.rasdItem+xml"/>
            <Link rel="edit" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/cpu" type="application/vnd.vmware.vcloud.rasdItem+json"/>
        </ovf:Item>
        <ovf:Item ns10:type="application/vnd.vmware.vcloud.rasdItem+xml" ns10:href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/memory">
            <rasd:Address xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:AddressOnParent xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:AllocationUnits>byte * 2^20</rasd:AllocationUnits>
            <rasd:AutomaticAllocation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:AutomaticDeallocation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Caption xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ChangeableType xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ConfigurationName xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ConsumerVisibility xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Description>Memory Size</rasd:Description>
            <rasd:ElementName>512 MB of memory</rasd:ElementName>
            <rasd:Generation xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:InstanceID>6</rasd:InstanceID>
            <rasd:Limit xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:MappingBehavior xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:OtherResourceType xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Parent xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:PoolID xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Reservation>0</rasd:Reservation>
            <rasd:ResourceSubType xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:ResourceType>4</rasd:ResourceType>
            <rasd:VirtualQuantity>512</rasd:VirtualQuantity>
            <rasd:VirtualQuantityUnits xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"/>
            <rasd:Weight>0</rasd:Weight>
            <Link rel="edit" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/memory" type="application/vnd.vmware.vcloud.rasdItem+xml"/>
            <Link rel="edit" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/memory" type="application/vnd.vmware.vcloud.rasdItem+json"/>
        </ovf:Item>
        <Link rel="edit" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/" type="application/vnd.vmware.vcloud.virtualHardwareSection+xml"/>
        <Link rel="edit" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/" type="application/vnd.vmware.vcloud.virtualHardwareSection+json"/>
        <Link rel="down" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/cpu" type="application/vnd.vmware.vcloud.rasdItem+xml"/>
        <Link rel="down" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/cpu" type="application/vnd.vmware.vcloud.rasdItem+json"/>
        <Link rel="edit" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/cpu" type="application/vnd.vmware.vcloud.rasdItem+xml"/>
        <Link rel="edit" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/cpu" type="application/vnd.vmware.vcloud.rasdItem+json"/>
        <Link rel="down" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/memory" type="application/vnd.vmware.vcloud.rasdItem+xml"/>
        <Link rel="down" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/memory" type="application/vnd.vmware.vcloud.rasdItem+json"/>
        <Link rel="edit" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/memory" type="application/vnd.vmware.vcloud.rasdItem+xml"/>
        <Link rel="edit" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/memory" type="application/vnd.vmware.vcloud.rasdItem+json"/>
        <Link rel="down" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/disks" type="application/vnd.vmware.vcloud.rasdItemsList+xml"/>
        <Link rel="down" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/disks" type="application/vnd.vmware.vcloud.rasdItemsList+json"/>
        <Link rel="edit" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/disks" type="application/vnd.vmware.vcloud.rasdItemsList+xml"/>
        <Link rel="edit" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/disks" type="application/vnd.vmware.vcloud.rasdItemsList+json"/>
        <Link rel="down" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/media" type="application/vnd.vmware.vcloud.rasdItemsList+xml"/>
        <Link rel="down" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/media" type="application/vnd.vmware.vcloud.rasdItemsList+json"/>
        <Link rel="down" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/networkCards" type="application/vnd.vmware.vcloud.rasdItemsList+xml"/>
        <Link rel="down" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/networkCards" type="application/vnd.vmware.vcloud.rasdItemsList+json"/>
        <Link rel="edit" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/networkCards" type="application/vnd.vmware.vcloud.rasdItemsList+xml"/>
        <Link rel="edit" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/networkCards" type="application/vnd.vmware.vcloud.rasdItemsList+json"/>
        <Link rel="down" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/serialPorts" type="application/vnd.vmware.vcloud.rasdItemsList+xml"/>
        <Link rel="down" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/serialPorts" type="application/vnd.vmware.vcloud.rasdItemsList+json"/>
        <Link rel="edit" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/serialPorts" type="application/vnd.vmware.vcloud.rasdItemsList+xml"/>
        <Link rel="edit" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/virtualHardwareSection/serialPorts" type="application/vnd.vmware.vcloud.rasdItemsList+json"/>
    </ovf:VirtualHardwareSection>
    <ovf:OperatingSystemSection xmlns:ns10="http://www.vmware.com/vcloud/v1.5" ovf:id="102" ns10:type="application/vnd.vmware.vcloud.operatingSystemSection+xml" vmw:osType="other3xLinux64Guest" ns10:href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/operatingSystemSection/">
        <ovf:Info>Specifies the operating system installed</ovf:Info>
        <ovf:Description>Other 3.x Linux (64-bit)</ovf:Description>
        <Link rel="edit" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/operatingSystemSection/" type="application/vnd.vmware.vcloud.operatingSystemSection+xml"/>
        <Link rel="edit" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/operatingSystemSection/" type="application/vnd.vmware.vcloud.operatingSystemSection+json"/>
    </ovf:OperatingSystemSection>
` + networkSection + `
    <GuestCustomizationSection href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/guestCustomizationSection/" type="application/vnd.vmware.vcloud.guestCustomizationSection+xml" ovf:required="false">
        <ovf:Info>Specifies Guest OS Customization Settings</ovf:Info>
        <Enabled>true</Enabled>
        <ChangeSid>false</ChangeSid>
        <VirtualMachineId>676e755e-766e-48e3-92bd-fabdb322bed9</VirtualMachineId>
        <JoinDomainEnabled>false</JoinDomainEnabled>
        <UseOrgSettings>false</UseOrgSettings>
        <AdminPasswordEnabled>true</AdminPasswordEnabled>
        <AdminPasswordAuto>true</AdminPasswordAuto>
        <AdminAutoLogonEnabled>false</AdminAutoLogonEnabled>
        <AdminAutoLogonCount>0</AdminAutoLogonCount>
        <ResetPasswordRequired>false</ResetPasswordRequired>
        <ComputerName>dhcp-vm</ComputerName>
        <Link rel="edit" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/guestCustomizationSection/" type="application/vnd.vmware.vcloud.guestCustomizationSection+xml"/>
        <Link rel="edit" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/guestCustomizationSection/" type="application/vnd.vmware.vcloud.guestCustomizationSection+json"/>
        <Link rel="down" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/guestcustomizationstatus/" type="application/vnd.vmware.vcloud.guestCustomizationStatusSection+xml"/>
        <Link rel="down" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/guestcustomizationstatus/" type="application/vnd.vmware.vcloud.guestCustomizationStatusSection+json"/>
        <Link rel="down" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/checkpostcustomizationscript/" type="application/vnd.vmware.vcloud.vm.checkPostGuestCustomizationSection+xml"/>
        <Link rel="down" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/checkpostcustomizationscript/" type="application/vnd.vmware.vcloud.vm.checkPostGuestCustomizationSection+json"/>
    </GuestCustomizationSection>
    <ovf:ProductSection ovf:class="" ovf:instance="" ovf:required="true">
        <ovf:Info>Information about the installed software</ovf:Info>
        <ovf:Product>Photon OS</ovf:Product>
        <ovf:Vendor>VMware Inc.</ovf:Vendor>
        <ovf:BuildVersion>3.0</ovf:BuildVersion>
        <ovf:FullVersion>3.0</ovf:FullVersion>
    </ovf:ProductSection>
    <RuntimeInfoSection xmlns:ns10="http://www.vmware.com/vcloud/v1.5" ns10:type="application/vnd.vmware.vcloud.virtualHardwareSection+xml" ns10:href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/runtimeInfoSection">
        <ovf:Info>Specifies Runtime info</ovf:Info>
        <VMWareTools version="2147483647"/>
    </RuntimeInfoSection>
    <SnapshotSection href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/snapshotSection" type="application/vnd.vmware.vcloud.snapshotSection+xml" ovf:required="false">
        <ovf:Info>Snapshot information section</ovf:Info>
    </SnapshotSection>
    <DateCreated>2020-03-10T22:39:10.832Z</DateCreated>
    <VAppScopedLocalId>9eee4c0a-6bc3-4099-bf8c-c83c8457f377</VAppScopedLocalId>
    <VmCapabilities href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/vmCapabilities/" type="application/vnd.vmware.vcloud.vmCapabilitiesSection+xml">
        <Link rel="edit" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/vmCapabilities/" type="application/vnd.vmware.vcloud.vmCapabilitiesSection+xml"/>
        <Link rel="edit" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/vmCapabilities/" type="application/vnd.vmware.vcloud.vmCapabilitiesSection+json"/>
        <MemoryHotAddEnabled>false</MemoryHotAddEnabled>
        <CpuHotAddEnabled>false</CpuHotAddEnabled>
    </VmCapabilities>
    <StorageProfile href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vdcStorageProfile/f5c4c2e8-2efc-4cbe-a0ca-39b809595c91" id="urn:vcloud:vdcstorageProfile:f5c4c2e8-2efc-4cbe-a0ca-39b809595c91" name="*" type="application/vnd.vmware.vcloud.vdcStorageProfile+xml"/>
    <VdcComputePolicy href="https://bos1-vcloud-static-170-211.eng.vmware.com/cloudapi/1.0.0/vdcComputePolicies/urn:vcloud:vdcComputePolicy:e811fe11-fee7-404a-ae98-8441c5ce6b4b" id="urn:vcloud:vdcComputePolicy:e811fe11-fee7-404a-ae98-8441c5ce6b4b" name="System Default" type="application/json"/>
    <BootOptions href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/bootOptions/" type="application/vnd.vmware.vcloud.bootOptionsSection+xml">
        <Link rel="edit" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/action/bootOptions" type="application/vnd.vmware.vcloud.bootOptionsSection+xml"/>
        <Link rel="edit" href="https://bos1-vcloud-static-170-211.eng.vmware.com/api/vApp/vm-676e755e-766e-48e3-92bd-fabdb322bed9/action/bootOptions" type="application/vnd.vmware.vcloud.bootOptionsSection+json"/>
        <BootDelay>0</BootDelay>
        <EnterBIOSSetup>false</EnterBIOSSetup>
    </BootOptions>
</Vm>`
}
