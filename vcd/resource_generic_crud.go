package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/util"
)

// crudConfig defines a generic approach for managing Terraform resources where the parent entity is
// a standard OpenAPI entity and the outer entity should satisfy 'updateDeleter' type constraint
// (have 'Update' and 'Delete' pointer receiver methods)
type crudConfig[O updateDeleter[O, I], I any] struct {
	// entityLabel to use
	entityLabel string

	// getTypeFunc is responsible for converting schema fields to inner type
	getTypeFunc func(*VCDClient, *schema.ResourceData) (*I, error)
	// stateStoreFunc is responsible for storing state
	stateStoreFunc func(vcdClient *VCDClient, d *schema.ResourceData, outerType O) error

	// createFunc is the function that can create an outer entity based on inner entity config
	// (which is created by 'getTypeFunc')
	createFunc func(config *I) (O, error)

	// resourceReadFunc that will be executed from Create and Update functions
	resourceReadFunc func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics

	// getEntityFunc is a function that retrieves the entity
	// It will use ID for resources and Name for data sources
	getEntityFunc func(idOrName string) (O, error)

	// preCreateHooks will be executed before the entity is created
	preCreateHooks []schemaHook

	// preUpdateHooks will be executed before submitting the data for update
	preUpdateHooks []outerEntityHookInnerEntityType[O, *I]

	// preDeleteHooks will be executed before the entity is deleted
	preDeleteHooks []outerEntityHook[O]

	// readHooks that will be executed after the entity is read, but before it is stored in state
	readHooks []outerEntityHook[O]
}

// updateDeleter is a type constraint to match only entities that have Update and Delete methods
type updateDeleter[O any, I any] interface {
	Update(*I) (O, error)
	Delete() error
}

// outerEntityHook defines a type for hook that can be fed into generic CRUD operations
type outerEntityHook[O any] func(O) error

// schemaHook defines a type for hook that can be fed into generic CRUD operations
type schemaHook func(*VCDClient, *schema.ResourceData) error

// outerEntityHookInnerEntityType defines a type for hook that will provide retrieved outer entity
// with a newly computed inner entity type (useful for modifying update body before submitting it)
type outerEntityHookInnerEntityType[O, I any] func(*schema.ResourceData, O, I) error

func createResource[O updateDeleter[O, I], I any](ctx context.Context, d *schema.ResourceData, meta interface{}, c crudConfig[O, I]) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	t, err := c.getTypeFunc(vcdClient, d)
	if err != nil {
		return diag.Errorf("error getting %s type on create: %s", c.entityLabel, err)
	}

	err = execSchemaHook(vcdClient, d, c.preCreateHooks)
	if err != nil {
		return diag.Errorf("error executing pre-create %s hooks: %s", c.entityLabel, err)
	}

	createdEntity, err := c.createFunc(t)
	if err != nil {
		return diag.Errorf("error creating %s: %s", c.entityLabel, err)
	}

	err = c.stateStoreFunc(vcdClient, d, createdEntity)
	if err != nil {
		return diag.Errorf("error storing %s to state during create: %s", c.entityLabel, err)
	}

	return c.resourceReadFunc(ctx, d, meta)
}

func updateResource[O updateDeleter[O, I], I any](ctx context.Context, d *schema.ResourceData, meta interface{}, c crudConfig[O, I]) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	t, err := c.getTypeFunc(vcdClient, d)
	if err != nil {
		return diag.Errorf("error getting %s type on update: %s", c.entityLabel, err)
	}

	if d.Id() == "" {
		return diag.Errorf("empty id for updating %s", c.entityLabel)
	}

	retrievedEntity, err := c.getEntityFunc(d.Id())
	if err != nil {
		return diag.Errorf("error getting %s for update: %s", c.entityLabel, err)
	}

	err = execUpdateEntityHookWithNewInnerType(d, retrievedEntity, t, c.preUpdateHooks)
	if err != nil {
		return diag.Errorf("error executing pre-update %s hooks: %s", c.entityLabel, err)
	}

	_, err = retrievedEntity.Update(t)
	if err != nil {
		return diag.Errorf("error updating %s with ID: %s", c.entityLabel, err)
	}

	return c.resourceReadFunc(ctx, d, meta)
}

