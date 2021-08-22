// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tencentgit

import (
	"context"
	"testing"

	"github.com/jenkins-x/go-scm/scm"
)

func TestReviewFind(t *testing.T) {
	service := new(reviewService)
	_, _, err := service.Find(context.Background(), "xinnjie/testme", 1, 1)
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestReviewList(t *testing.T) {
	service := new(reviewService)
	_, _, err := service.List(context.Background(), "xinnjie/testme", 1, scm.ListOptions{})
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestReviewCreate(t *testing.T) {
	service := new(reviewService)
	_, _, err := service.Create(context.Background(), "xinnjie/testme", 1, nil)
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestReviewDelete(t *testing.T) {
	service := new(reviewService)
	_, err := service.Delete(context.Background(), "xinnjie/testme", 1, 1)
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}
