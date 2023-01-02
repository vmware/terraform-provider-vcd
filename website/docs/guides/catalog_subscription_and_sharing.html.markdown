---
layout: "vcd"
page_title: "VMware Cloud Director: Catalog Subscription and Sharing"
sidebar_current: "docs-vcd-guides-catalog-subscription-and-sharing"
description: |-
 Provides guidance to VMware Cloud catalog publishing, subscribing, and sharing
---

# Catalog subscription and sharing

Supported in provider *v3.8+*.

-> In this document, when we mention **tenants**, the term can be substituted with **organizations**.

## Overview

This document explains some common scenarios in VMware Cloud Director catalog usage, with special emphasis on catalog sharing and subscribing.

## Glossary

* A [catalog][catalog] is a container of [catalog items][item], which in turn could be either [vApp templates][template] or [media items][media]. It could be published, subscribed, or shared.
* A [catalog item][item] is a generic object that represents a more specific entity: a [vApp template][template] or a [media item][media]
* A [media item][media] is a catalog element containing user-defined data. Usually, they are .ISO files, but users can put basically any file. It is located inside a [catalog item][item].
* A [vApp template][template] is a special file that either has been exported from a vApp or has been assembled to become one. It must have a specific format (usually a .OVA). It is located inside a [catalog item][item].
* A [subscribed_catalog][subscribed] is a catalog that was created by subscribing to another catalog. It is not modifiable and not publishable.
* A [published catalog][catalog] is a catalog that has been made available for subscription. It cannot be a subscribed catalog.
* A [shared catalog][shared] is a catalog that can be used by users other than their owner, who decides which type of access each one is granted.

## Differences between catalog types

As we can see from the glossary, we have 4 types of catalog usage:

### Regular catalog
A regular [`vcd_catalog`][catalog], i.e. a catalog that is not being published, shared, or subscribed, belongs to its
owners, who can do with it as they please. The owners could be the creators of the catalog, but can also be tenants, to
whom the catalog was assigned by a provider.

Bottom line: _It's my catalog, and only I (or my organization) can use it_.

### Shared catalog
A `vcd_catalog` can also be [shared][shared]. Here, the owners keep control on the catalog, but they can grant broad permissions
(including full control) to specific users, or read-only access to whole organizations.
In this type of catalog, the data stays in the original catalog, but authorised users and organizations can read and write
its contents, depending on the type of access they were granted.
A shared catalog is accessible by its authorised users until the access is revoked.

Bottom line: _It's still my catalog, but I allow others to use it_.

### Published catalog
A `vcd_catalog` is [published][catalog] when its owners open it for external access. Unlike a shared catalog, external users can't use
the catalog contents directly, but obtain a subscription to it. There are no degrees of publishing like in sharing: it's
either published or not published. When it is published, anyone who knows the subscription URL and the optional password
can access it. The publishing operation is blind, i.e. the owners may not know who will access the catalog, or from where.
Another important difference between publishing and sharing is that publishing can be practiced between different VCDs, while sharing
must be done within the same VCD.

Bottom line: _It's my catalog, and I'll let you copy it if you can see it_.

### Subscribed catalog
A [`vcd_subscribed_catalog`][subscribed] is created with a subscription to a published catalog. It becomes a new catalog, with
local contents that are copied over from the published catalog, and the owner can decide whether to copy either all-at-once 
or one item at the time, when needed.
Once the contents are copied (they are said to be _synchronised_) they stay available in the local storage even if the
subscription is rescinded or otherwise interrupted.
The subscribed catalog can synchronise its contents with the publisher, thus getting newly published items, as long as the
published status of the original catalog is not removed.

Bottom line: _It's a borrowed catalog that I can use if I have copied all the contents that I need_.

## Sharing operations

A [shared `vcd_catalog`][shared] is a regular catalog (it could also be published) whose access has been altered by its owners.
The normal access of a catalog is by users of the same organization with enough privileges to see and modify its items.
With access control, the owners can grant specific access (read-only, read/write, full access) to users, or read-only access
to organizations.