func readResource[O updateDeleter[O, I], I any](_ context.Context, d *schema.ResourceData, meta interface{}, c crudConfig[O, I]) diag.Diagnostics {
	retrievedEntity, err := c.getEntityFunc(d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			util.Logger.Printf("[DEBUG] entity '%s' with ID '%s' not found. Removing from state", c.entityLabel, d.Id())
			d.SetId("")
			return nil
		}
		return diag.Errorf("error getting %s: %s", c.entityLabel, err)
	}

	err = execEntityHook(retrievedEntity, c.readHooks)
	if err != nil {
		return diag.Errorf("error executing read %s hooks: %s", c.entityLabel, err)
	}

	vcdClient := meta.(*VCDClient)
	err = c.stateStoreFunc(vcdClient, d, retrievedEntity)
	if err != nil {
		return diag.Errorf("error storing %s to state during resource read: %s", c.entityLabel, err)
	}

	return nil
}

func deleteResource[O updateDeleter[O, I], I any](_ context.Context, d *schema.ResourceData, _ interface{}, c crudConfig[O, I]) diag.Diagnostics {
	retrievedEntity, err := c.getEntityFunc(d.Id())
	if err != nil {
		return diag.Errorf("error getting %s for delete: %s", c.entityLabel, err)
	}

	err = execEntityHook(retrievedEntity, c.preDeleteHooks)
	if err != nil {
		return diag.Errorf("error executing pre-delete %s hooks: %s", c.entityLabel, err)
	}

	err = retrievedEntity.Delete()
	if err != nil {
		return diag.Errorf("error deleting %s with ID '%s': %s", c.entityLabel, d.Id(), err)
	}

	return nil
}

func execSchemaHook(vcdClient *VCDClient, d *schema.ResourceData, runList []schemaHook) error {
	if len(runList) == 0 {
		util.Logger.Printf("[DEBUG] No hooks to execute")
		return nil
	}

	var err error
	for i := range runList {
		err = runList[i](vcdClient, d)
		if err != nil {
			return fmt.Errorf("error executing hook: %s", err)
		}

	}

	return nil
}

func execEntityHook[O any](outerEntity O, runList []outerEntityHook[O]) error {
	if len(runList) == 0 {
		util.Logger.Printf("[DEBUG] No hooks to execute")
		return nil
	}

	var err error
	for i := range runList {
		err = runList[i](outerEntity)
		if err != nil {
			return fmt.Errorf("error executing hook: %s", err)
		}

	}

	return nil
}

func execUpdateEntityHookWithNewInnerType[O, I any](d *schema.ResourceData, outerEntity O, newInnerEntity I, runList []outerEntityHookInnerEntityType[O, I]) error {
	if len(runList) == 0 {
		util.Logger.Printf("[DEBUG] No hooks to execute")
		return nil
	}

	var err error
	for i := range runList {
		err = runList[i](d, outerEntity, newInnerEntity)
		if err != nil {
			return fmt.Errorf("error executing hook: %s", err)
		}

	}

	return nil
}

// dsReadConfig is a generic type that can be used for data sources. It differs from `crudConfig` in
// the sense that it does not have `updateDeleter` type parameter constraint. This is needed for
// such data sources that have no API to Update and/or Delete an entity, but instead are read-only
// entities.
type dsReadConfig[O any, I any] struct {
	// entityLabel to use
	entityLabel string

	// stateStoreFunc is responsible for storing state
	stateStoreFunc func(vcdClient *VCDClient, d *schema.ResourceData, outerType O) error

	// getEntityFunc is a function that retrieves the entity
	// It will use ID for resources and Name for data sources
	getEntityFunc func(idOrName string) (O, error)

	// preReadHooks will be executed before the entity is created
	preReadHooks []schemaHook

	// overrideDefaultNameField permits to override default field ('name') that passed to
	// getEntityFunc. The field must be a string (schema.TypeString)
	overrideDefaultNameField string
}

// readDatasource will read a data source by a 'name' field in Terraform schema
func readDatasource[O any, I any](_ context.Context, d *schema.ResourceData, meta interface{}, c dsReadConfig[O, I]) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	err := execSchemaHook(vcdClient, d, c.preReadHooks)
	if err != nil {
		return diag.Errorf("error executing pre-read %s hooks: %s", c.entityLabel, err)
	}

	fieldName := "name"
	if c.overrideDefaultNameField != "" {
		fieldName = c.overrideDefaultNameField
		util.Logger.Printf("[DEBUG] Overriding %s field 'name' to '%s' for datasource lookup", c.entityLabel, c.overrideDefaultNameField)
	}
	entityName := d.Get(fieldName).(string)
	retrievedEntity, err := c.getEntityFunc(entityName)
	if err != nil {
		return diag.Errorf("error getting %s by Name '%s': %s", c.entityLabel, entityName, err)
	}

	err = c.stateStoreFunc(vcdClient, d, retrievedEntity)
	if err != nil {
		return diag.Errorf("error storing %s to state during data source read: %s", c.entityLabel, err)
	}

	return nil
}
