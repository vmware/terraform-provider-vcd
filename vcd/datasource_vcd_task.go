package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Note: This data source was created as help to troubleshooting vcd_subscribed_catalogs.
// With fortified code in that resource, however, the need to use the task data source has
// gone away. It is undocumented, for now (entry in vcd.erb commented out), and it will stay so unless we decide to make it public.

func datasourceVcdTask() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdTaskRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the task.",
			},
			"href": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The URI of the task",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the task. May not be unique",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "An optional description of the task",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of the task.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The execution status of the task. One of queued, preRunning, running, success, error, aborted",
			},
			"operation": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "A message describing the operation that is tracked by this task.",
			},
			"operation_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The short name of the operation that is tracked by this task.",
			},
			"start_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time the system started executing the task. May not be present if the task has not been executed yet.",
			},
			"end_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time that processing of the task was completed. May not be present if the task is still being executed.",
			},
			"expiry_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time at which the task resource will be destroyed and no longer available for retrieval. May not be present if the task has not been executed or is still being executed.",
			},
			"cancel_requested": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether user has requested this processing to be canceled.",
			},
			"owner_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the task owner. This is typically the object that the task is creating or updating.",
			},
			"owner_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The type of the task owner",
			},
			"owner_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique identifier of the task owner",
			},
			"error": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "error information from a failed task",
			},
			"user_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the user who started the task",
			},
			"user_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique identifier of the task user",
			},
			"org_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the org to which the user belongs",
			},
			"org_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique identifier of the user org",
			},
			"progress": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Indicator of task progress as an approximate percentage between 0 and 100. Not available for all tasks.",
			},
		},
	}
}

func datasourceVcdTaskRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	taskId := d.Get("id").(string)
	task, err := vcdClient.Client.GetTaskById(taskId)
	if err != nil {
		return diag.FromErr(err)
	}
	dSet(d, "href", task.Task.HREF)
	dSet(d, "type", task.Task.Type)
	dSet(d, "status", task.Task.Status)
	dSet(d, "operation", task.Task.Operation)
	dSet(d, "operation_name", task.Task.OperationName)
	dSet(d, "start_time", task.Task.StartTime)
	dSet(d, "end_time", task.Task.EndTime)
	dSet(d, "expiry_time", task.Task.ExpiryTime)
	dSet(d, "cancel_requested", task.Task.CancelRequested)
	dSet(d, "description", task.Task.Description)
	dSet(d, "progress", task.Task.Progress)
	if task.Task.Owner != nil {
		dSet(d, "owner_name", task.Task.Owner.Name)
		dSet(d, "owner_id", extractUuid(task.Task.Owner.HREF))
		dSet(d, "owner_type", task.Task.Owner.Type)
	}
	if task.Task.User != nil {
		dSet(d, "user_name", task.Task.User.Name)
		dSet(d, "user_id", extractUuid(task.Task.User.HREF))
	}
	if task.Task.Organization != nil {
		dSet(d, "org_name", task.Task.Organization.Name)
		dSet(d, "org_id", extractUuid(task.Task.Organization.HREF))
	}
	if task.Task.Error != nil {
		dSet(d, "error", task.Task.Error.Error())
	}
	d.SetId(extractUuid(task.Task.ID))
	return nil
}
