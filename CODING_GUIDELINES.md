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

## Testing

Every feature in the provider must include testing. See
[TESTING.md](https://github.com/terraform-providers/terraform-provider-vcd/blob/master/TESTING.md)
for more info.
