package tencentgit

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jenkins-x/go-scm/scm"
	"github.com/stretchr/testify/assert"
	"gopkg.in/h2non/gock.v1"
)

func TestReleaseFindByTag(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/xinnjie/testme/releases/v1.0.1").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/release.json")

	client := NewDefault()
	got, res, err := client.Releases.FindByTag(context.Background(), "xinnjie/testme", "v1.0.1")
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Release)
	raw, _ := ioutil.ReadFile("testdata/release.json.golden")
	err = json.Unmarshal(raw, want)
	assert.NoError(t, err)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)

		data, _ := json.Marshal(got)
		t.Log("got JSON:")
		t.Log(string(data))
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestReleaseList(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/projects/xinnjie/testme/releases").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/releases.json")

	client := NewDefault()
	got, res, err := client.Releases.List(context.Background(), "xinnjie/testme", scm.ReleaseListOptions{})
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.Release{}
	raw, _ := ioutil.ReadFile("testdata/releases.json.golden")
	err = json.Unmarshal(raw, &want)
	assert.NoError(t, err)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)

		data, _ := json.Marshal(got)
		t.Log("got JSON:")
		t.Log(string(data))
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestReleaseCreate(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Post("/api/v3/projects/xinnjie/testme/releases").
		File("testdata/release_create.json").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/release.json")

	client := NewDefault()
	input := &scm.ReleaseInput{
		Title:       "v1.0",
		Description: "Tracking release for version 1.0",
		Tag:         "v1.0",
	}

	got, res, err := client.Releases.Create(context.Background(), "xinnjie/testme", input)
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Release)
	raw, _ := ioutil.ReadFile("testdata/release.json.golden")
	err = json.Unmarshal(raw, want)
	assert.NoError(t, err)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)

		data, _ := json.Marshal(got)
		t.Log("got JSON:")
		t.Log(string(data))
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestReleaseUpdateByTag(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Put("/api/v3/projects/xinnjie/testme/releases/v1.0").
		File("testdata/release_create.json").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/release.json")

	client := NewDefault()
	input := &scm.ReleaseInput{
		Title:       "v1.0",
		Description: "Tracking release for version 1.0",
		Tag:         "v1.0",
	}

	got, res, err := client.Releases.UpdateByTag(context.Background(), "xinnjie/testme", "v1.0", input)
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Release)
	raw, _ := ioutil.ReadFile("testdata/release.json.golden")
	err = json.Unmarshal(raw, want)
	assert.NoError(t, err)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestReleaseDeleteByTag(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Delete("/api/v3/projects/xinnjie/testme/releases/v1.0").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders)

	client := NewDefault()
	res, err := client.Releases.DeleteByTag(context.Background(), "xinnjie/testme", "v1.0")
	if err != nil {
		t.Error(err)
		return
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}
