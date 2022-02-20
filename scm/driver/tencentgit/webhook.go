// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tencentgit

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jenkins-x/go-scm/scm"
	"github.com/sirupsen/logrus"
)

type webhookService struct {
	client *wrapper
	//need the user service as well
	userService webhookUserService
}

// an interface to provide a find login by id in tencentgit
type webhookUserService interface {
	FindLoginByID(ctx context.Context, id int) (*scm.User, error)
}

func (s *webhookService) Parse(req *http.Request, fn scm.SecretFunc) (scm.Webhook, error) {
	data, err := ioutil.ReadAll(
		io.LimitReader(req.Body, 10000000),
	)
	if err != nil {
		return nil, err
	}

	var hook scm.Webhook
	event := req.Header.Get("X-Event")

	log := logrus.WithFields(map[string]interface{}{
		"URL":     req.URL,
		"Headers": req.Header,
		"Body":    string(data),
	})
	log.Infof("received webhook")
	switch event {
	case "Push Hook", "Tag Push Hook":
		hook, err = parsePushHook(data)
	case "Issue Hook":
		return nil, scm.UnknownWebhook{Event: event}
	case "Merge Request Hook":
		hook, err = parsePullRequestHook(data)
	case "Note Hook":
		hook, err = parseCommentHook(s, data)
	default:
		return nil, scm.UnknownWebhook{Event: event}
	}
	if err != nil {
		return nil, err
	}

	// get the tencentgit shared token to verify the payload
	// authenticity. If no key is provided, no validation
	// is performed.
	token, err := fn(hook)
	if err != nil {
		return hook, err
	} else if token == "" {
		return hook, nil
	}

	if subtle.ConstantTimeCompare([]byte(req.Header.Get("X-Token")), []byte(token)) == 0 {
		return hook, scm.ErrSignatureInvalid
	}
	return hook, nil
}

func parsePushHook(data []byte) (scm.Webhook, error) {
	src := new(pushHook)
	err := json.Unmarshal(data, src)
	if err != nil {
		return nil, err
	}
	switch {
	case src.ObjectKind == "push" && src.Before == "0000000000000000000000000000000000000000":
		// TODO we previously considered returning a
		// branch creation hook, however, the push hook
		// returns more metadata (commit details).
		return convertPushHook(src), nil
	case src.ObjectKind == "push" && src.After == "0000000000000000000000000000000000000000":
		return converBranchHook(src), nil
	case src.ObjectKind == "tag_push" && src.Before == "0000000000000000000000000000000000000000":
		// TODO we previously considered returning a
		// tag creation hook, however, the push hook
		// returns more metadata (commit details).
		return convertPushHook(src), nil
	case src.ObjectKind == "tag_push" && src.After == "0000000000000000000000000000000000000000":
		return convertTagHook(src), nil
	default:
		return convertPushHook(src), nil
	}
}

func parsePullRequestHook(data []byte) (scm.Webhook, error) {
	src := new(pullRequestHook)
	err := json.Unmarshal(data, src)
	if err != nil {
		return nil, err
	}
	if src.ObjectAttributes.Action == "" {
		src.ObjectAttributes.Action = "open"
	}
	event := src.ObjectAttributes.Action
	switch event {
	case "open", "close", "reopen", "merge", "update":
		// no-op
	default:
		return nil, scm.UnknownWebhook{Event: event}
	}
	return convertPullRequestHook(src), nil
}

func parseCommentHook(s *webhookService, data []byte) (scm.Webhook, error) {
	src := new(commentHook)
	err := json.Unmarshal(data, src)
	if err != nil {
		return nil, err
	}
	if src.ObjectAttributes.Action == "" {
		src.ObjectAttributes.Action = "open"
	}
	kind := src.ObjectAttributes.NoteableType
	switch kind {
	case "MergeRequest":
		return convertMergeRequestCommentHook(s, src)
	case "Issue":
		return convertIssueCommentHook(s, src)
	default:
		return nil, scm.UnknownWebhook{Event: kind}
	}
}

