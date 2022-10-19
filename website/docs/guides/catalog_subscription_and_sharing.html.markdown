---
layout: "vcd"
page_title: "VMware Cloud Director: catalog-subscription-and-sharing"
sidebar_current: "docs-vcd-guides-catalog-subscription-and-sharing"
description: |-
 Provides guidance to VMware Cloud catalog subscription and sharing
---

# Catalog subscription and sharing

Supported in provider *v3.8+*.

-> In this document, when we mention **tenants**, the term can be substituted with **organizations**.

## Overview

* A [`vcd_catalog`][catalog] is a container of [catalog items][items], which could be either [vApp templates][template] or [media items][media].
* [`vcd_catalog_item`][item]
* [`vcd_catalog_media`][media]
* [`vcd_catalog_vapp_template`][template]
* [`vcd_subscribed_catalog`][subscribed]

## Difference between publishing, sharing, and subscribing


## Publishing operations

## Sharing operations

## Subscribing operations

### Full subscription


### Subscription without automatic download 

### Synchronisation


### The role of catalog items


[catalog]: </providers/vmware/vcd/latest/docs/resources/catalog> (vcd_catalog)
[subscribed]: </providers/vmware/vcd/latest/docs/resources/subscribed_catalog> (vcd_subscribed_catalog)
[item]: </providers/vmware/vcd/latest/docs/resources/catalog_item> (vcd_catalog_item)
[media]: </providers/vmware/vcd/latest/docs/resources/catalog_media> (vcd_catalog_media)
[template]: </providers/vmware/vcd/latest/docs/resources/catalog_vapp_template> (vcd_catalog_vapp_template)

