// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tencentgit

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/jenkins-x/go-scm/scm"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/h2non/gock.v1"
)

func TestIssueFind(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/xinnjie/testme/issues/1").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/issue.json")

	client := NewDefault()
	got, res, err := client.Issues.Find(context.Background(), "xinnjie/testme", 1)
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Issue)
	raw, _ := ioutil.ReadFile("testdata/issue.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestIssueCommentFind(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/xinnjie/testme/issues/2/notes/1").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/issue_note.json")

	client := NewDefault()
	got, res, err := client.Issues.FindComment(context.Background(), "xinnjie/testme", 2, 1)
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Comment)
	raw, _ := ioutil.ReadFile("testdata/issue_note.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestIssueList(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/xinnjie/testme/issues").
		MatchParam("page", "1").
		MatchParam("per_page", "30").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		SetHeaders(mockPageHeaders).
		File("testdata/issues.json")

	client := NewDefault()
	got, res, err := client.Issues.List(context.Background(), "xinnjie/testme", scm.IssueListOptions{Page: 1, Size: 30, Open: true, Closed: true})
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.Issue{}
	raw, _ := ioutil.ReadFile("testdata/issues.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Page", testPage(res))
}

func TestIssueListComments(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/xinnjie/testme/issues/1/notes").
		MatchParam("page", "1").
		MatchParam("per_page", "30").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		SetHeaders(mockPageHeaders).
		File("testdata/issue_notes.json")

	client := NewDefault()
	got, res, err := client.Issues.ListComments(context.Background(), "xinnjie/testme", 1, scm.ListOptions{Size: 30, Page: 1})
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.Comment{}
	raw, _ := ioutil.ReadFile("testdata/issue_notes.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
	t.Run("Page", testPage(res))
}

func TestIssueListEvents(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/xinnjie/testme/issues/253/resource_label_events").
		MatchParam("page", "1").
		MatchParam("per_page", "30").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		SetHeaders(mockPageHeaders).
		File("testdata/issue_events.json")

	client := NewDefault()
	got, res, err := client.Issues.ListEvents(context.Background(), "xinnjie/testme", 253, scm.ListOptions{Size: 30, Page: 1})
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.ListedIssueEvent{}
	raw, _ := ioutil.ReadFile("testdata/issue_events.golden.json")
	err = json.Unmarshal(raw, &want)
	if err != nil {
		t.Error(err)
		return
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
	t.Run("Page", testPage(res))
}

func TestIssueCreate(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Post("/api/v3/projects/xinnjie/testme/issues").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/issue.json")

	input := scm.IssueInput{
		Title: "Found a bug",
		Body:  "I'm having a problem with this.",
	}

	client := NewDefault()
	got, res, err := client.Issues.Create(context.Background(), "xinnjie/testme", &input)
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Issue)
	raw, _ := ioutil.ReadFile("testdata/issue.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestIssueCreateComment(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Post("/api/v3/projects/xinnjie/testme/issues/1/notes").
		MatchParam("body", "lgtm").
		Reply(201).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/issue_note.json")

	input := &scm.CommentInput{
		Body: "lgtm",
	}

	client := NewDefault()
	got, res, err := client.Issues.CreateComment(context.Background(), "xinnjie/testme", 1, input)
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Comment)
	raw, _ := ioutil.ReadFile("testdata/issue_note.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestIssueCommentDelete(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Delete("/api/v3/projects/xinnjie/testme/issues/2/notes/1").
		Reply(204).
		Type("application/json").
		SetHeaders(mockHeaders)

	client := NewDefault()
	res, err := client.Issues.DeleteComment(context.Background(), "xinnjie/testme", 2, 1)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestIssueEditComment(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Put("/api/v3/projects/xinnjie/testme/issues/2/notes/1").
		File("testdata/edit_issue_note.json").
		Reply(201).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/issue_note.json")

	input := &scm.CommentInput{
		Body: "closed",
	}

	client := NewDefault()
	got, res, err := client.Issues.EditComment(context.Background(), "xinnjie/testme", 2, 1, input)
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Comment)
	raw, _ := ioutil.ReadFile("testdata/issue_note.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestIssueClose(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Put("/api/v3/projects/xinnjie/testme/issues/1").
		MatchParam("state_event", "close").
		Reply(204).
		Type("application/json").
		SetHeaders(mockHeaders)

	client := NewDefault()
	res, err := client.Issues.Close(context.Background(), "xinnjie/testme", 1)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestIssueReopen(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Put("/api/v3/projects/xinnjie/testme/issues/1").
		MatchParam("state_event", "reopen").
		Reply(204).
		Type("application/json").
		SetHeaders(mockHeaders)

	client := NewDefault()
	res, err := client.Issues.Reopen(context.Background(), "xinnjie/testme", 1)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestIssueLock(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Put("/api/v3/projects/xinnjie/testme/issues/1").
		MatchParam("discussion_locked", "true").
		Reply(204).
		Type("application/json").
		SetHeaders(mockHeaders)

	client := NewDefault()
	res, err := client.Issues.Lock(context.Background(), "xinnjie/testme", 1)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestIssueUnlock(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Put("/api/v3/projects/xinnjie/testme/issues/1").
		MatchParam("discussion_locked", "false").
		Reply(204).
		Type("application/json").
		SetHeaders(mockHeaders)

	client := NewDefault()
	res, err := client.Issues.Unlock(context.Background(), "xinnjie/testme", 1)
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestIssueSetMilestone(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Put("/api/v3/projects/xinnjie/testme/issues/1").
		File("testdata/issue_or_pr_set_milestone.json").
		Reply(201).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/issue.json")

	client := NewDefault()
	_, err := client.Issues.SetMilestone(context.Background(), "xinnjie/testme", 1, 1)
	if err != nil {
		t.Fatal(err)
	}
}

func TestIssueClearMilestone(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Put("/api/v3/projects/xinnjie/testme/issues/1").
		File("testdata/issue_or_pr_clear_milestone.json").
		Reply(201).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/issue.json")

	client := NewDefault()
	_, err := client.Issues.ClearMilestone(context.Background(), "xinnjie/testme", 1)
	if err != nil {
		t.Fatal(err)
	}
}
