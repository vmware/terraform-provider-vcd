//go:build unit || ALL

package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"reflect"
	"testing"
)

// Test_getTkgVersionBundleFromVAppTemplateName is a unit test that tests the getTkgVersionBundleFromVAppTemplateName function
func Test_getTkgVersionBundleFromVAppTemplateName(t *testing.T) {
	tests := []struct {
		name    string
		ovaName string
		want    tkgVersionBundle
		wantErr string
	}{
		{
			name:    "wrong ova name",
			ovaName: "randomOVA",
			want:    tkgVersionBundle{},
			wantErr: "the vApp Template 'randomOVA' is not a Kubernetes template OVA",
		},
		{
			name:    "not supported ova",
			ovaName: "ubuntu-2004-kube-v9.99.9+vmware.9-tkg.9-b8c57a6c8c98d227f74e7b1a9eef27st",
			want:    tkgVersionBundle{},
			wantErr: "the Kubernetes OVA 'v9.99.9+vmware.9-tkg.9-b8c57a6c8c98d227f74e7b1a9eef27st' is not supported",
		},
		{
			name:    "not supported photon ova",
			ovaName: "photon-3-kube-v1.27.5+vmware.1-tkg.1-cac282289bb29b217b808a2b9b0c0c46",
			want:    tkgVersionBundle{},
			wantErr: "the vApp Template 'photon-3-kube-v1.27.5+vmware.1-tkg.1-cac282289bb29b217b808a2b9b0c0c46' uses Photon, and it is not supported",
		},
		{
			name:    "supported ova",
			ovaName: "ubuntu-2004-kube-v1.26.8+vmware.1-tkg.1-0edd4dafbefbdb503f64d5472e500cf8",
			want: tkgVersionBundle{
				EtcdVersion:       "v3.5.6_vmware.20",
				CoreDnsVersion:    "v1.9.3_vmware.16",
				TkgVersion:        "v2.3.1",
				TkrVersion:        "v1.26.8---vmware.1-tkg.1",
				KubernetesVersion: "v1.26.8+vmware.1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getTkgVersionBundleFromVAppTemplateName(tt.ovaName)
			if err != nil {
				if tt.wantErr == "" {
					t.Fatalf("getTkgVersionBundleFromVAppTemplateName() got error = %v, but should have not failed", err)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("getTkgVersionBundleFromVAppTemplateName() error = %v, wantErr = %v", err, tt.wantErr)
				}
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("getTkgVersionBundleFromVAppTemplateName() got = %v, want = %v", got, tt.want)
			}
		})
	}
}

// Test_generateCapiYaml tests generateCapiYaml function
func Test_generateCapiYaml(t *testing.T) {
	type args struct {
		resourceData   map[string]interface{}
		clusterDetails *createClusterDto
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "foo",
			args: args{
				resourceData: map[string]interface{}{},
				clusterDetails: &createClusterDto{
					Name:            "",
					VcdUrl:          "",
					Org:             nil,
					VdcName:         "",
					OvaName:         "",
					CatalogName:     "",
					NetworkName:     "",
					RdeType:         nil,
					UrnToNamesCache: nil,
					VCDKEConfig: struct {
						MaxUnhealthyNodesPercentage string
						NodeStartupTimeout          string
						NodeNotReadyTimeout         string
						NodeUnknownTimeout          string
						ContainerRegistryUrl        string
					}{},
					TkgVersion: tkgVersionBundle{},
					Owner:      "",
					ApiToken:   "",
				},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceVcdCseKubernetesClusterSchema, tt.args.resourceData)
			got, err := generateCapiYaml(d, tt.args.clusterDetails)
			if (err != nil) != tt.wantErr {
				t.Errorf("generateCapiYaml() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("generateCapiYaml() got = %v, want %v", got, tt.want)
			}
		})
	}
}