func (repo *repo) ToScmRepository(projectID int) *scm.Repository {
	httpUrl := repo.HttpUrl
	sshUrl := repo.SshUrl
	homePage := repo.HomePage

	if repo.GitHttpUrl != "" {
		httpUrl = repo.GitHttpUrl
	}

	if httpUrl == "" {
		httpUrl = repo.Url
	}

	if sshUrl == "" {
		sshUrl = repo.GitSshUrl
	}

	if homePage == "" {
		homePage = repo.WebUrl
	}

	namespace := repo.Namespace
	if namespace == "" {
		nameWithNameSpaceRegex := regexp.MustCompile("[\\w|-]+\\/[\\w|-]+$")
		nameWithNameSpace := nameWithNameSpaceRegex.FindString(homePage)
		if nameWithNameSpace != "" {
			namespace = strings.Split(nameWithNameSpace, "/")[0]
		}
	}

	return &scm.Repository{
		ID:        strconv.Itoa(projectID),
		Namespace: namespace,
		Name:      repo.Name,
		FullName:  fmt.Sprintf("%v/%v", namespace, repo.Name),
		Perm:      nil,   // TODO(xinnjie) need query repo info
		Branch:    "",    // TODO
		Private:   false, // TODO
		Archived:  false, // TODO
		Clone:     httpUrl,
		CloneSSH:  sshUrl,
		Link:      homePage,
		Created:   time.Time{}, // TODO
		Updated:   time.Time{}, // TODO
	}
}

func convertPushHook(src *pushHook) *scm.PushHook {
	repo := src.Repository.ToScmRepository(src.ProjectID)
	dst := &scm.PushHook{
		Ref:   scm.ExpandRef(src.Ref, "refs/heads/"),
		Repo:  *repo,
		After: src.After,
		Commit: scm.Commit{
			Sha:     "", // NOTE this is set below
			Message: "", // NOTE this is set below
			Author: scm.Signature{
				Login: src.UserName,
				Name:  src.UserName,
				Email: src.UserEmail,
			},
			Committer: scm.Signature{
				Login: src.UserName,
				Name:  src.UserName,
				Email: src.UserEmail,
			},
			Link: "", // NOTE this is set below
		},
		Sender: scm.User{
			Login: src.UserName,
			Name:  src.UserName,
			Email: src.UserEmail,
		},
	}
	if len(src.Commits) > 0 {
		// get the last commit (most recent)
		dst.Commit.Message = src.Commits[len(src.Commits)-1].Message
		dst.Commit.Link = src.Commits[len(src.Commits)-1].URL
		dst.Commit.Sha = src.Commits[len(src.Commits)-1].ID
		author := src.Commits[len(src.Commits)-1].Author
		dst.Commit.Author = scm.Signature{
			Name:  author.Name,
			Email: author.Email,
			Login: author.Name,
		}
	}
	// TODO(xinnjie) set dst.Commits
	return dst
}

func converBranchHook(src *pushHook) *scm.BranchHook {
	action := scm.ActionCreate
	commit := src.After
	if src.After == "0000000000000000000000000000000000000000" {
		action = scm.ActionDelete
		commit = src.Before
	}
	repo := *src.Repository.ToScmRepository(src.ProjectID)
	return &scm.BranchHook{
		Action: action,
		Ref: scm.Reference{
			Name: scm.TrimRef(src.Ref),
			Sha:  commit,
		},
		Repo: repo,
		Sender: scm.User{
			Login: src.UserName,
			Name:  src.UserName,
			Email: src.UserEmail,
		},
	}
}

func convertTagHook(src *pushHook) *scm.TagHook {
	action := scm.ActionCreate
	commit := src.After
	if src.After == "0000000000000000000000000000000000000000" {
		action = scm.ActionDelete
		commit = src.Before
	}
	repo := *src.Repository.ToScmRepository(src.ProjectID)
	return &scm.TagHook{
		Action: action,
		Ref: scm.Reference{
			Name: scm.TrimRef(src.Ref),
			Sha:  commit,
		},
		Repo: repo,
		Sender: scm.User{
			Login: src.UserName,
			Name:  src.UserName,
			Email: src.UserEmail,
		},
	}
}

