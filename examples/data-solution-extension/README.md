# About

This example is a 2-step approach to setup Data Solution Extension (DSE) from scratch and publish it  
to a tenant for consumption. One **must** have Data Solution .iso file to use this example.

After executing both steps, then the tenant should be able to login with newly provisioned user and see
"Mongo DB Community" Data Solution visible (in UI it is at the top bar "More -> Data Solutions ->
Solutions")

**Note**. Technically, these steps could be combined into one, but a new connection is required
after publishing the Data Solution Instance to tenant.

## Step 1

First step configures:
* Solution Landing Zone
* Uploads Data Solution Extension (DSE) Add-On ISO image and configures it in Solution Landing Zone
* Creates a Data Solutions Instance
* Publishes Data Solutions Instance to tenant

## Step 2

Step 2 focuses on Data Solutions Add-On itself:

* Create a global role that has a combination of rights for Container Services Extension (CSE) and
  Data Solutions Extension (DSE)
* Creates a tenant user that has this role and will be able to consume DSE
* Configures Data Solutions 'VCD Data Solutions' and 'MongoDB Community' with default repository
  settings
* Publishes "MongoDB Community" Data Solution to tenant
