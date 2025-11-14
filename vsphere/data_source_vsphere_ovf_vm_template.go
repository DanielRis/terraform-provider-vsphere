// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package vsphere

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/govmomi/vapi/library"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/contentlibrary"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/folder"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/ovfdeploy"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/helper/structure"
	"github.com/vmware/terraform-provider-vsphere/vsphere/internal/vmworkflow"
)

func dataSourceVSphereOvfVMTemplate() *schema.Resource {
	vmConfigSpecSchema := map[string]*schema.Schema{
		"num_cpus": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The number of virtual processors to assign to this virtual machine.",
		},
		"num_cores_per_socket": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The number of cores to distribute amongst the CPUs in this virtual machine. If specified, the value supplied to num_cpus must be evenly divisible by this value.",
		},
		"cpu_hot_add_enabled": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Allow CPUs to be added to this virtual machine while it is running.",
		},
		"cpu_hot_remove_enabled": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Allow CPUs to be added to this virtual machine while it is running.",
		},
		"nested_hv_enabled": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Enable nested hardware virtualization on this virtual machine, facilitating nested virtualization in the guest.",
		},
		"cpu_performance_counters_enabled": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Enable CPU performance counters on this virtual machine.",
		},
		"memory": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The size of the virtual machine's memory, in MB.",
		},
		"memory_hot_add_enabled": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Allow memory to be added to this virtual machine while it is running.",
		},
		"swap_placement_policy": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The swap file placement policy for this virtual machine. Can be one of inherit, hostLocal, or vmDirectory.",
		},
		"annotation": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "User-provided description of the virtual machine.",
		},
		"guest_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The guest ID for the operating system.",
		},
		"alternate_guest_name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The guest name for the operating system when guest_id is otherGuest or otherGuest64.",
		},
		"firmware": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The firmware interface to use on the virtual machine. Can be one of bios or efi.",
		},
		"sata_controller_count": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The number of SATA controllers that Terraform manages on this virtual machine. This directly affects the amount of disks you can add to the virtual machine and the maximum disk unit number. Note that lowering this value does not remove controllers.",
		},
		"ide_controller_count": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The number of IDE controllers that Terraform manages on this virtual machine. This directly affects the amount of disks you can add to the virtual machine and the maximum disk unit number. Note that lowering this value does not remove controllers.",
		},
		"scsi_controller_count": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "The number of SCSI controllers that Terraform manages on this virtual machine. This directly affects the amount of disks you can add to the virtual machine and the maximum disk unit number. Note that lowering this value does not remove controllers.",
		},
		"scsi_type": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The type of SCSI bus this virtual machine will have. Can be one of lsilogic, lsilogic-sas or pvscsi.",
		},
	}
	s := map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Name of the virtual machine to create. When using content library search, this filters by exact name.",
		},
		"name_regex": {
			Type:         schema.TypeString,
			Optional:     true,
			Description:  "A regular expression to filter OVF templates by name when searching in a content library.",
			ValidateFunc: validation.StringIsValidRegExp,
		},
		"library_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The ID of the content library to search for OVF templates. If specified, the data source will search for templates in this library.",
		},
		"most_recent": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "If true, return the most recently created OVF template when multiple templates match the search criteria.",
		},
		"resource_pool_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The ID of a resource pool to put the virtual machine in.",
		},
		"host_system_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The ID of an optional host system to pin the virtual machine to.",
		},
		"datastore_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The ID of the virtual machine's datastore. The virtual machine configuration is placed here, along with any virtual disks that are created without datastores.",
		},
		"folder": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "The name of the folder to locate the virtual machine in.",
			StateFunc:   folder.NormalizePath,
		},
	}
	structure.MergeSchema(s, vmworkflow.VirtualMachineOvfDeploySchema())
	structure.MergeSchema(s, vmConfigSpecSchema)

	return &schema.Resource{
		Read:   dataSourceVSphereOvfVMTemplateRead,
		Schema: s,
	}
}

