// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tencentgit

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/jenkins-x/go-scm/scm"
)

const SearchTimeFormat string = "2006-01-02T15:04:05-0700"

func encode(s string) string {
	return strings.Replace(s, "/", "%2F", -1)
}

func encodeListOptions(opts scm.ListOptions) string {
	params := url.Values{}
	if opts.Page != 0 {
		params.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.Size != 0 {
		params.Set("per_page", strconv.Itoa(opts.Size))
	}
	// does not support from/to
	return params.Encode()
}

func encodeMemberListOptions(opts scm.ListOptions) string {
	params := url.Values{}
	params.Set("membership", "true")
	if opts.Page != 0 {
		params.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.Size != 0 {
		params.Set("per_page", strconv.Itoa(opts.Size))
	}
	return params.Encode()
}

func encodeCommitListOptions(opts scm.CommitListOptions) string {
	params := url.Values{}
	if opts.Page != 0 {
		params.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.Size != 0 {
		params.Set("per_page", strconv.Itoa(opts.Size))
	}
	if opts.Ref != "" {
		params.Set("ref_name", opts.Ref)
	}
	return params.Encode()
}

func encodeIssueListOptions(opts scm.IssueListOptions) string {
	params := url.Values{}
	if opts.Page != 0 {
		params.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.Size != 0 {
		params.Set("per_page", strconv.Itoa(opts.Size))
	}
	if opts.Open && opts.Closed {
		params.Set("state", "all")
	} else if opts.Closed {
		params.Set("state", "closed")
	} else if opts.Open {
		params.Set("state", "opened")
	}
	return params.Encode()
}

func encodeMilestoneListOptions(opts scm.MilestoneListOptions) string {
	params := url.Values{}
	if opts.Page != 0 {
		params.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.Size != 0 {
		params.Set("per_page", strconv.Itoa(opts.Size))
	}
	if opts.Closed && !opts.Open {
		params.Set("state", "closed")
	} else if opts.Open && !opts.Closed {
		params.Set("state", "active")
	}
	return params.Encode()
}

func encodePullRequestListOptions(opts scm.PullRequestListOptions) string {
	params := url.Values{}
	if opts.Page != 0 {
		params.Set("page", strconv.Itoa(opts.Page))
	}
	if opts.Size != 0 {
		params.Set("per_page", strconv.Itoa(opts.Size))
	}

	// if opts.Closed/Open not set or all set, retrieve all
	if (!opts.Closed && opts.Open) || (opts.Closed && !opts.Open) {
		if opts.Closed {
			params.Set("state", "closed")
		} else if opts.Open {
			params.Set("state", "opened")
		}
	}

	if len(opts.Labels) > 0 {
		panic("not supported")
	}
	if opts.CreatedAfter != nil {
		params.Set("created_after", opts.CreatedAfter.Format(SearchTimeFormat))
	}
	if opts.CreatedBefore != nil {
		params.Set("created_before", opts.CreatedBefore.Format(SearchTimeFormat))
	}
	if opts.UpdatedAfter != nil {
		params.Set("updated_after", opts.UpdatedAfter.Format(SearchTimeFormat))
	}
	if opts.UpdatedBefore != nil {
		params.Set("updated_before", opts.UpdatedBefore.Format(SearchTimeFormat))
	}
	return params.Encode()
}

func encodePullRequestMergeOptions(opts *scm.PullRequestMergeOptions) *pullRequestMergeRequest {
	prRequest := &pullRequestMergeRequest{}
	if opts == nil {
		return prRequest
	}
	prRequest.MergeCommitType = opts.MergeMethod
	prRequest.MergeCommitMessage = opts.CommitTitle
	return prRequest
}

func tencentgitStateToSCMState(glState string) string {
	switch glState {
	case "opened":
		return "open"
	default:
		return "closed"
	}
}
