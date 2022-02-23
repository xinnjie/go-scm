// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tencentgit

import (
	"context"
	"encoding/json"
	"github.com/jenkins-x/go-scm/scm"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/h2non/gock.v1"
)

func TestGitFindCommit(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/xinnjie/testme/repository/commits/7fd1a60b01f91b314f59955a4e4d4e80d8edf11d").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/commit.json")

	client := NewDefault()
	got, _, err := client.Git.FindCommit(context.Background(), "xinnjie/testme", "7fd1a60b01f91b314f59955a4e4d4e80d8edf11d")
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Commit)
	raw, _ := ioutil.ReadFile("testdata/commit.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestGitFindBranch(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/xinnjie/testme/repository/branches/master").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/branch.json")

	client := NewDefault()
	got, _, err := client.Git.FindBranch(context.Background(), "xinnjie/testme", "master")
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Reference)
	raw, _ := ioutil.ReadFile("testdata/branch.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestGitFindTag(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/xinnjie/testme/repository/tags/v0.0.1").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/tag.json")

	client := NewDefault()
	got, _, err := client.Git.FindTag(context.Background(), "xinnjie/testme", "v0.0.1")
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Reference)
	raw, _ := ioutil.ReadFile("testdata/tag.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

}

func TestGitListCommits(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Get("api/v3/projects/xinnjie/testme/repository/commits").
		MatchParam("page", "1").
		MatchParam("per_page", "30").
		MatchParam("ref_name", "master").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		SetHeaders(mockPageHeaders).
		File("testdata/commits.json")

	client := NewDefault()
	got, _, err := client.Git.ListCommits(context.Background(), "xinnjie/testme", scm.CommitListOptions{Ref: "master", Page: 1, Size: 30})
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.Commit{}
	raw, _ := ioutil.ReadFile("testdata/commits.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

}

func TestGitListBranches(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Get("api/v3/projects/xinnjie/testme/repository/branches").
		MatchParam("page", "1").
		MatchParam("per_page", "30").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		SetHeaders(mockPageHeaders).
		File("testdata/branches.json")

	client := NewDefault()
	got, _, err := client.Git.ListBranches(context.Background(), "xinnjie/testme", scm.ListOptions{Page: 1, Size: 30})
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.Reference{}
	raw, _ := ioutil.ReadFile("testdata/branches.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

}

func TestGitListTags(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/xinnjie/testme/repository/tags").
		MatchParam("page", "1").
		MatchParam("per_page", "30").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		SetHeaders(mockPageHeaders).
		File("testdata/tags.json")

	client := NewDefault()
	got, _, err := client.Git.ListTags(context.Background(), "xinnjie/testme", scm.ListOptions{Page: 1, Size: 30})
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.Reference{}
	raw, _ := ioutil.ReadFile("testdata/tags.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestGitListChanges(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/xinnjie/testme/repository/commits/dc238bca5c226bb7a2c7cc5bdbfc9930d6abd8b6/diff").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/commit_diff.json")

	client := NewDefault()
	got, _, err := client.Git.ListChanges(context.Background(), "xinnjie/testme", "dc238bca5c226bb7a2c7cc5bdbfc9930d6abd8b6", scm.ListOptions{})
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.Change{}
	raw, _ := ioutil.ReadFile("testdata/commit_diff.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

}

func TestGitCompareCommits(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/xinnjie/testme/repository/compare").
		MatchParam("from", "dc238bca").
		MatchParam("to", "652a91b8").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/compare.json")

	client := NewDefault()
	got, _, err := client.Git.CompareCommits(context.Background(), "xinnjie/testme", "dc238bca", "652a91b8", scm.ListOptions{})
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.Change{}
	raw, _ := ioutil.ReadFile("testdata/compare.json.golden")
	assert.NoError(t, json.Unmarshal(raw, &want))

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

}

func TestGitCreateRef(t *testing.T) {
	baseSHA := "dc238bca"
	defer gock.Off()
	gock.New("https://git.code.tencent.com").
		Post("/api/v3/projects/xinnjie/testme/repository/branches").
		MatchParam("branch_name", "testing").
		MatchParam("ref", baseSHA).
		Reply(http.StatusCreated).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/create_branch.json")

	client := NewDefault()
	got, _, err := client.Git.CreateRef(context.Background(), "xinnjie/testme", "testing", baseSHA)
	if err != nil {
		t.Fatal(err)
	}

	want := &scm.Reference{}
	raw, _ := ioutil.ReadFile("testdata/create_branch.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

}
