// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tencentgit

import (
	"context"
	"encoding/json"
	"github.com/google/go-cmp/cmp"
	"github.com/jenkins-x/go-scm/scm"
	"gopkg.in/h2non/gock.v1"
	"io/ioutil"
	"testing"
)

func TestPullFind(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/179129").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/repo.json")

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/179129").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/repo.json")

	gock.New("https://git.code.tencent.com").
		Get("api/v3/projects/xinnjie/testme/merge_request/iid/1").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/merge.json")

	client := NewDefault()
	got, _, err := client.PullRequests.Find(context.Background(), "xinnjie/testme", 1)
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.PullRequest)
	raw, err := ioutil.ReadFile("testdata/merge.json.golden")
	if err != nil {
		t.Fatalf("ioutil.ReadFile: %v", err)
	}
	if err := json.Unmarshal(raw, want); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

}

func TestPullList(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Persist().
		Get("/api/v3/projects/179129").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/repo.json")
	// FIXME(xinnjie) gock not match request after add updatedAfter param
	//updatedAfter, _ := time.Parse(timeFormat, "2019-03-25T00:10:19+0000")
	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/xinnjie/testme/merge_requests").
		MatchParam("page", "1").
		MatchParam("per_page", "30").
		//MatchParam("updated_after", "2019-03-25T00:10:19+0000").
		Reply(200).
		Type("application/json").
		File("testdata/merges.json")

	client := NewDefault()
	got, _, err := client.PullRequests.List(context.Background(), "xinnjie/testme", scm.PullRequestListOptions{
		Page:   1,
		Size:   30,
		Open:   true,
		Closed: true,
		//UpdatedAfter: &updatedAfter,
	})
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.PullRequest{}
	raw, err := ioutil.ReadFile("testdata/merges.json.golden")
	if err != nil {
		t.Fatalf("ioutil.ReadFile: %v", err)
	}
	if err := json.Unmarshal(raw, &want); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestPullListChanges(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/xinnjie/testme/merge_requests/339869/changes").
		Reply(200).
		Type("application/json").
		File("testdata/merge_diff.json")

	client := NewDefault()
	got, _, err := client.PullRequests.ListChanges(context.Background(), "xinnjie/testme", 339869, scm.ListOptions{Page: 1, Size: 30})
	if err != nil {
		t.Error(err)
		return
	}
	want := []*scm.Change{}
	raw, _ := ioutil.ReadFile("testdata/merge_diff.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

}

func TestPullMerge(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Put("/api/v3/projects/xinnjie/testme/merge_requests/339869/merge").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders)

	client := NewDefault()
	_, err := client.PullRequests.Merge(context.Background(), "xinnjie/testme", 339869, nil)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestPullClose(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Put("/api/v3/projects/xinnjie/testme/merge_requests/1347").
		MatchParam("state_event", "closed").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders)

	client := NewDefault()
	_, err := client.PullRequests.Close(context.Background(), "xinnjie/testme", 1347)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestPullReopen(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Put("/api/v3/projects/xinnjie/testme/merge_requests/1347").
		MatchParam("state_event", "reopen").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders)

	client := NewDefault()
	_, err := client.PullRequests.Reopen(context.Background(), "xinnjie/testme", 1347)
	if err != nil {
		t.Error(err)
		return
	}
}

func TestPullCommentFind(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/xinnjie/testme/merge_requests/2/notes/1").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/merge_note.json")

	client := NewDefault()
	got, _, err := client.PullRequests.FindComment(context.Background(), "xinnjie/testme", 2, 1)
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Comment)
	raw, _ := ioutil.ReadFile("testdata/merge_note.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestPullListComments(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/xinnjie/testme/merge_requests/1/notes").
		MatchParam("page", "1").
		MatchParam("per_page", "30").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		SetHeaders(mockPageHeaders).
		File("testdata/merge_notes.json")

	client := NewDefault()
	got, res, err := client.PullRequests.ListComments(context.Background(), "xinnjie/testme", 1, scm.ListOptions{Size: 30, Page: 1})
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.Comment{}
	raw, _ := ioutil.ReadFile("testdata/merge_notes.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Page", testPage(res))
}

func TestPullCreateComment(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Post("/api/v3/projects/xinnjie/testme/merge_requests/1/notes").
		MatchParam("body", "lgtm").
		Reply(201).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/merge_note.json")

	input := &scm.CommentInput{
		Body: "lgtm",
	}

	client := NewDefault()
	got, _, err := client.PullRequests.CreateComment(context.Background(), "xinnjie/testme", 1, input)
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Comment)
	raw, _ := ioutil.ReadFile("testdata/merge_note.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

}

