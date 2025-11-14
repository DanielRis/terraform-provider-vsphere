// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/testhelper"
)

// TestAccDataSourceVSphereOvfVMTemplate_fileBasic tests the original file-based functionality
func TestAccDataSourceVSphereOvfVMTemplate_fileBasic(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereOvfVMTemplateConfigFile(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vsphere_ovf_vm_template.template", "id"),
					resource.TestCheckResourceAttrSet("data.vsphere_ovf_vm_template.template", "num_cpus"),
					resource.TestCheckResourceAttrSet("data.vsphere_ovf_vm_template.template", "memory"),
					resource.TestCheckResourceAttrSet("data.vsphere_ovf_vm_template.template", "guest_id"),
					resource.TestCheckResourceAttrSet("data.vsphere_ovf_vm_template.template", "scsi_type"),
				),
			},
		},
	})
}

// TestAccDataSourceVSphereOvfVMTemplate_libraryExactName tests content library search with exact name
func TestAccDataSourceVSphereOvfVMTemplate_libraryExactName(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereOvfVMTemplateConfigLibraryExactName(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.vsphere_ovf_vm_template.template", "name", "test-template"),
					resource.TestCheckResourceAttrSet("data.vsphere_ovf_vm_template.template", "id"),
					resource.TestCheckResourceAttrSet("data.vsphere_ovf_vm_template.template", "num_cpus"),
					resource.TestCheckResourceAttrSet("data.vsphere_ovf_vm_template.template", "memory"),
				),
			},
		},
	})
}

// TestAccDataSourceVSphereOvfVMTemplate_libraryNameRegex tests content library search with regex
func TestAccDataSourceVSphereOvfVMTemplate_libraryNameRegex(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereOvfVMTemplateConfigLibraryRegex(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("data.vsphere_ovf_vm_template.template", "name", regexp.MustCompile("^test-.*")),
					resource.TestCheckResourceAttrSet("data.vsphere_ovf_vm_template.template", "id"),
					resource.TestCheckResourceAttrSet("data.vsphere_ovf_vm_template.template", "num_cpus"),
				),
			},
		},
	})
}

// TestAccDataSourceVSphereOvfVMTemplate_libraryMostRecent tests most_recent flag
func TestAccDataSourceVSphereOvfVMTemplate_libraryMostRecent(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceVSphereOvfVMTemplateConfigLibraryMostRecent(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vsphere_ovf_vm_template.template", "id"),
					resource.TestCheckResourceAttrSet("data.vsphere_ovf_vm_template.template", "name"),
					resource.TestCheckResourceAttrSet("data.vsphere_ovf_vm_template.template", "num_cpus"),
				),
			},
		},
	})
}

// TestAccDataSourceVSphereOvfVMTemplate_libraryMultipleMatchError tests error when multiple matches without most_recent
func TestAccDataSourceVSphereOvfVMTemplate_libraryMultipleMatchError(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceVSphereOvfVMTemplateConfigLibraryMultipleMatch(),
				ExpectError: regexp.MustCompile("multiple OVF templates match the specified filters"),
			},
		},
	})
}

// TestAccDataSourceVSphereOvfVMTemplate_libraryNoMatch tests error when no templates match
func TestAccDataSourceVSphereOvfVMTemplate_libraryNoMatch(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceVSphereOvfVMTemplateConfigLibraryNoMatch(),
				ExpectError: regexp.MustCompile("no OVF templates match the specified filters"),
			},
		},
	})
}

// TestAccDataSourceVSphereOvfVMTemplate_libraryBothNameError tests error when both name and name_regex specified
func TestAccDataSourceVSphereOvfVMTemplate_libraryBothNameError(t *testing.T) {
	testAccSkipUnstable(t)
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			RunSweepers()
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceVSphereOvfVMTemplateConfigLibraryBothName(),
				ExpectError: regexp.MustCompile("cannot specify both 'name' and 'name_regex'"),
			},
		},
	})
}

// Configuration functions

func testAccDataSourceVSphereOvfVMTemplateConfigFile() string {
	return fmt.Sprintf(`
%s

variable "file" {
  type    = string
  default = "%s"
}

data "vsphere_ovf_vm_template" "template" {
  name              = "test-ovf-template"
  resource_pool_id  = data.vsphere_resource_pool.pool1.id
  datastore_id      = vsphere_nas_datastore.ds1.id
  host_system_id    = data.vsphere_host.roothost1.id
  remote_ovf_url    = var.file
  disk_provisioning = "thin"
}
`, testhelper.CombineConfigs(
		testhelper.ConfigDataRootDC1(),
		testhelper.ConfigDataRootHost1(),
		testhelper.ConfigResDS1(),
		testhelper.ConfigDataRootComputeCluster1(),
		testhelper.ConfigResResourcePool1(),
	),
		testhelper.ContentLibraryFiles,
	)
}

