package tencentgit

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jenkins-x/go-scm/scm"
	"gopkg.in/h2non/gock.v1"
)

func TestMilestoneFind(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/xinnjie/testme/milestones/1").
		Reply(200).
		Type("application/json").
		File("testdata/milestone.json")

	client := NewDefault()
	got, _, err := client.Milestones.Find(context.Background(), "xinnjie/testme", 1)
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Milestone)
	raw, err := ioutil.ReadFile("testdata/milestone.json.golden")
	if err != nil {
		t.Fatalf("ioutil.ReadFile: %v", err)
	}
	if err := json.Unmarshal(raw, want); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

}

func TestMilestoneList(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/xinnjie/testme/milestones").
		Reply(200).
		Type("application/json").
		File("testdata/milestones.json")

	client := NewDefault()
	got, _, err := client.Milestones.List(context.Background(), "xinnjie/testme", scm.MilestoneListOptions{})
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.Milestone{}
	raw, err := ioutil.ReadFile("testdata/milestones.json.golden")
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

func TestMilestoneCreate(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Post("/api/v3/projects/xinnjie/testme/milestones").
		File("testdata/milestone_create.json").
		Reply(200).
		Type("application/json").
		File("testdata/milestone.json")

	client := NewDefault()
	dueDate, _ := time.Parse(scm.SearchTimeFormat, "2012-10-09T23:39:01Z")
	input := &scm.MilestoneInput{
		Title:       "v1.0",
		Description: "Tracking milestone for version 1.0",
		State:       "open",
		DueDate:     &dueDate,
	}
	got, _, err := client.Milestones.Create(context.Background(), "xinnjie/testme", input)
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Milestone)
	raw, err := ioutil.ReadFile("testdata/milestone.json.golden")
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

func TestMilestoneUpdate(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Patch("/api/v3/projects/xinnjie/testme/milestones/1").
		File("testdata/milestone_update.json").
		Reply(200).
		Type("application/json").
		File("testdata/milestone.json")

	client := NewDefault()
	dueDate, _ := time.Parse(scm.SearchTimeFormat, "2012-10-09T23:39:01Z")
	input := &scm.MilestoneInput{
		Title:       "v1.0",
		Description: "Tracking milestone for version 1.0",
		State:       "close",
		DueDate:     &dueDate,
	}
	got, _, err := client.Milestones.Update(context.Background(), "xinnjie/testme", 1, input)
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Milestone)
	raw, err := ioutil.ReadFile("testdata/milestone.json.golden")
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

func TestMilestoneDelete(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Delete("/api/v3/projects/xinnjie/testme/milestones/1").
		Reply(200).
		Type("application/json")
	client := NewDefault()
	_, err := client.Milestones.Delete(context.Background(), "xinnjie/testme", 1)
	if err != nil {
		t.Error(err)
		return
	}
}
