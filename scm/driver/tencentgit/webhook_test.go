// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tencentgit

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/jenkins-x/go-scm/scm"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type mockUserService struct {
	users map[int]*scm.User
}

func (m *mockUserService) FindLoginByID(ctx context.Context, id int) (*scm.User, error) {
	u, ok := m.users[id]
	if ok {
		return u, nil
	}
	return nil, scm.ErrNotFound
}

func TestWebhooks(t *testing.T) {
	tests := []struct {
		event           string
		before          string
		after           string
		obj             interface{}
		mockUserService webhookUserService
	}{
		// tag hooks
		{
			event:  "Tag Push Hook",
			before: "testdata/webhooks/tag_create.json",
			after:  "testdata/webhooks/tag_create.json.golden",
			obj:    new(scm.PushHook),
		},
		// push hooks
		{
			event:  "Push Hook",
			before: "testdata/webhooks/push.json",
			after:  "testdata/webhooks/push.json.golden",
			obj:    new(scm.PushHook),
		},
		{
			event:  "Note Hook",
			before: "testdata/webhooks/issue_comment_create.json",
			after:  "testdata/webhooks/issue_comment_create.json.golden",
			obj:    new(scm.IssueCommentHook),
			mockUserService: &mockUserService{
				users: map[int]*scm.User{
					23: {
						ID:     23,
						Login:  "git_user2",
						Name:   "git_user2",
						Email:  "",
						Avatar: "https://blog.bobo.com.cn/s/blog_6e572cd60101qls0.html",
					},
					11323: {
						ID:     11323,
						Login:  "issue_user",
						Name:   "issue_user",
						Email:  "",
						Avatar: "https://blog.bobo.com.cn/s/blog_6e572cd60101qls0.html",
					},
				},
			},
		},
		// pull request comment hooks
		{
			event:  "Note Hook",
			before: "testdata/webhooks/pull_request_comment_create.json",
			after:  "testdata/webhooks/pull_request_comment_create.json.golden",
			obj:    new(scm.PullRequestCommentHook),
			mockUserService: &mockUserService{
				users: map[int]*scm.User{
					29: {
						ID:     29,
						Login:  "git_user2",
						Name:   "git_user2",
						Email:  "",
						Avatar: "https://blog.bobo.com.cn/s/blog_6e572cd60101qls0.html",
					},
					11322: {
						ID:     11322,
						Login:  "mr_user",
						Name:   "mr_user",
						Email:  "",
						Avatar: "https://blog.bobo.com.cn/s/blog_6e572cd60101qls0.html",
					},
				},
			},
		},
		// pull request hooks
		{
			event:  "Merge Request Hook",
			before: "testdata/webhooks/pull_request_create.json",
			after:  "testdata/webhooks/pull_request_create.json.golden",
			obj:    new(scm.PullRequestHook),
		},
	}

	for _, test := range tests {
		t.Run(test.before, func(t *testing.T) {
			before, err := ioutil.ReadFile(test.before)
			if err != nil {
				t.Error(err)
				return
			}
			after, err := ioutil.ReadFile(test.after)
			if err != nil {
				t.Error(err)
				return
			}

			buf := bytes.NewBuffer(before)
			r, _ := http.NewRequest("GET", "/", buf)
			r.Header.Set("X-Event", test.event)
			r.Header.Set("X-Token", "9edf3260d727e29d906bdb10c8a099a")
			r.Header.Set("X-Request-Id", "ee8d97b4-1479-43f1-9cac-fbbd1b80da55")

			s := new(webhookService)
			s.userService = test.mockUserService
			o, err := s.Parse(r, secretFunc)
			if err != nil && err != scm.ErrSignatureInvalid {
				t.Error(err)
				return
			}

			err = json.Unmarshal(after, &test.obj)
			if err != nil {
				t.Error(err)
				return
			}

			if diff := cmp.Diff(test.obj, o); diff != "" {
				t.Errorf("Error unmarshaling %s", test.before)
				t.Log(diff)

				// debug only. remove once implemented
				json.NewEncoder(os.Stdout).Encode(o)
			}

			switch event := o.(type) {
			case *scm.PushHook:
				if !strings.HasPrefix(event.Ref, "refs/") {
					t.Errorf("Push hook reference must start with refs/")
				}
			case *scm.BranchHook:
				if strings.HasPrefix(event.Ref.Name, "refs/") {
					t.Errorf("Branch hook reference must not start with refs/")
				}
			case *scm.TagHook:
				if strings.HasPrefix(event.Ref.Name, "refs/") {
					t.Errorf("Branch hook reference must not start with refs/")
				}
			}
		})
	}
}

func TestWebhook_SignatureValid(t *testing.T) {
	f, _ := ioutil.ReadFile("testdata/webhooks/branch_delete.json")
	r, _ := http.NewRequest("GET", "/", bytes.NewBuffer(f))
	r.Header.Set("X-Gitlab-Event", "Push Hook")
	r.Header.Set("X-Gitlab-Token", "topsecret")
	r.Header.Set("X-Request-Id", "ee8d97b4-1479-43f1-9cac-fbbd1b80da55")

	s := new(webhookService)
	_, err := s.Parse(r, secretFunc)
	if err != nil {
		t.Error(err)
	}
}

func TestWebhook_SignatureInvalid(t *testing.T) {
	f, _ := ioutil.ReadFile("testdata/webhooks/branch_delete.json")
	r, _ := http.NewRequest("GET", "/", bytes.NewBuffer(f))
	r.Header.Set("X-Gitlab-Event", "Push Hook")
	r.Header.Set("X-Gitlab-Token", "void")
	r.Header.Set("X-Request-Id", "ee8d97b4-1479-43f1-9cac-fbbd1b80da55")

	s := new(webhookService)
	_, err := s.Parse(r, secretFunc)
	if err != scm.ErrSignatureInvalid {
		t.Errorf("Expect invalid signature error, got %v", err)
	}
}

func secretFunc(scm.Webhook) (string, error) {
	return "topsecret", nil
}
