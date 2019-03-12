/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"strings"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

type EjectTask struct {
	*Task
	vm *VM
}

var timeBetweenRefresh = 3 * time.Second

// Question Message from vCD API
const questionMessage = "Disconnect anyway and override the lock?"

// Creates wrapped Task which is dedicated for eject media functionality and
// provides additional functionality to answer VM questions
func NewEjectTask(task *Task, vm *VM) *EjectTask {
	return &EjectTask{
		task,
		vm,
	}
}

// Checks the status of the task every 3 seconds and returns when the
// eject task is either completed or failed
func (ejectTask *EjectTask) WaitTaskCompletion(isAnswerYes bool) error {
	return ejectTask.WaitInspectTaskCompletion(isAnswerYes, timeBetweenRefresh)
}

// function which handles answers for ejecting
func (ejectTask *EjectTask) WaitInspectTaskCompletion(isAnswerYes bool, delay time.Duration) error {

	if ejectTask.Task == nil {
		return fmt.Errorf("cannot refresh, Object is empty")
	}

	for {
		err := ejectTask.Refresh()
		if err != nil {
			return fmt.Errorf("error retrieving task: %s", err)
		}

		// If task is not in a waiting status we're done, check if there's an error and return it.
		if ejectTask.Task.Task.Status != "queued" && ejectTask.Task.Task.Status != "preRunning" && ejectTask.Task.Task.Status != "running" {
			if ejectTask.Task.Task.Status == "error" {
				return fmt.Errorf("task did not complete succesfully: %s", ejectTask.Task.Task.Error.Message)
			}
			return nil
		}

		question, err := ejectTask.vm.GetQuestion()
		if err != nil {
			return fmt.Errorf("task did not complete succesfully: %s, quering question for VM failed: %s", ejectTask.Task.Task.Description, err.Error())
		}

		if question.QuestionId != "" && strings.Contains(question.Question, questionMessage) {
			var choiceToUse *types.VmQuestionAnswerChoiceType
			for _, choice := range question.Choices {
				if isAnswerYes {
					if strings.Contains(choice.Text, "yes") {
						choiceToUse = choice
					}
				} else {
					if strings.Contains(choice.Text, "no") {
						choiceToUse = choice
					}
				}
			}

			if choiceToUse != nil {
				err = ejectTask.vm.AnswerQuestion(question.QuestionId, choiceToUse.Id)
				if err != nil {
					return fmt.Errorf("task did not complete succesfully: %s, answering question for eject in VM failed: %s", ejectTask.Task.Task.Description, err.Error())
				}
			}

		}

		// Sleep for a given period and try again.
		time.Sleep(delay)
	}
}
