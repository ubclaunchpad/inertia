package webhook

import (
	"fmt"
	"net/url"
)

// Github Push Event
// see https://developer.github.com/v3/activity/events/types/#pushevent

var githubPushRawJSONStr = `
{
	"ref": "refs/heads/master",
	"before": "0000000000000000000000000000000000000000",
	"after": "f7da6e2506829ef3ee8e3f1a2bfae534a5ab5dfa",
	"created": true,
	"deleted": false,
	"forced": false,
	"base_ref": "refs/heads/master",
	"compare": "https://github.com/brian-nguyen/inertia-deploy-test/compare/master",
	"commits": [],
	"head_commit": {
	  "id": "f7da6e2506829ef3ee8e3f1a2bfae534a5ab5dfa",
	  "tree_id": "093ac15ad7a9b8df5deb553b026f430bc8b5b52d",
	  "distinct": true,
	  "message": "Local parse test",
	  "timestamp": "2018-07-07T13:56:43-07:00",
	  "url": "https://github.com/brian-nguyen/inertia-deploy-test/commit/f7da6e2506829ef3ee8e3f1a2bfae534a5ab5dfa",
	  "author": {
		"name": "brian-nguyen",
		"email": "briannguyen992@gmail.com",
		"username": "brian-nguyen"
	  },
	  "committer": {
		"name": "brian-nguyen",
		"email": "briannguyen992@gmail.com",
		"username": "brian-nguyen"
	  },
	  "added": [],
	  "removed": [],
	  "modified": [
		"README.md"
	  ]
	},
	"repository": {
	  "id": 133707414,
	  "node_id": "MDEwOlJlcG9zaXRvcnkxMzM3MDc0MTQ=",
	  "name": "inertia-deploy-test",
	  "full_name": "brian-nguyen/inertia-deploy-test",
	  "owner": {
		"name": "brian-nguyen",
		"email": "briannguyen992@gmail.com",
		"login": "brian-nguyen",
		"id": 11564004,
		"node_id": "MDQ6VXNlcjExNTY0MDA0",
		"avatar_url": "https://avatars3.githubusercontent.com/u/11564004?v=4",
		"gravatar_id": "",
		"url": "https://api.github.com/users/brian-nguyen",
		"html_url": "https://github.com/brian-nguyen",
		"followers_url": "https://api.github.com/users/brian-nguyen/followers",
		"following_url": "https://api.github.com/users/brian-nguyen/following{/other_user}",
		"gists_url": "https://api.github.com/users/brian-nguyen/gists{/gist_id}",
		"starred_url": "https://api.github.com/users/brian-nguyen/starred{/owner}{/repo}",
		"subscriptions_url": "https://api.github.com/users/brian-nguyen/subscriptions",
		"organizations_url": "https://api.github.com/users/brian-nguyen/orgs",
		"repos_url": "https://api.github.com/users/brian-nguyen/repos",
		"events_url": "https://api.github.com/users/brian-nguyen/events{/privacy}",
		"received_events_url": "https://api.github.com/users/brian-nguyen/received_events",
		"type": "User",
		"site_admin": false
	  },
	  "private": false,
	  "html_url": "https://github.com/brian-nguyen/inertia-deploy-test",
	  "description": ":warning: a repository for testing Inertia deployments",
	  "fork": true,
	  "url": "https://github.com/brian-nguyen/inertia-deploy-test",
	  "forks_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/forks",
	  "keys_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/keys{/key_id}",
	  "collaborators_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/collaborators{/collaborator}",
	  "teams_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/teams",
	  "hooks_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/hooks",
	  "issue_events_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/issues/events{/number}",
	  "events_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/events",
	  "assignees_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/assignees{/user}",
	  "branches_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/branches{/branch}",
	  "tags_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/tags",
	  "blobs_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/git/blobs{/sha}",
	  "git_tags_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/git/tags{/sha}",
	  "git_refs_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/git/refs{/sha}",
	  "trees_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/git/trees{/sha}",
	  "statuses_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/statuses/{sha}",
	  "languages_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/languages",
	  "stargazers_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/stargazers",
	  "contributors_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/contributors",
	  "subscribers_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/subscribers",
	  "subscription_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/subscription",
	  "commits_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/commits{/sha}",
	  "git_commits_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/git/commits{/sha}",
	  "comments_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/comments{/number}",
	  "issue_comment_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/issues/comments{/number}",
	  "contents_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/contents/{+path}",
	  "compare_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/compare/{base}...{head}",
	  "merges_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/merges",
	  "archive_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/{archive_format}{/ref}",
	  "downloads_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/downloads",
	  "issues_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/issues{/number}",
	  "pulls_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/pulls{/number}",
	  "milestones_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/milestones{/number}",
	  "notifications_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/notifications{?since,all,participating}",
	  "labels_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/labels{/name}",
	  "releases_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/releases{/id}",
	  "deployments_url": "https://api.github.com/repos/brian-nguyen/inertia-deploy-test/deployments",
	  "created_at": 1526495113,
	  "updated_at": "2018-07-07T20:56:52Z",
	  "pushed_at": 1532154928,
	  "git_url": "git://github.com/brian-nguyen/inertia-deploy-test.git",
	  "ssh_url": "git@github.com:brian-nguyen/inertia-deploy-test.git",
	  "clone_url": "https://github.com/brian-nguyen/inertia-deploy-test.git",
	  "svn_url": "https://github.com/brian-nguyen/inertia-deploy-test",
	  "homepage": "https://github.com/ubclaunchpad/inertia",
	  "size": 5,
	  "stargazers_count": 0,
	  "watchers_count": 0,
	  "language": "Python",
	  "has_issues": false,
	  "has_projects": true,
	  "has_downloads": true,
	  "has_wiki": false,
	  "has_pages": false,
	  "forks_count": 0,
	  "mirror_url": null,
	  "archived": false,
	  "open_issues_count": 0,
	  "license": null,
	  "forks": 0,
	  "open_issues": 0,
	  "watchers": 0,
	  "default_branch": "master",
	  "stargazers": 0,
	  "master_branch": "master"
	},
	"pusher": {
	  "name": "brian-nguyen",
	  "email": "briannguyen992@gmail.com"
	},
	"sender": {
	  "login": "brian-nguyen",
	  "id": 11564004,
	  "node_id": "MDQ6VXNlcjExNTY0MDA0",
	  "avatar_url": "https://avatars3.githubusercontent.com/u/11564004?v=4",
	  "gravatar_id": "",
	  "url": "https://api.github.com/users/brian-nguyen",
	  "html_url": "https://github.com/brian-nguyen",
	  "followers_url": "https://api.github.com/users/brian-nguyen/followers",
	  "following_url": "https://api.github.com/users/brian-nguyen/following{/other_user}",
	  "gists_url": "https://api.github.com/users/brian-nguyen/gists{/gist_id}",
	  "starred_url": "https://api.github.com/users/brian-nguyen/starred{/owner}{/repo}",
	  "subscriptions_url": "https://api.github.com/users/brian-nguyen/subscriptions",
	  "organizations_url": "https://api.github.com/users/brian-nguyen/orgs",
	  "repos_url": "https://api.github.com/users/brian-nguyen/repos",
	  "events_url": "https://api.github.com/users/brian-nguyen/events{/privacy}",
	  "received_events_url": "https://api.github.com/users/brian-nguyen/received_events",
	  "type": "User",
	  "site_admin": false
	}
}`

var githubPushRawJSON = []byte(githubPushRawJSONStr)

var githubPushFormEncoded = []byte(fmt.Sprintf("payload=%v", url.QueryEscape(githubPushRawJSONStr)))
