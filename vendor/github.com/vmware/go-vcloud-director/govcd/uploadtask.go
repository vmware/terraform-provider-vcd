/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"time"
)

type UploadTask struct {
	uploadProgress *float64
	*Task
}

// Creates wrapped Task which is dedicated for upload functionality and
// provides additional functionality to monitor upload progress.
func NewUploadTask(task *Task, uploadProgress *float64) *UploadTask {
	return &UploadTask{
		uploadProgress,
		task,
	}
}

func (uploadTask *UploadTask) GetUploadProgress() string {
	return fmt.Sprintf("%.2f", *uploadTask.uploadProgress)
}

func (uploadTask *UploadTask) ShowUploadProgress() {
	fmt.Printf("Waiting...")
	for {
		fmt.Printf("\rUpload progress %.2f%%", *uploadTask.uploadProgress)
		if *uploadTask.uploadProgress == 100.00 {
			break
		}
		time.Sleep(1 * time.Second)
	}
}
