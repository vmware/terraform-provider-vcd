// +build unit ALL

package vcd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	semver "github.com/hashicorp/go-version"
)

// Checks that the provider header in index.html.markdown
// has the version defined in the VERSION file
func TestProviderVersion(t *testing.T) {
	indexFile := path.Join(getCurrentDir(), "..", "website", "docs", "index.html.markdown")
	_, err := os.Stat(indexFile)
	if os.IsNotExist(err) {
		fmt.Printf("%s\n", indexFile)
		panic("Could not find index.html.markdown file")
	}

	indexText, err := ioutil.ReadFile(indexFile)
	if err != nil {
		panic(fmt.Errorf("could not read index file %s: %v", indexFile, err))
	}

	vcdHeader := `# VMware vCloud Director Provider`
	expectedText := vcdHeader + ` ` + currentProviderVersion
	reExpectedVersion := regexp.MustCompile(`(?m)^` + expectedText)
	reFoundVersion := regexp.MustCompile(`(?m)^` + vcdHeader + ` \d+\.\d+`)
	if reExpectedVersion.MatchString(string(indexText)) {
		if vcdTestVerbose {
			t.Logf("Found expected version <%s> in index.html.markdown", currentProviderVersion)
		}
	} else {
		foundList := reFoundVersion.FindAllStringSubmatch(string(indexText), -1)
		foundText := ""
		if len(foundList) > 0 && len(foundList[0]) > 0 {
			foundText = foundList[0][0]
			t.Logf("Expected text: <%s>", expectedText)
			t.Logf("Found text   : <%s> in index.html.markdown", foundText)
		} else {
			t.Logf("No version found in index.html.markdown")
		}
		t.Fail()
	}
}

// Checks that a PREVIOUS_VERSION file exists, and it contains a version lower than the one in VERSION
func TestProviderUpgradeVersion(t *testing.T) {
	currentVersionText, err := getVersionFromFile("VERSION")
	if err != nil {
		t.Logf("error retrieving version from VERSION file: %s", err)
		t.Fail()
		return
	}
	previousVersionText, err := getVersionFromFile("PREVIOUS_VERSION")
	if err != nil {
		t.Logf("error retrieving version from PREVIOUS_VERSION file: %s", err)
		t.Fail()
		return
	}

	currentVersion, err := semver.NewVersion(currentVersionText)
	if err != nil {
		t.Logf("error converting current version to Hashicorp version: %s", err)
		t.Fail()
		return
	}
	previousVersion, err := semver.NewVersion(previousVersionText)
	if err != nil {
		t.Logf("error converting previous version to Hashicorp version: %s", err)
		t.Fail()
		return
	}
	result := currentVersion.Compare(previousVersion)
	// result < 0 means current version is lower than previous version
	// result == 0 means current version is the same as previous version
	// result == 1 means current version is higher than previous version
	if result < 0 {
		t.Logf("current version (%s) is lower than previous version (%s)", currentVersionText, previousVersionText)
		t.Fail()
	}
	if result == 0 {
		t.Logf("current version (%s) is the same as previous version (%s)", currentVersionText, previousVersionText)
		t.Fail()
	}
}

func TestGetMajorVersion(t *testing.T) {
	version := getMajorVersion()

	reVersion := regexp.MustCompile(`^\d+\.\d+$`)
	if !reVersion.MatchString(version) {
		t.Fail()
	}
	t.Logf("%s", version)
}

func init() {
	testingTags["unit"] = "provider_unit_test.go"
}

