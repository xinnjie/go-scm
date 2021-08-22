// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package integration

import (
	"context"
	"testing"

	"github.com/jenkins-x/go-scm/scm"
)

//
// git sub-tests
//

func testGit(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		t.Run("Branches", testBranches(client))
		t.Run("Commits", testCommits(client))
		t.Run("Tags", testTags(client))
	}
}

//
// branch sub-tests
//

func testBranches(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		t.Run("Find", testBranchFind(client))
		t.Run("List", testBranchList(client))
	}
}

func testBranchFind(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		result, _, err := client.Git.FindBranch(context.Background(), "xinnjie/testme", "feature")
		if err != nil {
			t.Error(err)
			return
		}
		t.Run("Branch", testBranch(result))
	}
}

func testBranchList(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		opts := scm.ListOptions{}
		result, _, err := client.Git.ListBranches(context.Background(), "xinnjie/testme", opts)
		if err != nil {
			t.Error(err)
			return
		}
		if len(result) == 0 {
			t.Errorf("Want a non-empty branch list")
		}
		for _, branch := range result {
			if branch.Name == "feature" {
				t.Run("Branch", testBranch(branch))
			}
		}
	}
}

//
// branch sub-tests
//

func testTags(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		t.Run("Find", testTagFind(client))
		t.Run("List", testTagList(client))
	}
}

func testTagFind(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		result, _, err := client.Git.FindTag(context.Background(), "xinnjie/testme", "v1.1.0")
		if err != nil {
			t.Error(err)
			return
		}
		t.Run("Tag", testTag(result))
	}
}

func testTagList(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		opts := scm.ListOptions{}
		result, _, err := client.Git.ListTags(context.Background(), "xinnjie/testme", opts)
		if err != nil {
			t.Error(err)
			return
		}
		if len(result) == 0 {
			t.Errorf("Want a non-empty tag list")
		}
		for _, tag := range result {
			if tag.Name == "v1.1.0" {
				t.Run("Tag", testTag(tag))
			}
		}
	}
}

//
// commit sub-tests
//

func testCommits(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		t.Run("Find", testCommitFind(client))
		t.Run("List", testCommitList(client))
	}
}

func testCommitFind(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		result, _, err := client.Git.FindCommit(context.Background(), "xinnjie/testme", "dc238bca5c226bb7a2c7cc5bdbfc9930d6abd8b6")
		if err != nil {
			t.Error(err)
			return
		}
		t.Run("Commit", testCommit(result))
	}
}

func testCommitList(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		opts := scm.CommitListOptions{
			Ref: "master",
		}
		result, _, err := client.Git.ListCommits(context.Background(), "xinnjie/testme", opts)
		if err != nil {
			t.Error(err)
			return
		}
		if len(result) == 0 {
			t.Errorf("Want a non-empty commit list")
		}
		for _, commit := range result {
			if commit.Sha == "dc238bca5c226bb7a2c7cc5bdbfc9930d6abd8b6" {
				t.Run("Commit", testCommit(commit))
			}
		}
	}
}

//
// change sub-tests
//

func testChangeList(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		result, _, err := client.Git.ListChanges(context.Background(), "xinnjie/testme", "c5895070235cadd2d839136dad79e01838ee2de1", scm.ListOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		if len(result) == 0 {
			t.Errorf("Want a non-empty change list")
		}
		for _, change := range result {
			t.Run("Change", testCommitChange(change))
		}
	}
}

func testCompareCommits(client *scm.Client) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		result, _, err := client.Git.CompareCommits(context.Background(), "xinnjie/testme", "6da006adb7cafe15b8495e3b7811fc318e485553", "c5895070235cadd2d839136dad79e01838ee2de1", scm.ListOptions{})
		if err != nil {
			t.Error(err)
			return
		}
		if len(result) == 0 {
			t.Errorf("Want a non-empty change list")
		}
		for _, change := range result {
			t.Run("Change", testCommitChange(change))
		}
	}
}

//
// struct sub-tests
//

func testBranch(branch *scm.Reference) func(t *testing.T) {
	return func(t *testing.T) {
		if got, want := branch.Name, "feature"; got != want {
			t.Errorf("Want branch Name %q, got %q", want, got)
		}
		if got, want := branch.Sha, "0b4bc9a49b562e85de7cc9e834518ea6828729b9"; got != want {
			t.Errorf("Want branch Avatar %q, got %q", want, got)
		}
	}
}

func testTag(tag *scm.Reference) func(t *testing.T) {
	return func(t *testing.T) {
		if got, want := tag.Name, "v1.1.0"; got != want {
			t.Errorf("Want tag Name %q, got %q", want, got)
		}
		if got, want := tag.Sha, "5937ac0a7beb003549fc5fd26fc247adbce4a52e"; got != want {
			t.Errorf("Want tag Avatar %q, got %q", want, got)
		}
	}
}

func testCommit(commit *scm.Commit) func(t *testing.T) {
	return func(t *testing.T) {
		if got, want := commit.Message, "add README"; got != want {
			t.Errorf("Want commit Message %q, got %q", want, got)
		}
		if got, want := commit.Sha, "dc238bca5c226bb7a2c7cc5bdbfc9930d6abd8b6"; got != want {
			t.Errorf("Want commit Sha %q, got %q", want, got)
		}
		if got, want := commit.Author.Name, "xinnjie"; got != want {
			t.Errorf("Want commit author Name %q, got %q", want, got)
		}
		if got, want := commit.Author.Email, "wx_1600e990b7cb469a97087b334c94acee@git.code.tencent.com"; got != want {
			t.Errorf("Want commit author Email %q, got %q", want, got)
		}
		if got, want := commit.Author.Date.Unix(), int64(1629456348); got != want {
			t.Errorf("Want commit author Date %d, got %d", want, got)
		}
		if got, want := commit.Committer.Name, "xinnjie"; got != want {
			t.Errorf("Want commit author Name %q, got %q", want, got)
		}
		if got, want := commit.Committer.Email, "wx_1600e990b7cb469a97087b334c94acee@git.code.tencent.com"; got != want {
			t.Errorf("Want commit author Email %q, got %q", want, got)
		}
		if got, want := commit.Committer.Date.Unix(), int64(1629456348); got != want {
			t.Errorf("Want commit author Date %d, got %d", want, got)
		}
		if got, want := commit.Link, "https://git.code.tencent.com/xinnjie/testme/commit/dc238bca5c226bb7a2c7cc5bdbfc9930d6abd8b6"; got != want {
			t.Errorf("Want commit link %q, got %q", want, got)
		}
	}
}

func testCommitChange(change *scm.Change) func(t *testing.T) {
	return func(t *testing.T) {
		if got, want := change.Path, "config/config.prod.json"; got != want {
			t.Errorf("Want commit Path %q, got %q", want, got)
		}
		if got, want := change.PreviousPath, "config/config.prod.json"; got != want {
			t.Errorf("Want commit PreviousPath %q, got %q", want, got)
		}
		if got, want := change.Added, false; got != want {
			t.Errorf("Want commit Added %t, got %t", want, got)
		}
		if got, want := change.Renamed, false; got != want {
			t.Errorf("Want commit Renamed %t, got %t", want, got)
		}
		if got, want := change.Deleted, false; got != want {
			t.Errorf("Want commit Deleted %t, got %t", want, got)
		}
	}
}