func testAccDataSourceVSphereOvfVMTemplateConfigLibraryExactName() string {
	return fmt.Sprintf(`
%s

variable "file" {
  type    = string
  default = "%s"
}

data "vsphere_datastore" "ds" {
  datacenter_id = data.vsphere_datacenter.rootdc1.id
  name          = vsphere_nas_datastore.ds1.name
}

resource "vsphere_content_library" "library" {
  name            = "test-content-library"
  storage_backing = [data.vsphere_datastore.ds.id]
  description     = "Test library for OVF templates"
}

resource "vsphere_content_library_item" "item" {
  name        = "test-template"
  description = "Test OVF Template"
  library_id  = vsphere_content_library.library.id
  type        = "ovf"
  file_url    = var.file
}

data "vsphere_ovf_vm_template" "template" {
  name             = vsphere_content_library_item.item.name
  library_id       = vsphere_content_library.library.id
  resource_pool_id = data.vsphere_resource_pool.pool1.id
  host_system_id   = data.vsphere_host.roothost1.id
  datastore_id     = data.vsphere_datastore.ds.id
}
`, testhelper.CombineConfigs(
		testhelper.ConfigDataRootDC1(),
		testhelper.ConfigDataRootHost1(),
		testhelper.ConfigDataRootHost2(),
		testhelper.ConfigResDS1(),
		testhelper.ConfigDataRootComputeCluster1(),
		testhelper.ConfigResResourcePool1(),
	),
		testhelper.ContentLibraryFiles,
	)
}

func testAccDataSourceVSphereOvfVMTemplateConfigLibraryRegex() string {
	return fmt.Sprintf(`
%s

variable "file" {
  type    = string
  default = "%s"
}

data "vsphere_datastore" "ds" {
  datacenter_id = data.vsphere_datacenter.rootdc1.id
  name          = vsphere_nas_datastore.ds1.name
}

resource "vsphere_content_library" "library" {
  name            = "test-content-library-regex"
  storage_backing = [data.vsphere_datastore.ds.id]
  description     = "Test library for OVF templates"
}

resource "vsphere_content_library_item" "item" {
  name        = "test-template-v1"
  description = "Test OVF Template"
  library_id  = vsphere_content_library.library.id
  type        = "ovf"
  file_url    = var.file
}

data "vsphere_ovf_vm_template" "template" {
  name_regex       = "^test-.*"
  library_id       = vsphere_content_library.library.id
  resource_pool_id = data.vsphere_resource_pool.pool1.id
  host_system_id   = data.vsphere_host.roothost1.id
  datastore_id     = data.vsphere_datastore.ds.id
  most_recent      = true
}
`, testhelper.CombineConfigs(
		testhelper.ConfigDataRootDC1(),
		testhelper.ConfigDataRootHost1(),
		testhelper.ConfigDataRootHost2(),
		testhelper.ConfigResDS1(),
		testhelper.ConfigDataRootComputeCluster1(),
		testhelper.ConfigResResourcePool1(),
	),
		testhelper.ContentLibraryFiles,
	)
}

func testAccDataSourceVSphereOvfVMTemplateConfigLibraryMostRecent() string {
	return fmt.Sprintf(`
%s

variable "file" {
  type    = string
  default = "%s"
}

data "vsphere_datastore" "ds" {
  datacenter_id = data.vsphere_datacenter.rootdc1.id
  name          = vsphere_nas_datastore.ds1.name
}

resource "vsphere_content_library" "library" {
  name            = "test-content-library-recent"
  storage_backing = [data.vsphere_datastore.ds.id]
  description     = "Test library for OVF templates"
}

resource "vsphere_content_library_item" "item1" {
  name        = "test-template-v1"
  description = "Test OVF Template v1"
  library_id  = vsphere_content_library.library.id
  type        = "ovf"
  file_url    = var.file
}

resource "vsphere_content_library_item" "item2" {
  name        = "test-template-v2"
  description = "Test OVF Template v2"
  library_id  = vsphere_content_library.library.id
  type        = "ovf"
  file_url    = var.file
  depends_on  = [vsphere_content_library_item.item1]
}

data "vsphere_ovf_vm_template" "template" {
  name_regex       = "^test-template-.*"
  library_id       = vsphere_content_library.library.id
  most_recent      = true
  resource_pool_id = data.vsphere_resource_pool.pool1.id
  host_system_id   = data.vsphere_host.roothost1.id
  datastore_id     = data.vsphere_datastore.ds.id
  depends_on       = [vsphere_content_library_item.item2]
}
`, testhelper.CombineConfigs(
		testhelper.ConfigDataRootDC1(),
		testhelper.ConfigDataRootHost1(),
		testhelper.ConfigDataRootHost2(),
		testhelper.ConfigResDS1(),
		testhelper.ConfigDataRootComputeCluster1(),
		testhelper.ConfigResResourcePool1(),
	),
		testhelper.ContentLibraryFiles,
	)
}