func convertPullRequestHook(src *pullRequestHook) *scm.PullRequestHook {
	action := scm.ActionUpdate
	switch src.ObjectAttributes.Action {
	case "open":
		action = scm.ActionOpen
	case "close":
		action = scm.ActionClose
	case "reopen":
		action = scm.ActionReopen
	case "merge":
		action = scm.ActionMerge
	case "update":
		action = scm.ActionUpdate
	}
	fork := scm.Join(
		src.ObjectAttributes.Source.Namespace,
		src.ObjectAttributes.Source.Name,
	)
	ref := fmt.Sprintf("refs/merge-requests/%d/head", src.ObjectAttributes.Iid)
	sha := src.ObjectAttributes.LastCommit.ID
	srcRepo := *src.ObjectAttributes.Source.ToScmRepository(src.ObjectAttributes.SourceProjectID)
	targetRepo := *src.ObjectAttributes.Target.ToScmRepository(src.ObjectAttributes.TargetProjectID)
	createAt, _ := time.Parse(timeFormat, src.ObjectAttributes.CreatedAt)
	updateAt, _ := time.Parse(timeFormat, src.ObjectAttributes.UpdatedAt)

	pr := scm.PullRequest{
		Number: src.ObjectAttributes.Iid,
		Title:  src.ObjectAttributes.Title,
		Body:   src.ObjectAttributes.Description,
		State:  tencentgitStateToSCMState(src.ObjectAttributes.State),
		Sha:    sha,
		Ref:    ref,
		Base: scm.PullRequestBranch{
			Ref:  src.ObjectAttributes.TargetBranch,
			Repo: targetRepo,
		},
		Head: scm.PullRequestBranch{
			Sha:  sha,
			Ref:  src.ObjectAttributes.SourceBranch,
			Repo: srcRepo,
		},
		Source:  src.ObjectAttributes.SourceBranch,
		Target:  src.ObjectAttributes.TargetBranch,
		Fork:    fork,
		Link:    src.ObjectAttributes.URL,
		Closed:  src.ObjectAttributes.State != "opened",
		Merged:  src.ObjectAttributes.State == "merged",
		Created: createAt,
		Updated: updateAt,
		Author: scm.User{
			Login:  src.User.Username,
			Name:   src.User.Name,
			Email:  "", // TODO how do we get the pull request author email?
			Avatar: src.User.AvatarURL,
		},
	}
	return &scm.PullRequestHook{
		Action:      action,
		PullRequest: pr,
		Sender: scm.User{
			Login:  src.User.Username,
			Name:   src.User.Name,
			Email:  "", // TODO how do we get the pull request author email?
			Avatar: src.User.AvatarURL,
		},
		Repo: targetRepo,
	}
}

func convertIssueCommentHook(s *webhookService, src *commentHook) (*scm.IssueCommentHook, error) {
	commentAuthor, err := s.userService.FindLoginByID(context.TODO(), src.ObjectAttributes.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("unable to find comment author %w", err)
	}

	issueAuthor, err := s.userService.FindLoginByID(context.TODO(), src.Issue.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("unable to find issue author %w", err)
	}

	repo :=
		*src.Repository.ToScmRepository(src.ProjectID)
	createdAt, _ := time.Parse(timeFormat, src.ObjectAttributes.CreatedAt)
	updatedAt, _ := time.Parse(timeFormat, src.ObjectAttributes.UpdatedAt)

	issue := scm.Issue{
		Number:      src.Issue.Iid,
		Title:       src.Issue.Title,
		Body:        src.Issue.Description,
		Author:      *issueAuthor,
		Created:     createdAt,
		Updated:     updatedAt,
		Closed:      src.Issue.State != "opened",
		PullRequest: false,
	}

	hook := &scm.IssueCommentHook{
		Action: scm.ActionCreate,
		Repo:   repo,
		Issue:  issue,
		Comment: scm.Comment{
			ID:      src.ObjectAttributes.ID,
			Body:    src.ObjectAttributes.Note,
			Author:  *commentAuthor,
			Created: createdAt,
			Updated: updatedAt,
		},
		Sender: *commentAuthor,
	}
	return hook, nil
}

