// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tencentgit

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jenkins-x/go-scm/scm"
)

type userService struct {
	client *wrapper
}

func (s *userService) CreateToken(context.Context, string, string) (*scm.UserToken, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}

func (s *userService) DeleteToken(context.Context, int64) (*scm.Response, error) {
	return nil, scm.ErrNotSupported
}

func (s *userService) Find(ctx context.Context) (*scm.User, *scm.Response, error) {
	out := new(user)
	res, err := s.client.do(ctx, "GET", "api/v3/user", nil, out)
	return convertUser(out), res, err
}

func (s *userService) FindLogin(ctx context.Context, login string) (*scm.User, *scm.Response, error) {
	var resp *scm.Response
	var err error
	firstRun := false
	opts := scm.ListOptions{
		Page: 1,
	}
	for !firstRun || (resp != nil && opts.Page <= resp.Page.Last) {
		out := []*user{}
		path := fmt.Sprintf("api/v3/users?search=%s&%s", login, encodeListOptions(opts))
		resp, err = s.client.do(ctx, "GET", path, nil, &out)
		if err != nil {
			return nil, nil, err
		}
		firstRun = true
		for _, u := range out {
			if strings.EqualFold(u.Username, login) {
				return convertUser(u), resp, err
			}
		}
		opts.Page++
	}
	return nil, resp, scm.ErrNotFound
}

// FindLoginByID returns the scm.User object for the specified user id
func (s *userService) FindLoginByID(ctx context.Context, id int) (*scm.User, error) {
	path := fmt.Sprintf("api/v3/users/%d", id)
	out := &user{}
	resp, err := s.client.do(ctx, "GET", path, nil, &out)
	if err != nil {
		return nil, err
	}
	if resp.Status == http.StatusOK {
		return convertUser(out), err
	}
	return nil, scm.ErrNotFound
}

func (s *userService) FindEmail(ctx context.Context) (string, *scm.Response, error) {
	user, res, err := s.Find(ctx)
	return user.Email, res, err
}

func (s *userService) ListInvitations(context.Context) ([]*scm.Invitation, *scm.Response, error) {
	return nil, nil, scm.ErrNotSupported
}

func (s *userService) AcceptInvitation(context.Context, int64) (*scm.Response, error) {
	return nil, scm.ErrNotSupported
}

type user struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	WebUrl    string `json:"web_url"`
	Avatar    string `json:"avatar_url"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	IsAdmin   bool   `json:"is_admin"`
}

func convertUser(from *user) *scm.User {
	createAt, _ := time.Parse(timeFormat, from.CreatedAt)
	return &scm.User{
		ID:      from.ID,
		Avatar:  from.Avatar,
		Email:   from.Email,
		Login:   from.Username,
		Name:    from.Name,
		Created: createAt,
		Link:    from.WebUrl,
		IsAdmin: from.IsAdmin,
	}
}

func convertUserList(users []*user) []scm.User {
	dst := []scm.User{}
	for _, src := range users {
		dst = append(dst, *convertUser(src))
	}
	return dst
}
