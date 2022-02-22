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

func TestOrganizationFind(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/groups/Twitter").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/group.json")

	client := NewDefault()
	got, _, err := client.Organizations.Find(context.Background(), "Twitter")
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Organization)
	raw, _ := ioutil.ReadFile("testdata/group.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

}

func TestOrganizationList(t *testing.T) {
	defer gock.Off()

	gock.New("https://git.code.tencent.com").
		Get("/api/v3/groups").
		MatchParam("per_page", "30").
		MatchParam("page", "1").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		SetHeaders(mockPageHeaders).
		File("testdata/groups.json")

	client := NewDefault()
	got, res, err := client.Organizations.List(context.Background(), scm.ListOptions{Size: 30, Page: 1})
	if err != nil {
		t.Error(err)
		return
	}

	want := []*scm.Organization{}
	raw, _ := ioutil.ReadFile("testdata/groups.json.golden")
	json.Unmarshal(raw, &want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Page", testPage(res))
}