func convertMergeRequestCommentHook(s *webhookService, src *commentHook) (*scm.PullRequestCommentHook, error) {

	// There are two users needed here: the comment author and the MergeRequest author.
	// Since we only have the user name, we need to use the user service to fetch these.
	commentAuthor, err := s.userService.FindLoginByID(context.TODO(), src.ObjectAttributes.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("unable to find comment author %w", err)
	}

	mrAuthor, err := s.userService.FindLoginByID(context.TODO(), src.MergeRequest.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("unable to find mr author %w", err)
	}

	fork := scm.Join(
		src.MergeRequest.Source.Namespace,
		src.MergeRequest.Source.Name,
	)

	repo := *src.Repository.ToScmRepository(src.ProjectID)

	ref := fmt.Sprintf("refs/merge-requests/%d/head", src.MergeRequest.Iid)
	sha := src.MergeRequest.LastCommit.ID

	// Mon Jan 2 15:04:05 -0700 MST 2006
	prCreatedAt, _ := time.Parse("2006-01-02 15:04:05 MST", src.MergeRequest.CreatedAt)
	prUpdatedAt, _ := time.Parse("2006-01-02 15:04:05 MST", src.MergeRequest.UpdatedAt)
	pr := scm.PullRequest{
		Number: src.MergeRequest.Iid,
		Title:  src.MergeRequest.Title,
		Body:   src.MergeRequest.Description,
		State:  tencentgitStateToSCMState(src.MergeRequest.State),
		Sha:    sha,
		Ref:    ref,
		Base: scm.PullRequestBranch{
			Ref: src.MergeRequest.TargetBranch,
		},
		Head: scm.PullRequestBranch{
			Ref: src.MergeRequest.SourceBranch,
			Sha: sha,
		},
		Source:  src.MergeRequest.SourceBranch,
		Target:  src.MergeRequest.TargetBranch,
		Fork:    fork,
		Link:    src.MergeRequest.URL,
		Closed:  src.MergeRequest.State != "opened",
		Merged:  src.MergeRequest.State == "merged",
		Created: prCreatedAt,
		Updated: prUpdatedAt, // 2017-12-10 17:01:11 UTC
		Author:  *mrAuthor,
	}

	pr.Base.Repo = *src.MergeRequest.Target.ToScmRepository(src.MergeRequest.TargetProjectID)
	pr.Head.Repo = *src.MergeRequest.Source.ToScmRepository(src.MergeRequest.SourceProjectID)

	createdAt, _ := time.Parse("2006-01-02 15:04:05 MST", src.ObjectAttributes.CreatedAt)
	updatedAt, _ := time.Parse("2006-01-02 15:04:05 MST", src.ObjectAttributes.UpdatedAt)

	hook := &scm.PullRequestCommentHook{
		Action:      scm.ActionCreate,
		Repo:        repo,
		PullRequest: pr,
		Comment: scm.Comment{
			ID:      src.ObjectAttributes.ID,
			Body:    src.ObjectAttributes.Note,
			Author:  *commentAuthor,
			Created: createdAt,
			Updated: updatedAt,
		},
		Sender: *commentAuthor,
	}
	return hook, nil
}

func convertAction(src string) (action scm.Action) {
	switch src {
	case "create":
		return scm.ActionCreate
	case "open":
		return scm.ActionOpen
	default:
		return
	}
}

