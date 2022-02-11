// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tencentgit

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strconv"

	"github.com/mitchellh/copystructure"

	"github.com/jenkins-x/go-scm/scm"
)

type pullService struct {
	client *wrapper
}

func (s *pullService) Find(ctx context.Context, repo string, number int) (*scm.PullRequest, *scm.Response, error) {
	path := fmt.Sprintf("api/v3/projects/%s/merge_request/iid/%d", encode(repo), number)
	out := new(pr)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	if err != nil {
		return nil, res, err
	}
	convRepo, convRes, err := s.convertPullRequest(ctx, repo, out)
	if err != nil {
		return nil, convRes, err
	}
	return convRepo, res, nil
}

func (s *pullService) FindComment(ctx context.Context, repo string, index, id int) (*scm.Comment, *scm.Response, error) {
	path := fmt.Sprintf("api/v3/projects/%s/merge_requests/%d/notes/%d", encode(repo), index, id)
	out := new(issueComment)
	res, err := s.client.do(ctx, "GET", path, nil, out)
	return convertIssueComment(out), res, err
}

func (s *pullService) List(ctx context.Context, repo string, opts scm.PullRequestListOptions) ([]*scm.PullRequest, *scm.Response, error) {
	// label not supported refer to  https://git.woa.com/help/menu/api/merge_requests.html#获取合并请求列表
	if len(opts.Labels) != 0 {
		opts.Labels = nil
	}
	path := fmt.Sprintf("api/v3/projects/%s/merge_requests?%s", encode(repo), encodePullRequestListOptions(opts))
	out := []*pr{}
	res, err := s.client.do(ctx, "GET", path, nil, &out)
	if err != nil {
		return nil, res, err
	}
	convRepos, convRes, err := s.convertPullRequestList(ctx, repo, out)
	if err != nil {
		return nil, convRes, err
	}
	return convRepos, res, nil
}

func (s *pullService) ListChanges(ctx context.Context, repo string, number int, opts scm.ListOptions) ([]*scm.Change, *scm.Response, error) {
	// tencentgit list changes do not support pagination
	path := fmt.Sprintf("api/v3/projects/%s/merge_request/%d/changes", encode(repo), number)
	out := new(changes)
	res, err := s.client.do(ctx, "GET", path, nil, &out)
	return convertChangeList(out.Changes), res, err
}

func (s *pullService) ListComments(ctx context.Context, repo string, index int, opts scm.ListOptions) ([]*scm.Comment, *scm.Response, error) {
	path := fmt.Sprintf("api/v3/projects/%s/merge_requests/%d/notes?%s", encode(repo), index, encodeListOptions(opts))
	out := []*issueComment{}
	res, err := s.client.do(ctx, "GET", path, nil, &out)
	return convertIssueCommentList(out), res, err
}

func (s *pullService) ListLabels(ctx context.Context, repo string, number int, opts scm.ListOptions) ([]*scm.Label, *scm.Response, error) {
	mr, _, err := s.Find(ctx, repo, number)
	if err != nil {
		return nil, nil, err
	}

	return mr.Labels, nil, nil
}

func (s *pullService) ListEvents(ctx context.Context, repo string, index int, opts scm.ListOptions) ([]*scm.ListedIssueEvent, *scm.Response, error) {
	path := fmt.Sprintf("api/v3/projects/%s/merge_requests/%d/resource_label_events?%s", encode(repo), index, encodeListOptions(opts))
	out := []*labelEvent{}
	res, err := s.client.do(ctx, "GET", path, nil, &out)
	return convertLabelEvents(out), res, err
}

func (s *pullService) AddLabel(ctx context.Context, repo string, number int, label string) (*scm.Response, error) {
	return s.setLabels(ctx, repo, number, label, "add_labels")
}

func (s *pullService) DeleteLabel(ctx context.Context, repo string, number int, label string) (*scm.Response, error) {
	return s.setLabels(ctx, repo, number, label, "remove_labels")
}

func (s *pullService) setLabels(ctx context.Context, repo string, number int, labelsStr string, operation string) (*scm.Response, error) {
	in := url.Values{}
	in.Set(operation, labelsStr)
	path := fmt.Sprintf("api/v3/projects/%s/merge_requests/%d?%s", encode(repo), number, in.Encode())

	return s.client.do(ctx, "PUT", path, nil, nil)
}

