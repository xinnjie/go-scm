// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/jenkins-x/go-scm/scm"
)

//
// pull request sub-tests
//

func testPullRequests(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		t.Run("List", testPullRequestList(client))
		t.Run("Find", testPullRequestFind(client))
		t.Run("Changes", testPullRequestChanges(client))
		t.Run("Comments", testPullRequestComments(client))
	}
}

func testPullRequestList(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		updatedAfter := time.Date(2021, 8, 10, 0, 0, 0, 0, time.Local)
		opts := scm.PullRequestListOptions{
			Open:         true,
			Closed:       true,
			UpdatedAfter: &updatedAfter,
		}
		result, _, err := client.PullRequests.List(context.Background(), "xinnjie/testme", opts)
		if err != nil {
			t.Error(err)
		}
		if len(result) == 0 {
			t.Errorf("Got empty pull request list")
		}
		for _, pr := range result {
			if pr.Number == 1 {
				t.Run("PullRequest", testPullRequest(pr))
			}
		}
	}
}

func testPullRequestFind(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		result, _, err := client.PullRequests.Find(context.Background(), "xinnjie/testme", 1)
		if err != nil {
			t.Error(err)
		}
		t.Run("PullRequest", testPullRequest(result))
	}
}

//
// pull request comment sub-tests
//

func testPullRequestComments(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		t.Run("List", testPullRequestCommentFind(client))
		t.Run("Find", testPullRequestCommentList(client))
	}
}

func testPullRequestCommentFind(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		result, _, err := client.PullRequests.FindComment(context.Background(), "xinnjie/testme", 3, 1383976)
		if err != nil {
			t.Error(err)
		}
		t.Run("Comment", testPullRequestComment(result))
	}
}

func testPullRequestCommentList(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		opts := scm.ListOptions{}
		result, _, err := client.PullRequests.ListComments(context.Background(), "xinnjie/testme", 1, opts)
		if err != nil {
			t.Error(err)
		}
		if len(result) == 0 {
			t.Errorf("Got empty pull request comment list")
		}
		for _, comment := range result {
			if comment.ID == 2990882 {
				t.Run("Comment", testPullRequestComment(comment))
			}
		}
	}
}

//
// pull request changes sub-tests
//

func testPullRequestChanges(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		opts := scm.ListOptions{}
		result, _, err := client.PullRequests.ListChanges(context.Background(), "xinnjie/testme", 339869, opts)
		if err != nil {
			t.Error(err)
		}
		if len(result) == 0 {
			t.Errorf("Got empty pull request change list")
			return
		}
		t.Run("File", testChange(result[0]))
	}
}

//
// struct sub-tests
//

func testPullRequest(pr *scm.PullRequest) func(t *testing.T) {
	return func(t *testing.T) {
		if got, want := pr.Number, 1; got != want {
			t.Errorf("Want pr Number %d, got %d", want, got)
		}
		if got, want := pr.Title, "edit README"; got != want {
			t.Errorf("Want pr Title %q, got %q", want, got)
		}
		if got, want := pr.Body, "edit README"; got != want {
			t.Errorf("Want pr Body %q, got %q", want, got)
		}
		if got, want := pr.Source, "mr-test"; got != want {
			t.Errorf("Want pr Source %q, got %q", want, got)
		}
		if got, want := pr.Target, "master"; got != want {
			t.Errorf("Want pr Target %q, got %q", want, got)
		}
		if got, want := pr.Ref, "refs/merge-requests/1/head"; got != want {
			t.Errorf("Want pr Ref %q, got %q", want, got)
		}
		if got, want := pr.Sha, "a398282a168e6919e8ff3806e1fa11f66ab53d30"; got != want {
			t.Errorf("Want pr Sha %q, got %q", want, got)
		}
		if got, want := pr.Link, "https://git.code.tencent.com/xinnjie/testme/merge_requests/1"; got != want {
			t.Errorf("Want pr Link %q, got %q", want, got)
		}
		if got, want := pr.Author.Login, "xinnjie"; got != want {
			t.Errorf("Want pr Author Login %q, got %q", want, got)
		}
		if got, want := pr.Author.Name, "xinnjie"; got != want {
			t.Errorf("Want pr Author Name %q, got %q", want, got)
		}
		if got, want := pr.Author.Avatar, "https://git.code.tencent.com/uploads/user/avatar/176405/7e72ab97e19e430287c34c54af3bf7e1."; got != want {
			t.Errorf("Want pr Author Avatar %q, got %q", want, got)
		}
		if got, want := pr.Closed, true; got != want {
			t.Errorf("Want pr Closed %v, got %v", want, got)
		}
		if got, want := pr.Merged, true; got != want {
			t.Errorf("Want pr Merged %v, got %v", want, got)
		}
		if got, want := pr.Created.Unix(), int64(1629649358); got != want {
			t.Errorf("Want pr Created %d, got %d", want, got)
		}
		if got, want := pr.Updated.Unix(), int64(1629649768); got != want {
			t.Errorf("Want pr Updated %d, got %d", want, got)
		}
		if got, want := pr.State, "closed"; got != want {
			t.Errorf("Want pr state %s, got %s", want, got)
		}
	}
}

func testPullRequestComment(comment *scm.Comment) func(t *testing.T) {
	return func(t *testing.T) {
		if got, want := comment.ID, 2990882; got != want {
			t.Errorf("Want pr comment ID %d, got %d", want, got)
		}
		if got, want := comment.Body, "Status changed to closed"; got != want {
			t.Errorf("Want pr comment Body %q, got %q", want, got)
		}
		if got, want := comment.Author.Login, "dblessing"; got != want {
			t.Errorf("Want pr comment Author Login %q, got %q", want, got)
		}
		if got, want := comment.Author.Name, "Drew Blessing"; got != want {
			t.Errorf("Want pr comment Author Name %q, got %q", want, got)
		}
		if got, want := comment.Author.Avatar, "https://assets.gitlab-static.net/uploads/-/system/user/avatar/13356/avatar.png"; got != want {
			t.Errorf("Want pr comment Author Avatar %q, got %q", want, got)
		}
		if got, want := comment.Created.Unix(), int64(1450463422); got != want {
			t.Errorf("Want pr comment Created %d, got %d", want, got)
		}
		if got, want := comment.Updated.Unix(), int64(1450463422); got != want {
			t.Errorf("Want pr comment Updated %d, got %d", want, got)
		}
	}
}

func testChange(change *scm.Change) func(t *testing.T) {
	return func(t *testing.T) {
		if got, want := change.Path, "README.md"; got != want {
			t.Errorf("Want file change Path %q, got %q", want, got)
		}
		if got, want := change.Added, false; got != want {
			t.Errorf("Want file Added %v, got %v", want, got)
		}
		if got, want := change.Deleted, false; got != want {
			t.Errorf("Want file Deleted %v, got %v", want, got)
		}
		if got, want := change.Renamed, false; got != want {
			t.Errorf("Want file Renamed %v, got %v", want, got)
		}
	}
}
