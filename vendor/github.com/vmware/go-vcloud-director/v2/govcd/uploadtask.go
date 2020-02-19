/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"time"
)

type UploadTask struct {
	uploadProgress *mutexedProgress
	*Task
	uploadError *error
}

// Creates wrapped Task which is dedicated for upload functionality and
// provides additional functionality to monitor upload progress.
func NewUploadTask(task *Task, uploadProgress *mutexedProgress, uploadError *error) *UploadTask {
	return &UploadTask{
		uploadProgress,
		task,
		uploadError,
	}
}

func (uploadTask *UploadTask) GetUploadProgress() string {
	return fmt.Sprintf("%.2f", uploadTask.uploadProgress.LockedGet())
}

func (uploadTask *UploadTask) ShowUploadProgress() error {
	fmt.Printf("Waiting...")

	for {
		if *uploadTask.uploadError != nil {
			return *uploadTask.uploadError
		}

		fmt.Printf("\rUpload progress %.2f%%", uploadTask.uploadProgress.LockedGet())
		if uploadTask.uploadProgress.LockedGet() == 100.00 {
			break
		}
		time.Sleep(1 * time.Second)
	}
	return nil
}

func (uploadTask *UploadTask) GetUploadError() error {
	return *uploadTask.uploadError
}
