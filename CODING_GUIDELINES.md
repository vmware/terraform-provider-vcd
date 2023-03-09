# Coding guidelines


## Selecting schema types for fields 

Terraform has a built-in schema types as per [official documentation](https://www.terraform.io/docs/extend/schemas/schema-types.html)
The types are pretty much obvious but sometimes a question may arise whether to use
[TypeSet](https://www.terraform.io/docs/extend/schemas/schema-types.html#typeset) or
[TypeList](https://www.terraform.io/docs/extend/schemas/schema-types.html#typelist)
when an aggregate type is needed.

### When to use [TypeList](https://www.terraform.io/docs/extend/schemas/schema-types.html#typelist)

TypeList with single element block OR where vCD API returns them with a field that allows ordering

*Note*. Always use `TypeList` when one block is needed (limited by `MaxItems = 1`). This will prevent
user and tests from dealing with hashed IDs and is easier to work with in general.

### When to use [TypeSet](https://www.terraform.io/docs/extend/schemas/schema-types.html#typeset)

TypeSet with more than one element blocks AND where vCD API returns them without a field to order on

*Note*. The schema definition may optionally use a `Set` function of type `SchemaSetFunc`. It may be
used when a default hashing function (which calculates hash based on all fields) is not suitable.

## Filtering

Filters provide the ability of retrieving a data source by criteria other than the plain name. They depend on
the query engine in go-vcloud-director. Only types supported by such engine can have a filter.

To add filtering for a new data source:

1. Make sure the underlying entity supports filtering (See "Supporting a new type in the query engine" in 
[go-vcloud-director coding guidelines](https://github.com/vmware/go-vcloud-director/blob/main/CODING_GUIDELINES.md))

2. Add a "filter" field in the data source definition, using the elements listed in `filter.go`

3. Add a function `Get_TYPE_ByFilter` in `filter_get.go`, using the existing ones as model

4. In the data source `Read` function add:

```go
        var instance *govcd._TYPE_
		if !nameOrFilterIsSet(d) {
			return fmt.Errorf(noNameOrFilterError, "vcd_TYPE")
		}
		filter, hasFilter = d.GetOk("filter")
		if hasFilter {
			instance, err = get_TYPE_ByFilter(vdc, filter)
			if err != nil {
				return err
			}
		}
        // use instance
```

5. Extend the test `TestAccSearchEngine` in `datasource_filter_test.go` to include the new type.


## Listings

Every new **resource** needs to have a listing function, to be added in `datasource_vcd_resource_list`, so that we can
return a list of such resource entities.
See `externalNetworkList`, `networkList`, and `lb{ServerPool|ServiceMonitor|VirtualServer}List` for examples.

Once the listing function is ready, we need to add one `case` item to `datasourceVcdResourceListRead` and the name of
the resource in the documentation (`website/docs/d/resource_list.html.markdown`)

## Testing

Every feature in the provider must include testing. See
[TESTING.md](https://github.com/vmware/terraform-provider-vcd/blob/main/TESTING.md)
for more info.

## Handling Terraform Read of disappeared (removed by other means than Terraform) entities

There are some specifics about handling entities in read code (handled in `schema.Resource.Read`)
functions. It also depends on whether it is a resource or a data source.

### Handling Terraform Data Source read

Data source is an entity that is created by other means and only referenced by Terraform to access
some specific data (*ID* or some other value). Whenever a data source is not found - an error must
be returned directly to a data source consumer. Sample snippet of a read function in a data source
code:

```go
...
func datasourceEntityRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
...
  entity, err = vdc.GetEntityByNameOrId(identifier, false)
  if err != nil {
  	return nil, diag.Errorf("entity %s not found: %s", identifier, err)
  }
...
}
```

### Handling Terraform Resource read

Resource is an entity that is created and managed by Terraform. It can happen that a resource created
with Terraform is removed from VCD by other means (UI, CLI, etc.). This would cause an error when a
`refresh` operation (which happens as part of `plan` and `apply` operations) is triggered as the
resource is no longer found. This can *"brick"* workflow as all the user would get on any Terraform
operation, is an error. By convention - if a resource is not found during `refresh` operation, it
must be removed from statefile (using `d.SetId("")`). That way next `refresh` operation removes
those entities from `statefile` and offers to create them again.

**Note**. We want to be sure that we remove the the entity from state **only when an entity is not
found**, but not because of other causes like network error. To solve this, go-vcloud-director SDK
has a special type of error that it returns `govcd.ErrorEntityNotFound`. All functions in the SDK
are expected to return this error when an exact entity is not found. Additionally there are helper
functions `govcd.IsNotFound` and `govcd.ContainsNotFound` which help to ensure that parent entity is
not found because it does not exist anymore.

The expected code behavior for entity lookup in Terraform resource read is:

```go
...
func resourceEntityRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
...
entity, err := vdc.GetEntityById(identifier, false)
if err != nil {
	if govcd.ContainsNotFound(err) {
		d.SetId("")
		return nil
	}
	return diag.Errorf("unable to find entity %s, err: %s", identifier, err)
}
...
```

## Documenting your changes

The documentation of every contribution is as important as the code and related tests. Their combined information
makes it possible for maintainers to understand the changes and expand on them.

The documentation comes in three flavors:

* One [article per resource](#documenting-data-sources-and-resources), located in the directory `./website/docs/r`, and one for the corresponding data source, located 
  in `./website/docs/d`.
* One article about [topics that involve more than one resource](#documenting-broad-topics-guides), and need a wider breadth of documentation to be properly
  explained, located in `./website/guides` (for example: the article **roles management** includes operations with **Roles**,
  **Global Roles**, **Rights**, and **Rights bundles**)
* One or more [**entries for the change log**](#documenting-each-pull-request-for-the-changelog), explaining what each pull request has contributed.

### Documenting data sources and resources

Each entity, be it a resource or a data source, needs to have its documentation page, containing at least the following information:

* A **metadata header**, containing information related to the page location, and a summary of its contents.
* The resource or data source **name**
* An optional indication about when it started to be available (such as `Supported in provider *v3.9+*.`)
* At least one **Usage example**
* The list of **Arguments**, i.e. the properties that the user needs to fill, with the indication of whether they are
  *Required* or *Optional*
* The list of **Attributes**, i.e. the properties that are computed by Terraform and exposed in the resource or data source state
* Resources should also include a section about **importing**, explaining how an existing resource could be imported into
  Terraform state

Each page must then be linked appropriately in the file `./website/docs/vcd.erb`

### Documenting broad topics (guides)

When a topic is too wide to be comprised in the description of one resource or data source, we can make a **Guide**, which
is a free-form article that explains operations including several resources and data sources.
Such articles should include explanations and examples of workflows, relationship between resources, common cases, and
troubleshooting methods, when applicable.

Each guide page must also be linked in `./website/docs/vcd.erb`.

### Documenting each Pull Request for the CHANGELOG

Each pull request (PR) should include some information to be included in the CHANGELOG.
We could modify directly the file `CHANGELOG.md`, but this means that almost every PR will have conflicts, which need
to be resolved manually, with the risk of losing or mangling valuable information. (It has happened)

To avoid conflicts, but still keep recording detailed information about the changes, we adopt a cumulative system.

1. When a cycle starts (immediately after a release), we put a note in `CHANGELOG.md`, informing readers that the changes,
   until the next release, are provisionally stored in the directory `./.changes/v#.#.#` where `#.#.#` is the version
   currently being developed.
2. Each PR will include one or more of the following files in `./.changes/v#.#.#`
   * `###-features.md` containing information about new resources and data sources
   * `###-improvements.md` containing information about enhancements to existing resources and data sources
   * `###-bug-fixes.md` containing reference to one or more bugs being fixed
   * `###-deprecations.md` containing the list of resources, data sources, or properties being deprecated
   * `###-removals.md` containing the list of resources, data sources, or properties that have been removed
   * `###-notes.md`, containing changes that don't fit into the above categories

In the above list, the placeholder `###` stands for the PR number. Each line in each file will have at its end the 
indication of the PR itself, with the format `[GH-###]`.

Before the release, we use two scripts:
* `./scripts/make-changelog.sh`, which collects all the files in `./.changes/v#.#.#` and produces a nicely formatted text
  that we paste into `CHANGELOG.md`
* `./scripts/changelog-links.sh`, which parses `CHANGELOG.md` and produces the URLs that replace every occurrence of `[GH-###]`

As a last operation before the release, we open a PR with the updated `CHANGELOG.md`, giving the team a final chance to review
the text of what will mostly constitute the release notes.

### References:
* https://github.com/vmware/terraform-provider-vcd/pull/800#discussion_r825335060
* https://github.com/vmware/terraform-provider-vcd/issues/855
* https://github.com/vmware/terraform-provider-vcd/pull/925
* https://github.com/vmware/terraform-provider-vcd/issues/611
* https://github.com/vmware/terraform-provider-vcd/pull/783
* https://github.com/vmware/terraform-provider-vcd/pull/451