```hcl
resource "vcd_catalog_access_control" "AC-users-and-orgs" {
  catalog_id = data.vcd_catalog.Catalog-AC-2.id

  shared_with_everyone = false

  shared_with {
    user_id      = data.vcd_org_user.ac-admin1.id
    access_level = "FullControl"
  }
  shared_with {
    user_id      = data.vcd_org_user.ac-vapp-creator2.id
    access_level = "Change"
  }
  shared_with {
    org_id       = data.vcd_org.another-org.id
    access_level = "ReadOnly"
  }
}
```

The users of a shared catalog will be able to use the catalog immediately, once granted permission. There is no copy
or synchronisation involved.

There are a few restrictions: 

* a catalog can only be shared within the same VCD.
* a shared catalog cannot be edited by users of the granted organization. It may happen, then, that more than one catalog
  with the same name appear within one organization.

Shared grants can be revoked. The organization or user whose permission was revoked will cease seeing the catalog and its items.
Virtual applications and VMs that have been created using shared items that are no longer available will **not** be
affected, i.e. they will continue existing.

## Publishing operations

Publishing a catalog is an operation that only involves the owner of the catalog. The operation can happen during the
catalog creation or as an update.

```hcl
resource "vcd_catalog" "publisher" {
  org                = "my-org"
  name               = "publisher"
  description        = "publisher catalog"
  storage_profile_id = data.vcd_storage_profile.storage_profile.id
  delete_force       = true
  delete_recursive   = true

  # publishing parameters
  publish_enabled               = true
  cache_enabled                 = true
  preserve_identity_information = false
  password                      = var.password
}
```
Once a catalog has been published, there will be some computed properties that will be useful, especially when exposed
through a data source:

* `publish_subscription_url` is the complete URL that users need to use to create a [subscribed catalog][subscribed].
* `number_of_vapp_templates` and `number_of_media` are useful to know how many items to expect.
* `vapp_template_list` and `media_item_list` show the names of all items contained in the catalog. They could be useful 
  to help subscribers see the difference between available items and already synchronised ones.   

The moment when the catalog is published is important for subscribers, although they will have no say in that. It's important,
however, that the publishing entity –either a provider or a designated tenant– understands these points:

* publishing a catalog after it is filled with all its items will give subscribers immediate access to the published resources;
* publishing a catalog at creation may force subscribers to run synchronisation operations more often.

Owners of a published catalog have no direct visibility over the subscribers. They only advertise the subscription URL
and pass along the optional password. The recipients can be in the same VCD or in a different one.

## Subscribing operations

A [subscribed catalog][subscribed] is created by subscribing to a published URL from a remote catalog. Users of this catalog
can set the catalog name, and decide how they want to manage the flux of data.

This is the simplest subscription operation:

```hcl
resource "vcd_subscribed_catalog" "test-subscriber" {
  org  = "another-org"
  name = "subscriber"

  delete_force     = "true"
  delete_recursive = "true"

  subscription_url      = var.subscription_url
  subscription_password = var.password
}
```

To define how the subscription happens, there are a few properties that can be used:

* During creation:
  * `make_local_copy` enables the automatic download of all data on subscription or synchronisation (update)
* During update:
  * `sync_on_refresh=true` (recommended) enables the synchronisation every time the resource is read. 
  * `sync_catalog=true` will synchronise the whole catalog. When `make_local_copy` was used, it also synchronises every catalog item.
  * `sync_all_vapp_templates=true` synchronises all vApp templates. 
  * `sync_all_media_items=true` synchronises all media items.
  * `sync_vapp_templates=[name1, name2]` will synchronise only the vApp templates in the list.
  * `sync_media_items=[name1,name2` will synchronise only the media items in the list.
  * `sync_all` corresponds to `sync_catalog` + `sync_all_vapp_templates` + `sync_all_media_items`

### Full subscription

A full subscription occurs when `make_local_copy` is set to true.
Its effects depend on when the subscription occurs.

* Scenario 1. The subscription happens when the publishing catalog has already created all its items.
* Scenario 2. The subscription happens when the publishing catalog is still empty.
* Scenario 3. The subscription happens when the publishing catalog is full, but more items may be added later on.

