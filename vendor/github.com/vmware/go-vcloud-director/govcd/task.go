/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/vmware/go-vcloud-director/types/v56"
	"github.com/vmware/go-vcloud-director/util"
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

// If the error is not nil, composes an error message
// made of the error itself + the information from the task's Error component.
// See:
//    https://code.vmware.com/apis/220/vcloud#/doc/doc/types/TaskType.html
//    https://code.vmware.com/apis/220/vcloud#/doc/doc/types/ErrorType.html
func (task *Task) getErrorMessage(err error) string {
	errorMessage := ""
	if err != nil {
		errorMessage = err.Error()
	}
	if task.Task.Error != nil {
		errorMessage += " [" +
			fmt.Sprintf("%d:%s",
				task.Task.Error.MajorErrorCode,   // The MajorError is a numeric code
				task.Task.Error.MinorErrorCode) + // The MinorError is a string with a generic definition of the error
			"] - " + task.Task.Error.Message
	}
	return errorMessage
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
		return fmt.Errorf("error decoding task response: %s", task.getErrorMessage(err))
	}

	// The request was successful
	return nil
}

// This callback function can be passed to task.WaitInspectTaskCompletion
// to perform user defined operations
// * task is the task object being processed
// * howManyTimes is the number of times the task has been refreshed
// * elapsed is how much time since the task was initially processed
// * first is true if this is the first refresh of the task
// * last is true if the function is being called for the last time.
type InspectionFunc func(task *types.Task, howManyTimes int, elapsed time.Duration, first, last bool)

// Customizable version of WaitTaskCompletion.
// Users can define the sleeping duration and an optional callback function for
// extra monitoring.
func (task *Task) WaitInspectTaskCompletion(inspectionFunc InspectionFunc, delay time.Duration) error {

	if task.Task == nil {
		return fmt.Errorf("cannot refresh, Object is empty")
	}

	howManyTimesRefreshed := 0
	startTime := time.Now()
	for {
		howManyTimesRefreshed++
		elapsed := time.Since(startTime)
		err := task.Refresh()
		if err != nil {
			return fmt.Errorf("error retrieving task: %s", err)
		}

		// If an inspection function is provided, we pass information about the task processing:
		// * the task itself
		// * the number of iterations
		// * how much time we have spent querying the task so far
		// * whether this is the first iteration
		// * whether this is the last iteration
		// It's up to the inspection function to render this information fittingly.

		// If task is not in a waiting status we're done, check if there's an error and return it.
		if task.Task.Status != "queued" && task.Task.Status != "preRunning" && task.Task.Status != "running" {
			if inspectionFunc != nil {
				inspectionFunc(task.Task,
					howManyTimesRefreshed,
					elapsed,
					howManyTimesRefreshed == 1,                                   // first
					task.Task.Status == "error" || task.Task.Status == "success", // last
				)
			}
			if task.Task.Status == "error" {
				return fmt.Errorf("task did not complete successfully: %s", task.getErrorMessage(err))
			}
			return nil
		}

		if inspectionFunc != nil {
			inspectionFunc(task.Task,
				howManyTimesRefreshed,
				elapsed,
				howManyTimesRefreshed == 1, // first
				false, // last
			)
		}

		// Sleep for a given period and try again.
		time.Sleep(delay)
	}
}

// Checks the status of the task every 3 seconds and returns when the
// task is either completed or failed
func (task *Task) WaitTaskCompletion() error {
	return task.WaitInspectTaskCompletion(nil, 3*time.Second)
}

func (task *Task) GetTaskProgress() (string, error) {
	if task.Task == nil {
		return "", fmt.Errorf("cannot refresh, Object is empty")
	}

	err := task.Refresh()
	if err != nil {
		return "", fmt.Errorf("error retreiving task: %s", err)
	}

	if task.Task.Status == "error" {
		return "", fmt.Errorf("task did not complete successfully: %s", task.getErrorMessage(err))
	}

	return strconv.Itoa(task.Task.Progress), nil
}

func (task *Task) CancelTask() error {
	cancelTaskURL, err := url.ParseRequestURI(task.Task.HREF + "/action/cancel")
	if err != nil {
		util.Logger.Printf("[Error] Error cancelling task %v: %s", cancelTaskURL.String(), err)
		return err
	}

	request := task.client.NewRequest(map[string]string{}, "POST", *cancelTaskURL, nil)
	_, err = checkResp(task.client.Http.Do(request))
	if err != nil {
		util.Logger.Printf("[Error] Error cancelling task  %v: %s", cancelTaskURL.String(), err)
		return err
	}
	return nil
}
