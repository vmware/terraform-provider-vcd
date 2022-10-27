---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_task"
sidebar_current: "docs-vcd-data-source-task"
description: |-
 Provides a VMware Cloud Director Organization Task data source. This can be used to read existing tasks.
---

# vcd\_task

Provides a data source for available tasks.

Supported in provider *v3.8+*

## Example usage

```hcl
data "vcd_task" "some-task" {
  id = "d4fdcaa9-8db4-45a9-80b8-69de49901bc7"
}

output "some-task" {
  value = data.vcd_task.some-task
}
```

```
Sample output:

some_task = {
  "cancel_requested" = false
  "description" = ""
  "end_time" = "2022-10-27T06:17:27.326Z"
  "error" = tostring(null)
  "expiry_time" = "2023-01-25T06:17:26.122Z"
  "href" = "https://example.com/api/task/d4fdcaa9-8db4-45a9-80b8-69de49901bc7"
  "id" = "d4fdcaa9-8db4-45a9-80b8-69de49901bc7"
  "name" = tostring(null)
  "operation" = "Created Catalog remote-subscriber(919e800b-088b-40ea-813c-5458b55829e7)"
  "operation_name" = "catalogCreateCatalog"
  "org_id" = "79b61f14-34f2-4b65-98cb-b5427ce57d67"
  "org_name" = "datacloud"
  "owner_id" = "919e800b-088b-40ea-813c-5458b55829e7"
  "owner_name" = "remote-subscriber"
  "owner_type" = "application/vnd.vmware.admin.catalog+xml"
  "progress" = 0
  "start_time" = "2022-10-27T06:17:26.122Z"
  "status" = "success"
  "type" = "application/vnd.vmware.vcloud.task+xml"
  "user_id" = "cb5df7fb-34c4-4ce1-99e2-f0094458c486"
  "user_name" = "administrator"
}
```

## Argument Reference

The following arguments are supported:

* `id` - (Required) The ID of the task

## Attribute reference

* `href` - The URI of the task.
* `name` - Name of the task. May not be unique. Defines the general operation being performed.
* `description` - An optional description of the task.
* `type` - Type of the task.
* `status` - The execution status of the task. One of queued, preRunning, running, success, error, aborted.
* `operation` - A message describing the operation that is tracked by this task.
* `operation_name` - The short name of the operation that is tracked by this task.
* `start_time` - The date and time the system started executing the task. May not be present if the task has not been executed yet.
* `end_time` - The date and time that processing of the task was completed. May not be present if the task is still being executed.
* `expiry_time` - The date and time at which the task resource will be destroyed and no longer available for retrieval. May not be present if the task has not been executed or is still being executed.
* `cancel_requested` - Whether user has requested this processing to be canceled.
* `owner_name` - The name of the task owner. This is typically the object that the task is creating or updating.
* `owner_type` - The type of the task owner.
* `owner_id` - The unique identifier of the task owner.
* `error` - error information from a failed task.
* `user_name` - The name of the user who started the task.
* `user_id` - The unique identifier of the task user.
* `org_name` - The name of the org to which the user belongs.
* `org_id` - The unique identifier of the user org.
* `progress` - Indicator of task progress as an approximate percentage between 0 and 100. Not available for all tasks.