func (s *pullService) CreateComment(ctx context.Context, repo string, index int, input *scm.CommentInput) (*scm.Comment, *scm.Response, error) {
	in := url.Values{}
	in.Set("body", input.Body)
	path := fmt.Sprintf("api/v3/projects/%s/merge_requests/%d/notes?%s", encode(repo), index, in.Encode())
	out := new(issueComment)
	res, err := s.client.do(ctx, "POST", path, nil, out)
	return convertIssueComment(out), res, err
}

func (s *pullService) DeleteComment(ctx context.Context, repo string, index, id int) (*scm.Response, error) {
	path := fmt.Sprintf("api/v3/projects/%s/merge_requests/%d/notes/%d", encode(repo), index, id)
	res, err := s.client.do(ctx, "DELETE", path, nil, nil)
	return res, err
}

func (s *pullService) EditComment(ctx context.Context, repo string, number int, id int, input *scm.CommentInput) (*scm.Comment, *scm.Response, error) {
	in := &updateNoteOptions{Body: input.Body}
	path := fmt.Sprintf("api/v3/projects/%s/merge_requests/%d/notes/%d", encode(repo), number, id)
	out := new(issueComment)
	res, err := s.client.do(ctx, "PUT", path, in, out)
	return convertIssueComment(out), res, err
}

func (s *pullService) Merge(ctx context.Context, repo string, number int, options *scm.PullRequestMergeOptions) (*scm.Response, error) {
	path := fmt.Sprintf("api/v3/projects/%s/merge_requests/%d/merge", encode(repo), number)
	res, err := s.client.do(ctx, "PUT", path, encodePullRequestMergeOptions(options), nil)
	// tencentgit do not support DeleteSourceBranch, MergeWhenPipelineSucceeds
	// TODO(xinnjie) support DeleteSourceBranch manually
	// TODO(xinnjie) support MergeWhenPipelineSucceeds manually
	return res, err
}

func (s *pullService) Close(ctx context.Context, repo string, number int) (*scm.Response, error) {
	path := fmt.Sprintf("api/v3/projects/%s/merge_requests/%d?state_event=closed", encode(repo), number)
	res, err := s.client.do(ctx, "PUT", path, nil, nil)
	return res, err
}

func (s *pullService) Reopen(ctx context.Context, repo string, number int) (*scm.Response, error) {
	path := fmt.Sprintf("api/v3/projects/%s/merge_requests/%d?state_event=reopen", encode(repo), number)
	res, err := s.client.do(ctx, "PUT", path, nil, nil)
	return res, err
}

func (s *pullService) AssignIssue(ctx context.Context, repo string, number int, logins []string) (*scm.Response, error) {
	pr, _, err := s.Find(ctx, repo, number)
	if err != nil {
		return nil, err
	}

	allAssignees := map[int]struct{}{}
	for _, assignee := range pr.Assignees {
		allAssignees[assignee.ID] = struct{}{}
	}
	for _, l := range logins {
		u, _, err := s.client.Users.FindLogin(ctx, l)
		if err != nil {
			return nil, err
		}
		allAssignees[u.ID] = struct{}{}
	}

	var assigneeIDs []int
	for i := range allAssignees {
		assigneeIDs = append(assigneeIDs, i)
	}

	return s.setAssignees(ctx, repo, number, assigneeIDs)
}

func (s *pullService) setAssignees(ctx context.Context, repo string, number int, ids []int) (*scm.Response, error) {
	if len(ids) == 0 {
		ids = append(ids, 0)
	}
	in := &updateMergeRequestOptions{
		AssigneeIDs: ids,
	}
	path := fmt.Sprintf("api/v3/projects/%s/merge_requests/%d", encode(repo), number)

	return s.client.do(ctx, "PUT", path, in, nil)
}

func (s *pullService) UnassignIssue(ctx context.Context, repo string, number int, logins []string) (*scm.Response, error) {
	pr, _, err := s.Find(ctx, repo, number)
	if err != nil {
		return nil, err
	}
	var assignees []int
	for _, assignee := range pr.Assignees {
		shouldKeep := true
		for _, l := range logins {
			if assignee.Login == l {
				shouldKeep = false
			}
		}
		if shouldKeep {
			assignees = append(assignees, assignee.ID)
		}
	}

	return s.setAssignees(ctx, repo, number, assignees)
}

func (s *pullService) RequestReview(ctx context.Context, repo string, number int, logins []string) (*scm.Response, error) {
	return s.AssignIssue(ctx, repo, number, logins)
}

func (s *pullService) UnrequestReview(ctx context.Context, repo string, number int, logins []string) (*scm.Response, error) {
	return s.UnassignIssue(ctx, repo, number, logins)
}