func TestVcdResources(t *testing.T) {
	type args struct {
		nameRegexp        string
		includeDeprecated bool
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]*schema.Resource
		wantLen int
		lenOnly bool // whether to ignore actual 'want' value if 'len' is ok
		wantErr bool
	}{
		{
			name:    "GetAllResources",
			args:    args{nameRegexp: "", includeDeprecated: true},
			want:    globalResourceMap,
			wantLen: len(Provider().Resources()),
			wantErr: false,
		},
		{
			name:    "MatchExactResourceName",
			args:    args{nameRegexp: "vcd_vapp_vm", includeDeprecated: false},
			wantLen: 1, // should return only one because exact name was given
			lenOnly: true,
			wantErr: false,
		},
		{
			name:    "MatchNoResources",
			args:    args{nameRegexp: "NonExistingName", includeDeprecated: false},
			want:    make(map[string]*schema.Resource),
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "InvalidRegexpError",
			args:    args{nameRegexp: "[0-9]++", includeDeprecated: false},
			want:    nil,
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Resources(tt.args.nameRegexp, tt.args.includeDeprecated)
			if (err != nil) != tt.wantErr {
				t.Errorf("Resources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != tt.wantLen {
				t.Errorf("Resources() returned = %d elements, want %d", len(got), tt.wantLen)
			}

			if !tt.lenOnly && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Resources() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVcdDataSources(t *testing.T) {
	type args struct {
		nameRegexp        string
		includeDeprecated bool
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]*schema.Resource
		wantLen int
		lenOnly bool // whether to ignore actual 'want' value if 'len' is ok
		wantErr bool
	}{
		{
			name:    "GetAllDataSources",
			args:    args{nameRegexp: "", includeDeprecated: true},
			want:    globalDataSourceMap,
			wantLen: len(Provider().DataSources()),
			wantErr: false,
		},
		{
			name:    "MatchExactDataSourceName",
			args:    args{nameRegexp: "vcd_vapp_vm", includeDeprecated: false},
			wantLen: 1, // should return only one because exact name was given
			lenOnly: true,
			wantErr: false,
		},
		{
			name:    "MatchNoDataSources",
			args:    args{nameRegexp: "NonExistingName", includeDeprecated: false},
			want:    make(map[string]*schema.Resource),
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "InvalidRegexpError",
			args:    args{nameRegexp: "[0-9]++", includeDeprecated: false},
			want:    nil,
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DataSources(tt.args.nameRegexp, tt.args.includeDeprecated)
			if (err != nil) != tt.wantErr {
				t.Errorf("Resources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != tt.wantLen {
				t.Errorf("Resources() returned = %d elements, want %d", len(got), tt.wantLen)
			}

			if !tt.lenOnly && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Resources() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVcdSchemaFilter(t *testing.T) {

	fakeSchema := make(map[string]*schema.Resource)
	terraformObject := schema.Resource{}
	deprecatedTerraformObject := schema.Resource{DeprecationMessage: "Deprecated"}
	fakeSchema["resource_one"] = &terraformObject
	fakeSchema["resource_two"] = &terraformObject
	fakeSchema["resource_three"] = &deprecatedTerraformObject

	type args struct {
		nameRegexp        string
		includeDeprecated bool
	}
	tests := []struct {
		name    string
		args    args
		want    map[string]*schema.Resource
		wantLen int
		lenOnly bool // whether to ignore actual 'want' value if 'len' is ok
		wantErr bool
	}{
		{
			name:    "GetAllResources",
			args:    args{nameRegexp: "", includeDeprecated: true},
			want:    fakeSchema,
			wantLen: len(fakeSchema),
			wantErr: false,
		},
		{
			name:    "MatchExactDataSourceName",
			args:    args{nameRegexp: "resource_two", includeDeprecated: false},
			wantLen: 1, // should return only one because exact name was given
			lenOnly: true,
			wantErr: false,
		},
		{
			name:    "MatchNoDataSources",
			args:    args{nameRegexp: "NonExistingName", includeDeprecated: false},
			want:    make(map[string]*schema.Resource),
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "OnlyNonDeprecated",
			args:    args{nameRegexp: "", includeDeprecated: false},
			want:    nil,
			wantLen: 2,
			lenOnly: true,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := vcdSchemaFilter(fakeSchema, tt.args.nameRegexp, tt.args.includeDeprecated)
			if (err != nil) != tt.wantErr {
				t.Errorf("Resources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != tt.wantLen {
				t.Errorf("Resources() returned = %d elements, want %d", len(got), tt.wantLen)
			}

			if !tt.lenOnly && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Resources() got = %v, want %v", got, tt.want)
			}
		})
	}
}
