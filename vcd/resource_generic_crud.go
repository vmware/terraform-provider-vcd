package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// updateDeleter is a type constraint to match only entities that have Update and Delete methods
type updateDeleter[O any, I any] interface {
	Update(*I) (O, error)
	Delete() error
}

type resourceHook[O any] func(O) error
type beforeCreateHook func(*VCDClient, *schema.ResourceData) error

type crudConfig[O updateDeleter[O, I], I any] struct {
	// entityLabel to use in logs and
	entityLabel string

	// getTypeFunc is responsible for converting schema fields to inner type
	getTypeFunc func(d *schema.ResourceData) (*I, error)
	// stateStoreFunc is responsible for storing state
	stateStoreFunc func(d *schema.ResourceData, outerType O) error

	// createFunc is the function that can create an outer entity based on inner entity config
	// (which is created by 'getTypeFunc')
	createFunc func(config *I) (O, error)

	// resourceReadFunc
	resourceReadFunc func(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics

	getEntityFunc func(id string) (O, error)

	// pre-create hook
	preCreateHooks []beforeCreateHook

	// Delete
	preDeleteHooks []resourceHook[O]
}

func createResource[O updateDeleter[O, I], I any](ctx context.Context, d *schema.ResourceData, meta interface{}, c crudConfig[O, I]) diag.Diagnostics {
	t, err := c.getTypeFunc(d)
	if err != nil {
		return diag.Errorf("error getting %s type: %s", c.entityLabel, err)
	}

	vcdClient := meta.(*VCDClient)
	err = executeBefore(vcdClient, d, c.preCreateHooks)
	if err != nil {
		return diag.Errorf("error executing pre-delete %s hooks: %s", c.entityLabel, err)
	}

	createdEntity, err := c.createFunc(t)
	if err != nil {
		return diag.Errorf("error creating %s: %s", c.entityLabel, err)
	}

	err = c.stateStoreFunc(d, createdEntity)
	if err != nil {
		return diag.Errorf("error storing %s to state: %s", c.entityLabel, err)
	}

	return c.resourceReadFunc(ctx, d, meta)
}

func updateResource[O updateDeleter[O, I], I any](ctx context.Context, d *schema.ResourceData, meta interface{}, c crudConfig[O, I]) diag.Diagnostics {
	t, err := c.getTypeFunc(d)
	if err != nil {
		return diag.Errorf("error getting %s type: %s", c.entityLabel, err)
	}

	retrievedEntity, err := c.getEntityFunc(d.Id())
	if err != nil {
		return diag.Errorf("error getting %s: %s", c.entityLabel, err)
	}

	_, err = retrievedEntity.Update(t)
	if err != nil {
		return diag.Errorf("error storing %s to state: %s", c.entityLabel, err)
	}

	return c.resourceReadFunc(ctx, d, meta)
}

func readResource[O updateDeleter[O, I], I any](_ context.Context, d *schema.ResourceData, _ interface{}, c crudConfig[O, I]) diag.Diagnostics {
	retrievedEntity, err := c.getEntityFunc(d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			util.Logger.Printf("[DEBUG] entity '%s' with ID '%s' not found. Removing from state", c.entityLabel, d.Id())
		}
		return diag.Errorf("error getting %s: %s", c.entityLabel, err)
	}

	err = c.stateStoreFunc(d, retrievedEntity)
	if err != nil {
		return diag.Errorf("error storing %s to state: %s", c.entityLabel, err)
	}

	return nil
}

func readDatasource[O updateDeleter[O, I], I any](_ context.Context, d *schema.ResourceData, _ interface{}, c crudConfig[O, I]) diag.Diagnostics {
	entityName := d.Get("name").(string)
	retrievedEntity, err := c.getEntityFunc(entityName)
	if err != nil {
		return diag.Errorf("error getting %s by Name '%s': %s", c.entityLabel, entityName, err)
	}

	err = c.stateStoreFunc(d, retrievedEntity)
	if err != nil {
		return diag.Errorf("error storing %s to state: %s", c.entityLabel, err)
	}

	return nil
}

func deleteResource[O updateDeleter[O, I], I any](_ context.Context, d *schema.ResourceData, _ interface{}, c crudConfig[O, I]) diag.Diagnostics {
	retrievedEntity, err := c.getEntityFunc(d.Id())
	if err != nil {
		return diag.Errorf("error getting %s: %s", c.entityLabel, err)
	}

	err = executeHooks(retrievedEntity, c.preDeleteHooks)
	if err != nil {
		return diag.Errorf("error executing pre-delete %s hooks: %s", c.entityLabel, err)
	}

	err = retrievedEntity.Delete()
	if err != nil {
		return diag.Errorf("error storing %s to state: %s", c.entityLabel, err)
	}

	return nil
}

func executeBefore(vcdClient *VCDClient, d *schema.ResourceData, runList []beforeCreateHook) error {
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

func executeHooks[O any](outerEntity O, runList []resourceHook[O]) error {
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