type (
	//doc at https://git.woa.com/help/menu/manual/webhooks.html#推送事件
	pushHook struct {
		ObjectKind    string `json:"object_kind"`
		OperationKind string `json:"operation_kind"`
		ActionKind    string `json:"action_kind"`
		Before        string `json:"before"`
		After         string `json:"after"`
		Ref           string `json:"ref"`
		UserID        int    `json:"user_id"`
		UserName      string `json:"user_name"`
		UserEmail     string `json:"user_email"`
		ProjectID     int    `json:"project_id"`
		Commits       []struct {
			ID        string `json:"id"`
			Message   string `json:"message"`
			Timestamp string `json:"timestamp"`
			URL       string `json:"url"`
			Author    struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			} `json:"author"`
			Added    []string      `json:"added"`
			Modified []interface{} `json:"modified"`
			Removed  []interface{} `json:"removed"`
		} `json:"commits"`
		TotalCommitsCount int  `json:"total_commits_count"`
		Repository        repo `json:"repository"`
	}
	repo struct {
		Name            string `json:"name"`
		Description     string `json:"description"`
		Namespace       string `json:"namespace"` // maybe empty
		VisibilityLevel int    `json:"visibility_level"`
		WebUrl          string `json:"web_url"`
		SshUrl          string `json:"ssh_url"`
		HttpUrl         string `json:"http_url"`
		HomePage        string `json:"homepage"`     // same as WebUrl, for compatibility
		GitHttpUrl      string `json:"git_http_url"` // same as HttpUrl, for compatibility
		GitSshUrl       string `json:"git_ssh_url"`  // same as SshUrl, for compatibility
		Url             string `json:"url"`          // same as HttpUrl, for compatibility
	}
	commentHook struct {
		ObjectKind string `json:"object_kind"`
		User       struct {
			Name      string `json:"name"`
			Username  string `json:"username"`
			AvatarURL string `json:"avatar_url"`
		} `json:"user"`
		ProjectID        int `json:"project_id"`
		ObjectAttributes struct {
			ID           int         `json:"id"`
			Note         string      `json:"note"`
			NoteableType string      `json:"noteable_type"`
			AuthorID     int         `json:"author_id"`
			CreatedAt    string      `json:"created_at"`
			UpdatedAt    string      `json:"updated_at"`
			ProjectID    int         `json:"project_id"`
			Attachment   interface{} `json:"attachment"`
			LineCode     string      `json:"line_code"`
			CommitID     string      `json:"commit_id"`
			NoteableID   int         `json:"noteable_id"`
			StDiff       interface{} `json:"st_diff"`
			System       bool        `json:"system"`
			UpdatedByID  interface{} `json:"updated_by_id"`
			Type         string      `json:"type"`
			Position     struct {
				BaseSha      string      `json:"base_sha"`
				StartSha     string      `json:"start_sha"`
				HeadSha      string      `json:"head_sha"`
				OldPath      string      `json:"old_path"`
				NewPath      string      `json:"new_path"`
				PositionType string      `json:"position_type"`
				OldLine      interface{} `json:"old_line"`
				NewLine      int         `json:"new_line"`
			} `json:"position"`
			OriginalPosition struct {
				BaseSha      string      `json:"base_sha"`
				StartSha     string      `json:"start_sha"`
				HeadSha      string      `json:"head_sha"`
				OldPath      string      `json:"old_path"`
				NewPath      string      `json:"new_path"`
				PositionType string      `json:"position_type"`
				OldLine      interface{} `json:"old_line"`
				NewLine      int         `json:"new_line"`
			} `json:"original_position"`
			ResolvedAt     interface{} `json:"resolved_at"`
			ResolvedByID   interface{} `json:"resolved_by_id"`
			DiscussionID   string      `json:"discussion_id"`
			ChangePosition struct {
				BaseSha      interface{} `json:"base_sha"`
				StartSha     interface{} `json:"start_sha"`
				HeadSha      interface{} `json:"head_sha"`
				OldPath      interface{} `json:"old_path"`
				NewPath      interface{} `json:"new_path"`
				PositionType string      `json:"position_type"`
				OldLine      interface{} `json:"old_line"`
				NewLine      interface{} `json:"new_line"`
			} `json:"change_position"`
			ResolvedByPush interface{} `json:"resolved_by_push"`
			URL            string      `json:"url"`
			Action         string      `json:"action"`
		} `json:"object_attributes"`
		Repository   repo `json:"repository"`
		MergeRequest struct {
			AssigneeID     interface{} `json:"assignee_id"`
			AuthorID       int         `json:"author_id"`
			CreatedAt      string      `json:"created_at"`
			DeletedAt      interface{} `json:"deleted_at"`
			Description    string      `json:"description"`
			HeadPipelineID interface{} `json:"head_pipeline_id"`
			ID             int         `json:"id"`
			Iid            int         `json:"iid"`
			LastEditedAt   interface{} `json:"last_edited_at"`
			LastEditedByID interface{} `json:"last_edited_by_id"`
			MergeCommitSha string      `json:"merge_commit_sha"`
			MergeError     interface{} `json:"merge_error"`
			MergeParams    struct {
				ForceRemoveSourceBranch string `json:"force_remove_source_branch"`
			} `json:"merge_params"`
			MergeStatus               string      `json:"merge_status"`
			MergeUserID               interface{} `json:"merge_user_id"`
			MergeWhenPipelineSucceeds bool        `json:"merge_when_pipeline_succeeds"`
			MilestoneID               interface{} `json:"milestone_id"`
			SourceBranch              string      `json:"source_branch"`
			SourceProjectID           int         `json:"source_project_id"`
			State                     string      `json:"state"`
			TargetBranch              string      `json:"target_branch"`
			TargetProjectID           int         `json:"target_project_id"`
			TimeEstimate              int         `json:"time_estimate"`
			Title                     string      `json:"title"`
			UpdatedAt                 string      `json:"updated_at"`
			UpdatedByID               interface{} `json:"updated_by_id"`
			URL                       string      `json:"url"`
			Source                    *repo       `json:"source"`
			Target                    *repo       `json:"target"`
			LastCommit                struct {
				ID        string `json:"id"`
				Message   string `json:"message"`
				Timestamp string `json:"timestamp"`
				URL       string `json:"url"`
				Author    struct {
					Name  string `json:"name"`
					Email string `json:"email"`
				} `json:"author"`
			} `json:"last_commit"`
			WorkInProgress      bool        `json:"work_in_progress"`
			TotalTimeSpent      int         `json:"total_time_spent"`
			HumanTotalTimeSpent interface{} `json:"human_total_time_spent"`
			HumanTimeEstimate   interface{} `json:"human_time_estimate"`
		} `json:"merge_request"`
		Issue struct {
			ID          int      `json:"id"`
			Title       string   `json:"title"`
			AssigneeIDs []string `json:"assignee_ids"`
			AssigneeID  string   `json:"assignee_id"`
			AuthorID    int      `json:"author_id"`
			ProjectID   int      `json:"project_id"`
			CreatedAt   string   `json:"created_at"`
			UpdatedAt   string   `json:"updated_at"`
			Position    int      `json:"position"`
			BranchName  string   `json:"branch_name"`
			Description string   `json:"description"`
			MilestoneID string   `json:"milestone_id"`
			State       string   `json:"state"`
			Iid         int      `json:"iid"`
			Labels      []struct {
				ID          int    `json:"id"`
				Title       string `json:"title"`
				Color       string `json:"color"`
				ProjectID   int    `json:"project_id"`
				CreatedAt   string `json:"created_at"`
				UpdatedAt   string `json:"updated_at"`
				Template    bool   `json:"template"`
				Description string `json:"description"`
				LabelType   string `json:"type"`
				GroupID     int    `json:"group_id"`
			} `json:"labels"`
		} `json:"issue"`
	}

	tagHook struct {
		ObjectKind    string `json:"object_kind"` // "tag_push"
		OperationKind string `json:"operation_kind"`
		Before        string `json:"before"`
		After         string `json:"after"`
		Ref           string `json:"ref"`
		UserID        int    `json:"user_id"`
		UserName      string `json:"user_name"`
		UserEmail     string `json:"user_email"`
		ProjectID     int    `json:"project_id"`
		Commits       []struct {
			ID        string `json:"id"`
			Message   string `json:"message"`
			Timestamp string `json:"timestamp"`
			URL       string `json:"url"`
			Author    struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			} `json:"author"`
			Added    []string      `json:"added"`
			Modified []interface{} `json:"modified"`
			Removed  []interface{} `json:"removed"`
		} `json:"commits"`
		TotalCommitsCount int    `json:"total_commits_count"`
		CreateFrom        string `json:"create_from"`
		Repository        struct {
			Name            string `json:"name"`
			URL             string `json:"url"`
			Description     string `json:"description"`
			Homepage        string `json:"homepage"`
			GitHTTPURL      string `json:"git_http_url"`
			GitSSHURL       string `json:"git_ssh_url"`
			VisibilityLevel int    `json:"visibility_level"`
		} `json:"repository"`
	}

	issueHook struct {
		ObjectKind string `json:"object_kind"`
		User       struct {
			Name      string `json:"name"`
			Username  string `json:"username"`
			AvatarURL string `json:"avatar_url"`
			Email     string `json:"email"`
		} `json:"user"`
		Project          repo `json:"repo"`
		ObjectAttributes struct {
			AssigneeID          interface{}   `json:"assignee_id"`
			AuthorID            int           `json:"author_id"`
			BranchName          interface{}   `json:"branch_name"`
			ClosedAt            interface{}   `json:"closed_at"`
			Confidential        bool          `json:"confidential"`
			CreatedAt           string        `json:"created_at"`
			DeletedAt           interface{}   `json:"deleted_at"`
			Description         string        `json:"description"`
			DueDate             interface{}   `json:"due_date"`
			ID                  int           `json:"id"`
			Iid                 int           `json:"iid"`
			LastEditedAt        string        `json:"last_edited_at"`
			LastEditedByID      int           `json:"last_edited_by_id"`
			MilestoneID         interface{}   `json:"milestone_id"`
			MovedToID           interface{}   `json:"moved_to_id"`
			ProjectID           int           `json:"project_id"`
			RelativePosition    int           `json:"relative_position"`
			State               string        `json:"state"`
			TimeEstimate        int           `json:"time_estimate"`
			Title               string        `json:"title"`
			UpdatedAt           string        `json:"updated_at"`
			UpdatedByID         int           `json:"updated_by_id"`
			URL                 string        `json:"url"`
			TotalTimeSpent      int           `json:"total_time_spent"`
			HumanTotalTimeSpent interface{}   `json:"human_total_time_spent"`
			HumanTimeEstimate   interface{}   `json:"human_time_estimate"`
			AssigneeIds         []interface{} `json:"assignee_ids"`
			Action              string        `json:"action"`
		} `json:"object_attributes"`
		Labels []struct {
			ID          int         `json:"id"`
			Title       string      `json:"title"`
			Color       string      `json:"color"`
			ProjectID   string      `json:"project_id"`
			CreatedAt   string      `json:"created_at"`
			UpdatedAt   string      `json:"updated_at"`
			Template    bool        `json:"template"`
			Description string      `json:"description"`
			Type        string      `json:"type"`
			GroupID     interface{} `json:"group_id"`
		} `json:"labels"`
		Changes struct {
			Labels struct {
				Previous []interface{} `json:"previous"`
				Current  []struct {
					ID          int         `json:"id"`
					Title       string      `json:"title"`
					Color       string      `json:"color"`
					ProjectID   int         `json:"project_id"`
					CreatedAt   string      `json:"created_at"`
					UpdatedAt   string      `json:"updated_at"`
					Template    bool        `json:"template"`
					Description string      `json:"description"`
					Type        string      `json:"type"`
					GroupID     interface{} `json:"group_id"`
				} `json:"current"`
			} `json:"labels"`
		} `json:"changes"`
		Repository struct {
			Name        string `json:"name"`
			URL         string `json:"url"`
			Description string `json:"description"`
			Homepage    string `json:"homepage"`
		} `json:"repository"`
	}

	pullRequestHook struct {
		ObjectKind string `json:"object_kind"` // "merge_request"
		User       struct {
			Name      string `json:"name"`
			Username  string `json:"username"`
			AvatarURL string `json:"avatar_url"`
		} `json:"user"`
		ObjectAttributes struct {
			ID              int         `json:"id"`
			Iid             int         `json:"iid"`
			TargetBranch    string      `json:"target_branch"`
			SourceBranch    string      `json:"source_branch"`
			SourceProjectID int         `json:"source_project_id"`
			AssigneeID      int         `json:"assignee_id"`
			AuthorID        int         `json:"author_id"`
			Title           string      `json:"title"`
			CreatedAt       string      `json:"created_at"` // "2018-03-13T09:51:06+0000"
			UpdatedAt       string      `json:"updated_at"`
			Description     string      `json:"description"`
			MilestoneID     interface{} `json:"milestone_id"`
			State           string      `json:"state"`
			TargetProjectID int         `json:"target_project_id"`
			URL             string      `json:"url"`
			Source          *repo       `json:"source"`
			Target          *repo       `json:"target"`
			LastCommit      struct {
				ID        string `json:"id"`
				Message   string `json:"message"`
				Timestamp string `json:"timestamp"`
				URL       string `json:"url"`
				Author    struct {
					Name  string `json:"name"`
					Email string `json:"email"`
				} `json:"author"`
			} `json:"last_commit"`
			Action string `json:"action"`
		} `json:"object_attributes"`
	}

	releaseHook struct {
		ID          int    `json:"id"`
		CreatedAt   string `json:"created_at"`
		Description string `json:"description"`
		Name        string `json:"name"`
		ReleasedAt  string `json:"released_at"`
		Tag         string `json:"tag"`
		ObjectKind  string `json:"object_kind"`
		Project     repo   `json:"repo"`
		URL         string `json:"url"`
		Action      string `json:"action"`
		Assets      struct {
			Count int `json:"count"`
			Links []struct {
				ID       int    `json:"id"`
				External bool   `json:"external"`
				LinkType string `json:"link_type"`
				Name     string `json:"name"`
				URL      string `json:"url"`
			} `json:"links"`
			Sources []struct {
				Format string `json:"format"`
				URL    string `json:"url"`
			} `json:"sources"`
		} `json:"assets"`
		Commit struct {
			ID        string `json:"id"`
			Message   string `json:"message"`
			Title     string `json:"title"`
			Timestamp string `json:"timestamp"`
			URL       string `json:"url"`
			Author    struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			} `json:"author"`
		} `json:"commit"`
	}
)
