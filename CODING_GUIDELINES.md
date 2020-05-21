# Coding guidelines


( **Work in progress** )

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
used when a default hashing function (which calcluates hash based on all fields) is not suitable.

## Filtering

Filters provide the ability of retrieving a data source by criteria other than the plain name. They depend on
the query engine in go-vcloud-director. Only types supported by such engine can have a filter.

To add filtering for a new data source:

1. Make sure the underlying entity supports filtering (See "Supporting a new type in the query engine" in 
[go-vcloud-director coding guidelines](https://github.com/vmware/go-vcloud-director/blob/master/CODING_GUIDELINES.md))

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

## Testing

Every feature in the provider must include testing. See
[TESTING.md](https://github.com/terraform-providers/terraform-provider-vcd/blob/master/TESTING.md)
for more info.