In **scenario 1**, the subscribing catalog will access immediately all items, and start synchronising them in background.
In **scenario 2**, the subscribing catalog will get nothing, and will need to synchronise later on.
In **scenario 3**, the subscribing catalog will access all items immediately, although it will have to be synchronised later on.

### Subscription without automatic download 

~> This option is only recommended to System administrators.

If we subscribe the catalog without automatic downloads (`make_local_copy = false`) there will be no immediate access to
the catalog resources. The catalog and its items will need to be synchronised explicitly.
When we are in this situation, adding `sync_catalog=true` to the creation script will only get general information about the
catalog items, but not their data. It's like in the web interface: when we run "catalog sync", we see the list of vApp templates
and media items, but the items are not available for usage.
In the web interface, we would need to synchronise item by item. In `vcd_subscribed_catalog`, we can run several operations at once,
i, e. we can choose among the following:

1. `sync_catalog` will synchronise only the catalog, meaning that it will only download the **list of items**.
2. `sync_all_media_items` and `sync_all_vapp_templates` will start the synchronisation of all the items that are visible
    at that moment, i.e. the ones that were fetched with `sync_catalog`.
3. `sync_vapp_templates` and `sync_media_items` will fetch only the items that we indicate.
4. `sync_all` corresponds to `sync_catalogz + `sync_all_vapp_templates` + `sync_all_media_items`
5. `sync_on_refresh` enables synchronisation every time that the catalog gets refreshed.

Running a subscribed catalog without automatic downloads (`make_local_copy=false`) could be convenient when we know that
the publishing catalog has a large number of items, and we don't need all of them, or we want to schedule the
synchronisation during hours of light network traffic.

The important thing to remember is that, when `synch_on_refresh` is not set, the synchronisation properties are only
used during update, and due to the way Terraform works, they will be ignored if they were set also during creation.

Thus, we have two ways of operating:

* **in stages**, i.e. using none of the `sync_*` properties when creating the catalog, and then adding which ones we want
  before running `terraform apply` a second time, but without setting `sync_on_refresh`. The synchronisation will only
  happen from the second `apply`.
* **all at once**, i.e. setting `sync_on_refresh` together with any of the options from 1. to 4. above. The synchronisation
  will happen every time the resource is read, which may introduce some delays in the operations if you were expecting
  `terraform refresh` to be quick. Despite the possible side effects, this is the recommended way.

In the [examples directory](https://github.com/vmware/terraform-provider-vcd/examples/subscribed_catalog) we can see a full
example of a published and a subscribed catalog.

### Synchronisation fine tuning

If we want to have fine control over the synchronisation, we can use the list of tasks being generated during the [subscribed
catalog][subscribed] operations. Such list is produced and updated for every update operation (and even during refresh
if `sync_on_refresh` is set). The list is optionally saved on file, when the property `store_tasks` is set. This allows
us to see which tasks are running, and even show the details of the tasks using a data source.

### Subscribed resources monitoring

Compared to other resources, `vcd_subscribed_catalog` is peculiar for several reasons:

* While we subscribe to one resource (catalog), we are getting several more (vApp templates and media items) that arrive
  in background.
* The additional resources are not visible to Terraform, and there is currently (as of *v3.8*) no easy way to import then
  into Terraform state.
* Each synchronisation cycle (which may happen stealthily if `sync_on_refresh` is set) may bring new resources under your
  control, and the only notice you have of that is an increase in the list of vApp templates or media items in your
  subscribed catalog state.

One way of achieving some monitoring of a synchronising catalog is by way of using the number of items as trigger to
establishing data source for the subscribed items, like in the example below:

```hcl
data "vcd_catalog_vapp_template" "subscribed_templates" {
  count      = vcd_subscribed_catalog.subscriber.number_of_vapp_templates
  org        = "other-org"
  catalog_id = vcd_subscribed_catalog.subscriber.id
  name       = vcd_subscribed_catalog.subscriber.vapp_template_list[count.index]
}