func (s *pullService) Create(ctx context.Context, repo string, input *scm.PullRequestInput) (*scm.PullRequest, *scm.Response, error) {
	path := fmt.Sprintf("api/v3/projects/%s/merge_requests", encode(repo))
	in := &prInput{
		Title:        input.Title,
		SourceBranch: input.Head,
		TargetBranch: input.Base,
		Description:  input.Body,
	}

	out := new(pr)
	res, err := s.client.do(ctx, "POST", path, in, out)
	if err != nil {
		return nil, res, err
	}
	convRepo, convRes, err := s.convertPullRequest(ctx, repo, out)
	if err != nil {
		return nil, convRes, err
	}
	return convRepo, res, nil
}

func (s *pullService) Update(ctx context.Context, repo string, number int, input *scm.PullRequestInput) (*scm.PullRequest, *scm.Response, error) {
	updateOpts := &updateMergeRequestOptions{}
	if input.Title != "" {
		updateOpts.Title = &input.Title
	}
	if input.Body != "" {
		updateOpts.Description = &input.Body
	}
	if input.Base != "" {
		updateOpts.TargetBranch = &input.Base
	}
	return s.updateMergeRequestField(ctx, repo, number, updateOpts)
}

func (s *pullService) SetMilestone(ctx context.Context, repo string, prID int, number int) (*scm.Response, error) {
	updateOpts := &updateMergeRequestOptions{
		MilestoneID: &number,
	}
	_, res, err := s.updateMergeRequestField(ctx, repo, prID, updateOpts)
	return res, err
}

func (s *pullService) ClearMilestone(ctx context.Context, repo string, prID int) (*scm.Response, error) {
	zeroVal := 0
	updateOpts := &updateMergeRequestOptions{
		MilestoneID: &zeroVal,
	}
	_, res, err := s.updateMergeRequestField(ctx, repo, prID, updateOpts)
	return res, err
}

type updateMergeRequestOptions struct {
	Title              *string `json:"title,omitempty"`
	Description        *string `json:"description,omitempty"`
	TargetBranch       *string `json:"target_branch,omitempty"`
	AssigneeID         *int    `json:"assignee_id,omitempty"`
	AssigneeIDs        []int   `json:"assignee_ids,omitempty"`
	Labels             *string `json:"labels,omitempty"`
	MilestoneID        *int    `json:"milestone_id,omitempty"`
	StateEvent         *string `json:"state_event,omitempty"`
	RemoveSourceBranch *bool   `json:"remove_source_branch,omitempty"`
	Squash             *bool   `json:"squash,omitempty"`
	DiscussionLocked   *bool   `json:"discussion_locked,omitempty"`
	AllowCollaboration *bool   `json:"allow_collaboration,omitempty"`
}

func (s *pullService) updateMergeRequestField(ctx context.Context, repo string, number int, input *updateMergeRequestOptions) (*scm.PullRequest, *scm.Response, error) {
	path := fmt.Sprintf("api/v3/projects/%s/merge_requests/%d", encode(repo), number)

	out := new(pr)
	res, err := s.client.do(ctx, "PUT", path, input, out)
	if err != nil {
		return nil, res, err
	}
	convRepo, convRes, err := s.convertPullRequest(ctx, repo, out)
	if err != nil {
		return nil, convRes, err
	}
	return convRepo, res, nil
}

type pr struct {
	Number          int       `json:"iid"`
	Sha             string    `json:"merge_commit_sha"`
	Title           string    `json:"title"`
	Desc            string    `json:"description"`
	State           string    `json:"state"`
	SourceProjectID int       `json:"source_project_id"`
	TargetProjectID int       `json:"target_project_id"`
	Labels          []*string `json:"labels"`
	WIP             bool      `json:"work_in_progress"`
	Author          user      `json:"author"`
	MergeStatus     string    `json:"merge_status"`
	SourceBranch    string    `json:"source_branch"`
	TargetBranch    string    `json:"target_branch"`
	Created         Time      `json:"created_at"`
	Updated         Time      `json:"updated_at"`
	Closed          Time
	BaseSHA         string  `json:"base_commit"`
	HeadSHA         string  `json:"source_commit"`
	Assignee        *user   `json:"assignee"`
	Assignees       []*user `json:"assignees"`
}

type changes struct {
	Changes []*change `json:"files"`
}