func TestPullCommentDelete(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Delete("/api/v3/projects/xinnjie/testme/merge_requests/2/notes/1").
		Reply(204).
		Type("application/json").
		SetHeaders(mockHeaders)

	client := NewDefault()
	_, err := client.PullRequests.DeleteComment(context.Background(), "xinnjie/testme", 2, 1)
	if err != nil {
		t.Error(err)
		return
	}

}

func TestPullEditComment(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Put("/api/v3/projects/xinnjie/testme/merge_requests/2/notes/1").
		File("testdata/edit_issue_note.json").
		Reply(201).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/merge_note.json")

	input := &scm.CommentInput{
		Body: "closed",
	}

	client := NewDefault()
	got, _, err := client.PullRequests.EditComment(context.Background(), "xinnjie/testme", 2, 1, input)
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Comment)
	raw, _ := ioutil.ReadFile("testdata/merge_note.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

}

func TestPullCreate(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/32732").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/repo.json")

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/2").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/other_repo.json")

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/2").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/other_repo.json")

	gock.New("https://git.code.tencent.com").
		Post("/api/v3/projects/xinnjie/testme/merge_requests").
		Reply(201).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/pr_create.json")

	input := &scm.PullRequestInput{
		Title: "Amazing new feature",
		Body:  "Please pull these awesome changes in!",
		Head:  "test1",
		Base:  "master",
	}

	client := NewDefault()
	got, _, err := client.PullRequests.Create(context.Background(), "xinnjie/testme", input)
	if err != nil {
		t.Fatal(err)
	}

	want := new(scm.PullRequest)
	raw, _ := ioutil.ReadFile("testdata/pr_create.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

}

func TestPullUpdate(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/32732").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/repo.json")

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/2").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/other_repo.json")

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/2").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/other_repo.json")

	gock.New("https://git.code.tencent.com").
		Put("/api/v3/projects/xinnjie/testme/merge_requests/1").
		File("testdata/pr_update.json").
		Reply(201).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/pr_create.json")

	input := &scm.PullRequestInput{
		Title: "A new title",
		Body:  "A new description",
	}

	client := NewDefault()
	got, _, err := client.PullRequests.Update(context.Background(), "xinnjie/testme", 1, input)
	if err != nil {
		t.Fatal(err)
	}

	want := new(scm.PullRequest)
	raw, _ := ioutil.ReadFile("testdata/pr_create.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

}

func TestPullListEvents(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/xinnjie/testme/merge_requests/28/resource_label_events").
		MatchParam("page", "1").
		MatchParam("per_page", "30").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		SetHeaders(mockPageHeaders).
		File("testdata/pr_events.json")

	client := NewDefault()
	got, res, err := client.PullRequests.ListEvents(context.Background(), "xinnjie/testme", 28, scm.ListOptions{Size: 30, Page: 1})
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.ListedIssueEvent{}
	raw, _ := ioutil.ReadFile("testdata/pr_events.golden.json")
	err = json.Unmarshal(raw, &want)
	if err != nil {
		t.Error(err)
		return
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Page", testPage(res))
}

func TestPullSetMilestone(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/32732").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/repo.json")

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/2").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/other_repo.json")

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/2").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/other_repo.json")

	gock.New("https://git.code.tencent.com").
		Put("/api/v3/projects/xinnjie/testme/merge_requests/1").
		File("testdata/issue_or_pr_set_milestone.json").
		Reply(201).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/pr_create.json")

	client := NewDefault()
	_, err := client.PullRequests.SetMilestone(context.Background(), "xinnjie/testme", 1, 1)
	if err != nil {
		t.Fatal(err)
	}
}

func TestPullClearMilestone(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/32732").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/repo.json")

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/2").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/other_repo.json")

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/2").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/other_repo.json")

	gock.New("https://git.code.tencent.com").
		Put("/api/v3/projects/xinnjie/testme/merge_requests/1").
		File("testdata/issue_or_pr_clear_milestone.json").
		Reply(201).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/pr_create.json")

	client := NewDefault()
	_, err := client.PullRequests.ClearMilestone(context.Background(), "xinnjie/testme", 1)
	if err != nil {
		t.Fatal(err)
	}
}