func testAccDataSourceVSphereOvfVMTemplateConfigLibraryMultipleMatch() string {
	return fmt.Sprintf(`
%s

variable "file" {
  type    = string
  default = "%s"
}

data "vsphere_datastore" "ds" {
  datacenter_id = data.vsphere_datacenter.rootdc1.id
  name          = vsphere_nas_datastore.ds1.name
}

resource "vsphere_content_library" "library" {
  name            = "test-content-library-multiple"
  storage_backing = [data.vsphere_datastore.ds.id]
  description     = "Test library for OVF templates"
}

resource "vsphere_content_library_item" "item1" {
  name        = "test-template-v1"
  description = "Test OVF Template v1"
  library_id  = vsphere_content_library.library.id
  type        = "ovf"
  file_url    = var.file
}

resource "vsphere_content_library_item" "item2" {
  name        = "test-template-v2"
  description = "Test OVF Template v2"
  library_id  = vsphere_content_library.library.id
  type        = "ovf"
  file_url    = var.file
}

data "vsphere_ovf_vm_template" "template" {
  name_regex       = "^test-template-.*"
  library_id       = vsphere_content_library.library.id
  most_recent      = false
  resource_pool_id = data.vsphere_resource_pool.pool1.id
  host_system_id   = data.vsphere_host.roothost1.id
  datastore_id     = data.vsphere_datastore.ds.id
  depends_on       = [vsphere_content_library_item.item1, vsphere_content_library_item.item2]
}
`, testhelper.CombineConfigs(
		testhelper.ConfigDataRootDC1(),
		testhelper.ConfigDataRootHost1(),
		testhelper.ConfigDataRootHost2(),
		testhelper.ConfigResDS1(),
		testhelper.ConfigDataRootComputeCluster1(),
		testhelper.ConfigResResourcePool1(),
	),
		testhelper.ContentLibraryFiles,
	)
}

func testAccDataSourceVSphereOvfVMTemplateConfigLibraryNoMatch() string {
	return fmt.Sprintf(`
%s

variable "file" {
  type    = string
  default = "%s"
}

data "vsphere_datastore" "ds" {
  datacenter_id = data.vsphere_datacenter.rootdc1.id
  name          = vsphere_nas_datastore.ds1.name
}

resource "vsphere_content_library" "library" {
  name            = "test-content-library-nomatch"
  storage_backing = [data.vsphere_datastore.ds.id]
  description     = "Test library for OVF templates"
}

resource "vsphere_content_library_item" "item" {
  name        = "test-template"
  description = "Test OVF Template"
  library_id  = vsphere_content_library.library.id
  type        = "ovf"
  file_url    = var.file
}

data "vsphere_ovf_vm_template" "template" {
  name_regex       = "^nonexistent-.*"
  library_id       = vsphere_content_library.library.id
  resource_pool_id = data.vsphere_resource_pool.pool1.id
  host_system_id   = data.vsphere_host.roothost1.id
  datastore_id     = data.vsphere_datastore.ds.id
  depends_on       = [vsphere_content_library_item.item]
}
`, testhelper.CombineConfigs(
		testhelper.ConfigDataRootDC1(),
		testhelper.ConfigDataRootHost1(),
		testhelper.ConfigDataRootHost2(),
		testhelper.ConfigResDS1(),
		testhelper.ConfigDataRootComputeCluster1(),
		testhelper.ConfigResResourcePool1(),
	),
		testhelper.ContentLibraryFiles,
	)
}

func testAccDataSourceVSphereOvfVMTemplateConfigLibraryBothName() string {
	return fmt.Sprintf(`
%s

variable "file" {
  type    = string
  default = "%s"
}

data "vsphere_datastore" "ds" {
  datacenter_id = data.vsphere_datacenter.rootdc1.id
  name          = vsphere_nas_datastore.ds1.name
}

resource "vsphere_content_library" "library" {
  name            = "test-content-library-bothname"
  storage_backing = [data.vsphere_datastore.ds.id]
  description     = "Test library for OVF templates"
}

resource "vsphere_content_library_item" "item" {
  name        = "test-template"
  description = "Test OVF Template"
  library_id  = vsphere_content_library.library.id
  type        = "ovf"
  file_url    = var.file
}

data "vsphere_ovf_vm_template" "template" {
  name             = "test-template"
  name_regex       = "^test-.*"
  library_id       = vsphere_content_library.library.id
  resource_pool_id = data.vsphere_resource_pool.pool1.id
  host_system_id   = data.vsphere_host.roothost1.id
  datastore_id     = data.vsphere_datastore.ds.id
  depends_on       = [vsphere_content_library_item.item]
}
`, testhelper.CombineConfigs(
		testhelper.ConfigDataRootDC1(),
		testhelper.ConfigDataRootHost1(),
		testhelper.ConfigDataRootHost2(),
		testhelper.ConfigResDS1(),
		testhelper.ConfigDataRootComputeCluster1(),
		testhelper.ConfigResResourcePool1(),
	),
		testhelper.ContentLibraryFiles,
	)
}