type change struct {
	OldPath string `json:"old_path"`
	NewPath string `json:"new_path"`
	Added   bool   `json:"new_file"`
	Renamed bool   `json:"renamed_file"`
	Deleted bool   `json:"deleted_file"`
	Diff    string `json:"diff"`
}

type prInput struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
}

type pullRequestMergeRequest struct {
	MergeCommitMessage string `json:"merge_commit_message,omitempty"` // merge/squash/rebase
	MergeCommitType    string `json:"merge_type,omitempty"`
}

func (s *pullService) convertPullRequestList(ctx context.Context, repo string, from []*pr) ([]*scm.PullRequest, *scm.Response, error) {
	to := []*scm.PullRequest{}
	for _, v := range from {
		converted, res, err := s.convertPullRequest(ctx, repo, v)
		if err != nil {
			return nil, res, err
		}
		to = append(to, converted)
	}
	return to, nil, nil
}

func (s *pullService) convertPullRequest(ctx context.Context, repo string, from *pr) (*scm.PullRequest, *scm.Response, error) {
	var assignees []scm.User
	if from.Assignee != nil {
		assignees = append(assignees, *convertUser(from.Assignee))
	}
	for _, a := range from.Assignees {
		assignees = append(assignees, *convertUser(a))
	}
	var res *scm.Response
	baseRepo, res, err := s.client.Repositories.Find(ctx, strconv.Itoa(from.TargetProjectID))
	if err != nil {
		return nil, res, err
	}
	var headRepo *scm.Repository
	if from.TargetProjectID == from.SourceProjectID {
		repoCopy, err := copystructure.Copy(baseRepo)
		if err != nil {
			return nil, nil, err
		}
		headRepo = repoCopy.(*scm.Repository)
	} else {
		headRepo, res, err = s.client.Repositories.Find(ctx, strconv.Itoa(from.SourceProjectID))
		if err != nil {
			return nil, res, err
		}
	}
	sourceRepo, err := s.getSourceFork(ctx, from)
	if err != nil {
		return nil, res, err
	}
	link := *s.client.BaseURL
	link.Path = path.Join(link.Path, repo, "merge_requests", strconv.Itoa(from.Number))
	return &scm.PullRequest{
		Number:         from.Number,
		Title:          from.Title,
		Body:           from.Desc,
		State:          tencentgitStateToSCMState(from.State),
		Labels:         convertPullRequestLabels(from.Labels),
		Sha:            from.Sha,
		Ref:            fmt.Sprintf("refs/merge-requests/%d/head", from.Number),
		Source:         from.SourceBranch,
		Target:         from.TargetBranch,
		Link:           link.String(),
		Draft:          from.WIP,
		Closed:         from.State != "opened",
		Merged:         from.State == "merged",
		Mergeable:      scm.ToMergeableState(from.MergeStatus) == scm.MergeableStateMergeable,
		MergeableState: scm.ToMergeableState(from.MergeStatus),
		Author:         *convertUser(&from.Author),
		Assignees:      assignees,
		Head: scm.PullRequestBranch{
			Ref:  from.SourceBranch,
			Sha:  from.HeadSHA,
			Repo: *headRepo,
		},
		Base: scm.PullRequestBranch{
			Ref:  from.TargetBranch,
			Sha:  from.BaseSHA,
			Repo: *baseRepo,
		},
		Created: from.Created.Time,
		Updated: from.Updated.Time,
		Fork:    sourceRepo.PathNamespace,
	}, nil, nil
}

func (s *pullService) getSourceFork(ctx context.Context, from *pr) (repository, error) {
	path := fmt.Sprintf("api/v3/projects/%d", from.SourceProjectID)
	sourceRepo := repository{}
	_, err := s.client.do(ctx, "GET", path, nil, &sourceRepo)
	if err != nil {
		return repository{}, err
	}
	return sourceRepo, nil
}

func convertPullRequestLabels(from []*string) []*scm.Label {
	var labels []*scm.Label
	for _, label := range from {
		l := *label
		labels = append(labels, &scm.Label{
			Name: l,
		})
	}
	return labels
}

func convertChangeList(from []*change) []*scm.Change {
	to := []*scm.Change{}
	for _, v := range from {
		to = append(to, convertChange(v))
	}
	return to
}

func convertChange(from *change) *scm.Change {
	to := &scm.Change{
		Path:         from.NewPath,
		PreviousPath: from.OldPath,
		Added:        from.Added,
		Deleted:      from.Deleted,
		Renamed:      from.Renamed,
		Patch:        from.Diff,
	}
	if to.Path == "" {
		to.Path = from.OldPath
	}
	return to
}