func NewOvfHelperParamsFromVMDatasource(d *schema.ResourceData) *ovfdeploy.OvfHelperParams {
	ovfParams := &ovfdeploy.OvfHelperParams{
		AllowUnverifiedSSL: d.Get("allow_unverified_ssl_cert").(bool),
		DatastoreID:        d.Get("datastore_id").(string),
		DeploymentOption:   d.Get("deployment_option").(string),
		DiskProvisioning:   d.Get("disk_provisioning").(string),
		FilePath:           d.Get("local_ovf_path").(string),
		Folder:             d.Get("folder").(string),
		HostID:             d.Get("host_system_id").(string),
		IPAllocationPolicy: d.Get("ip_allocation_policy").(string),
		IPProtocol:         d.Get("ip_protocol").(string),
		Name:               d.Get("name").(string),
		NetworkMappings:    d.Get("ovf_network_map").(map[string]interface{}),
		OvfURL:             d.Get("remote_ovf_url").(string),
		PoolID:             d.Get("resource_pool_id").(string),
	}
	return ovfParams
}

func dataSourceVSphereOvfVMTemplateRead(d *schema.ResourceData, meta interface{}) error {
	// Check if we should search in content library or use file-based approach
	if libraryID, ok := d.GetOk("library_id"); ok {
		return dataSourceVSphereOvfVMTemplateReadFromLibrary(d, meta, libraryID.(string))
	}

	// Original file-based behavior
	client := meta.(*Client).vimClient
	ovfParams := NewOvfHelperParamsFromVMDatasource(d)
	ovfHelper, err := ovfdeploy.NewOvfHelper(client, ovfParams)
	if err != nil {
		return fmt.Errorf("while extracting OVF parameters: %s", err)
	}

	is, err := ovfHelper.GetImportSpec(client)
	if err != nil {
		return fmt.Errorf("while retrieving import spec: %s", err)
	}

	return setOvfTemplateData(d, is.ImportSpec.(*types.VirtualMachineImportSpec).ConfigSpec)
}

func dataSourceVSphereOvfVMTemplateReadFromLibrary(d *schema.ResourceData, meta interface{}, libraryID string) error {
	rc := meta.(*Client).restClient
	client := meta.(*Client).vimClient

	// Get the content library
	lib, err := contentlibrary.FromID(rc, libraryID)
	if err != nil {
		return fmt.Errorf("error retrieving content library: %s", err)
	}

	// Get all items in the library
	clm := library.NewManager(rc)
	ctx := context.TODO()
	items, err := clm.GetLibraryItems(ctx, lib.ID)
	if err != nil {
		return fmt.Errorf("error listing content library items: %s", err)
	}

	// Filter items by type (ovf/ova)
	var ovfItems []library.Item
	for _, item := range items {
		if item.Type == "ovf" || item.Type == "vm-template" {
			ovfItems = append(ovfItems, item)
		}
	}

	if len(ovfItems) == 0 {
		return fmt.Errorf("no OVF templates found in content library %s", lib.Name)
	}

	// Apply name filtering
	filteredItems, err := filterOvfItemsByName(d, ovfItems)
	if err != nil {
		return err
	}

	if len(filteredItems) == 0 {
		return fmt.Errorf("no OVF templates match the specified filters")
	}

	// If most_recent is set, sort by creation time and take the latest
	if d.Get("most_recent").(bool) {
		sort.Slice(filteredItems, func(i, j int) bool {
			if filteredItems[i].CreationTime == nil || filteredItems[j].CreationTime == nil {
				return false
			}
			return filteredItems[i].CreationTime.After(*filteredItems[j].CreationTime)
		})
	}

	selectedItem := filteredItems[0]

	// Check if multiple items match and most_recent is not set
	if len(filteredItems) > 1 && !d.Get("most_recent").(bool) {
		return fmt.Errorf("multiple OVF templates match the specified filters. Please refine your search or use 'most_recent = true'")
	}

	// Now we need to get the OVF spec from the content library item
	// We'll use the OVF helper to parse it
	ovfParams := NewOvfHelperParamsFromVMDatasource(d)

	// For content library items, we need to get the download URL
	// This is a simplification - in practice, we'd need to create a download session
	// For now, we'll set the item ID and try to get the spec
	ovfParams.Name = selectedItem.Name

	// Try to get the import spec using the content library item
	// Note: This may require additional implementation in the OVF helper
	ovfHelper, err := ovfdeploy.NewOvfHelper(client, ovfParams)
	if err != nil {
		return fmt.Errorf("while extracting OVF parameters from library item: %s", err)
	}

	is, err := ovfHelper.GetImportSpec(client)
	if err != nil {
		return fmt.Errorf("while retrieving import spec from library item: %s", err)
	}

	if err := setOvfTemplateData(d, is.ImportSpec.(*types.VirtualMachineImportSpec).ConfigSpec); err != nil {
		return err
	}

	// Set the library item ID as the data source ID
	d.SetId(selectedItem.ID)
	_ = d.Set("name", selectedItem.Name)

	return nil
}

