/*
 * Copyright 2014 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"
	"time"

	types "github.com/vmware/go-vcloud-director/types/v56"
)

type Task struct {
	Task   *types.Task
	client *Client
}

func NewTask(cli *Client) *Task {
	return &Task{
		Task:   new(types.Task),
		client: cli,
	}
}

func (task *Task) Refresh() error {

	if task.Task == nil {
		return fmt.Errorf("cannot refresh, Object is empty")
	}

	refreshUrl, _ := url.ParseRequestURI(task.Task.HREF)

	req := task.client.NewRequest(map[string]string{}, "GET", *refreshUrl, nil)

	resp, err := checkResp(task.client.Http.Do(req))
	if err != nil {
		return fmt.Errorf("error retrieving task: %s", err)
	}

	// Empty struct before a new unmarshal, otherwise we end up with duplicate
	// elements in slices.
	task.Task = &types.Task{}

	if err = decodeBody(resp, task.Task); err != nil {
		return fmt.Errorf("error decoding task response: %s", err)
	}

	// The request was successful
	return nil
}

func (task *Task) WaitTaskCompletion() error {

	if task.Task == nil {
		return fmt.Errorf("cannot refresh, Object is empty")
	}

	for {
		err := task.Refresh()
		if err != nil {
			return fmt.Errorf("error retreiving task: %s", err)
		}

		// If task is not in a waiting status we're done, check if there's an error and return it.
		if task.Task.Status != "queued" && task.Task.Status != "preRunning" && task.Task.Status != "running" {
			if task.Task.Status == "error" {
				return fmt.Errorf("task did not complete succesfully: %s", task.Task.Description)
			}
			return nil
		}

		// Sleep for 3 seconds and try again.
		time.Sleep(3 * time.Second)
	}
}