data "vcd_catalog_media" "subscribed_media" {
  count   = vcd_subscribed_catalog.subscriber.number_of_media
  org     = "other-org"
  catalog = vcd_subscribed_catalog.subscriber.name
  name    = vcd_subscribed_catalog.subscriber.media_item_list[count.index]
}

output "templates" {
  value = data.vcd_catalog_vapp_template.subscribed_templates
}

output "media" {
  value = data.vcd_catalog_media.subscribed_media
}
```

Note: If the synchronisation has not happened yet, the data source definition may fail, as the subscribed catalog has not
received the items. If that happens, a refresh of the resource will solve the issue.

### The role of catalog items

While we work directly with vApp templates and media items, behind the scenes both these entities are catalog items.
Most of the time, this distinction is unimportant. However, it's worth noting a few concepts:

* The synchronisation happens with catalog items, which in turn contain either a vApp template or a media item;
* When a catalog has been synchronised, but not the individual items, the vApp templates and media items contained
  in the catalog items are **temporary**. As soon as each catalog item gets synchronised, the underlying entity
  (vApp template or media) is deleted and the real one is created through the subscription. You can observe this
  behavior in the web interface. As soon as you ask to synchronise one item, you will see a task for its deletion,
  immediately followed by a creation/synchronisation task.

The above points are important in one particular case: if you synchronise the catalog (`sync_catalog` property) but not
the vApp templates, and try creating a data source using the ID of a temporary entity, you will get an error.


## Troubleshooting

### Stuck on deletion: side effects of "sync\_on\_refresh"

We said before that the most efficient way of keeping the subscribed catalog synchronised is to use `sync_on_refresh`,
allowing a synchronisation every time the catalog is read, which means `terraform plan`, `terraform refresh`, and
`terraform apply`, even if no properties were altered since the last apply.
This property is a workaround to a well known Terraform limitation, which does not allow to update a resource unless
its schema has been changed.
The workaround has the side effect that a new synchronisation will start on almost every terraform command, including,
oddly enough, `terraform destroy`. The first operation before apply and destroy is a refresh, whether we ask for it or
not. In our case, this situation will trigger a synchronisation right when we don't need it, since we have decided
to remove the subscribed catalog.

To avoid this side effect, we should call `terraform destroy` with the option `-refresh=false`.

### Mixing "sync\_on\_refresh" and update

If we use `sync_on_refresh`, we should take care not to make changes to the configuration that would trigger an update.
The reason is that, when we run `terraform apply` for an update, the first operation is a refresh, which starts a synchronisation
because of `sync_on_refresh`. Then the update will kick in, invoking another synchronisation. Depending on the number and
size of the catalog items, this double request may cause an error.

Usually, repeating a simple refresh should fix the issue. 

### Subscribed vApp templates data sources 

There is one scenario where data sources of catalog entities (vApp templates and media items) may not be usable immediately.
This happens in one of the following cases:

1. The subscribed catalog has `make_local_copy=false`, and has only synchronised the catalog (`sync_catalog`) but not the dependant items.
2. The subscribed catalog has `make_local_copy=true`, but the full synchronisation is still happening.

If you try using a vApp template to create a VM, and get an error about the item being unavailable, the reason could be
an incomplete synchronisation.

The remedy, in case #1, is to add one of `sync_all_vapp_templates` or `sync_vapp_templates` to the configuration file, and
run a refresh (if `synch_on_refresh` was set) or an update.

In case #2 –which may also occur for #1 after the remedy was applied– the only remedy is **waiting** until the synchronisation is done.


[catalog]: </providers/vmware/vcd/latest/docs/resources/catalog> (vcd_catalog)
[shared]: </providers/vmware/vcd/latest/docs/resources/catalog_access_control> (vcd_catalog)
[subscribed]: </providers/vmware/vcd/latest/docs/resources/subscribed_catalog> (vcd_subscribed_catalog)
[item]: </providers/vmware/vcd/latest/docs/resources/catalog_item> (vcd_catalog_item)
[media]: </providers/vmware/vcd/latest/docs/resources/catalog_media> (vcd_catalog_media)
[template]: </providers/vmware/vcd/latest/docs/resources/catalog_vapp_template> (vcd_catalog_vapp_template)