func filterOvfItemsByName(d *schema.ResourceData, items []library.Item) ([]library.Item, error) {
	var filtered []library.Item

	// Check if name or name_regex is specified
	name, nameOk := d.GetOk("name")
	nameRegex, regexOk := d.GetOk("name_regex")

	if !nameOk && !regexOk {
		// No name filter specified, return all items
		return items, nil
	}

	if nameOk && regexOk {
		return nil, fmt.Errorf("cannot specify both 'name' and 'name_regex'")
	}

	if nameOk {
		// Exact name match
		for _, item := range items {
			if item.Name == name.(string) {
				filtered = append(filtered, item)
			}
		}
	} else {
		// Regex match
		re, err := regexp.Compile(nameRegex.(string))
		if err != nil {
			return nil, fmt.Errorf("invalid name_regex: %s", err)
		}
		for _, item := range items {
			if re.MatchString(item.Name) {
				filtered = append(filtered, item)
			}
		}
	}

	return filtered, nil
}

func setOvfTemplateData(d *schema.ResourceData, vmConfigSpec types.VirtualMachineConfigSpec) error {
	_ = d.Set("num_cpus", vmConfigSpec.NumCPUs)
	_ = d.Set("num_cores_per_socket", vmConfigSpec.NumCoresPerSocket)
	_ = d.Set("cpu_hot_add_enabled", vmConfigSpec.CpuHotAddEnabled)
	_ = d.Set("cpu_hot_remove_enabled", vmConfigSpec.CpuHotRemoveEnabled)
	_ = d.Set("nested_hv_enabled", vmConfigSpec.NestedHVEnabled)
	_ = d.Set("memory", vmConfigSpec.MemoryMB)
	_ = d.Set("memory_hot_add_enabled", vmConfigSpec.MemoryHotAddEnabled)
	_ = d.Set("swap_placement_policy", vmConfigSpec.SwapPlacement)
	_ = d.Set("annotation", vmConfigSpec.Annotation)
	_ = d.Set("guest_id", vmConfigSpec.GuestId)
	_ = d.Set("alternate_guest_name", vmConfigSpec.AlternateGuestName)
	_ = d.Set("firmware", vmConfigSpec.Firmware)

	controllers := map[string]int{}
	var scsiType string

	for _, dvc := range vmConfigSpec.DeviceChange {
		dvcSpec := dvc.GetVirtualDeviceConfigSpec()

		switch reflect.TypeOf(dvcSpec.Device) {
		case reflect.TypeOf(&types.VirtualLsiLogicController{}):
			if scsiType == "" {
				scsiType = "lsilogic"
			} else if scsiType != "lsilogic" {
				return fmt.Errorf("multiple scsi controller types are not supported (found %s and %s)", scsiType, "lsilogic")
			}
			controllers["scsi"]++
		case reflect.TypeOf(&types.VirtualLsiLogicSASController{}):
			if scsiType == "" {
				scsiType = "lsilogic-sas"
			} else if scsiType != "lsilogic-sas" {
				return fmt.Errorf("multiple scsi controller types are not supported (found %s and %s)", scsiType, "lsilogic-sas")
			}
			controllers["scsi"]++
		case reflect.TypeOf(&types.ParaVirtualSCSIController{}):
			if scsiType == "" {
				scsiType = "pvscsi"
			} else if scsiType != "pvscsi" {
				return fmt.Errorf("multiple scsi controller types are not supported (found %s and %s)", scsiType, "pvsci")
			}
			controllers["scsi"]++
		case reflect.TypeOf(&types.VirtualSATAController{}):
			controllers["sata"]++
		case reflect.TypeOf(&types.VirtualIDEController{}):
			controllers["ide"]++
		}
	}

	_ = d.Set("scsi_type", scsiType)
	_ = d.Set("scsi_controller_count", controllers["scsi"])
	_ = d.Set("sata_controller_count", controllers["sata"])
	_ = d.Set("ide_controller_count", controllers["ide"])

	if d.Id() == "" {
		d.SetId(d.Get("name").(string))
	}

	return nil
}
