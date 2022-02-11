// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package bitbucket

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/jenkins-x/go-scm/scm"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/h2non/gock.v1"
)

func TestOrganizationFind(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.bitbucket.org").
		Get("/2.0/workspaces/atlassian").
		Reply(200).
		Type("application/json").
		File("testdata/workspace.json")

	client, _ := New("https://api.bitbucket.org")
	got, _, err := client.Organizations.Find(context.Background(), "atlassian")
	if err != nil {
		t.Error(err)
	}

	want := new(scm.Organization)
	raw, _ := ioutil.ReadFile("testdata/workspace.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}

func TestOrganizationList(t *testing.T) {
	defer gock.Off()

	gock.New("https://api.bitbucket.org").
		Get("/2.0/workspaces").
		MatchParam("pagelen", "30").
		MatchParam("page", "1").
		Reply(200).
		Type("application/json").
		File("testdata/workspaces.json")

	client, _ := New("https://api.bitbucket.org")
	got, _, err := client.Organizations.List(context.Background(), scm.ListOptions{Size: 30, Page: 1})
	if err != nil {
		t.Error(err)
	}

	want := []*scm.Organization{}
	raw, _ := ioutil.ReadFile("testdata/workspaces.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}
}
