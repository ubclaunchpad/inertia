package webhook

// Gitlab Push Event
// see https://docs.gitlab.com/ee/user/project/integrations/webhooks.html#push-events
var gitlabPushRawJSON = []byte(`
{
	"object_kind": "push",
	"event_name": "push",
	"before": "782fc00feb08df381c7a7d94f52d32cf46fb4065",
	"after": "f7da6e2506829ef3ee8e3f1a2bfae534a5ab5dfa",
	"ref": "refs/heads/master",
	"checkout_sha": "f7da6e2506829ef3ee8e3f1a2bfae534a5ab5dfa",
	"message": null,
	"user_id": 2170981,
	"user_name": "Brian Nguyen",
	"user_username": "brian-nguyen",
	"user_email": "briannguyen992@gmail.com",
	"user_avatar": "https://secure.gravatar.com/avatar/c9054894165564c14023316f304e7297?s=80&d=identicon",
	"project_id": 7392827,
	"project": {
	  "id": 7392827,
	  "name": "inertia-deploy-test",
	  "description": "",
	  "web_url": "https://gitlab.com/brian-nguyen/inertia-deploy-test",
	  "avatar_url": null,
	  "git_ssh_url": "git@gitlab.com:brian-nguyen/inertia-deploy-test.git",
	  "git_http_url": "https://gitlab.com/brian-nguyen/inertia-deploy-test.git",
	  "namespace": "brian-nguyen",
	  "visibility_level": 20,
	  "path_with_namespace": "brian-nguyen/inertia-deploy-test",
	  "default_branch": "master",
	  "ci_config_path": null,
	  "homepage": "https://gitlab.com/brian-nguyen/inertia-deploy-test",
	  "url": "git@gitlab.com:brian-nguyen/inertia-deploy-test.git",
	  "ssh_url": "git@gitlab.com:brian-nguyen/inertia-deploy-test.git",
	  "http_url": "https://gitlab.com/brian-nguyen/inertia-deploy-test.git"
	},
	"commits": [
	  {
		"id": "f7da6e2506829ef3ee8e3f1a2bfae534a5ab5dfa",
		"message": "Local parse test\n",
		"timestamp": "2018-07-07T20:56:43Z",
		"url": "https://gitlab.com/brian-nguyen/inertia-deploy-test/commit/f7da6e2506829ef3ee8e3f1a2bfae534a5ab5dfa",
		"author": {
		  "name": "brian-nguyen",
		  "email": "briannguyen992@gmail.com"
		},
		"added": [],
		"modified": [
		  "README.md"
		],
		"removed": []
	  },
	  {
		"id": "a29ac9c00794f1325182d6afdbafad2558b64c85",
		"message": "Github remote test\n",
		"timestamp": "2018-07-07T06:16:38Z",
		"url": "https://gitlab.com/brian-nguyen/inertia-deploy-test/commit/a29ac9c00794f1325182d6afdbafad2558b64c85",
		"author": {
		  "name": "brian-nguyen",
		  "email": "briannguyen992@gmail.com"
		},
		"added": [],
		"modified": [
		  "README.md"
		],
		"removed": []
	  },
	  {
		"id": "782fc00feb08df381c7a7d94f52d32cf46fb4065",
		"message": "Enforce JSON for Github\n",
		"timestamp": "2018-07-07T06:15:45Z",
		"url": "https://gitlab.com/brian-nguyen/inertia-deploy-test/commit/782fc00feb08df381c7a7d94f52d32cf46fb4065",
		"author": {
		  "name": "brian-nguyen",
		  "email": "briannguyen992@gmail.com"
		},
		"added": [],
		"modified": [
		  "README.md"
		],
		"removed": []
	  }
	],
	"total_commits_count": 3,
	"repository": {
	  "name": "inertia-deploy-test",
	  "url": "git@gitlab.com:brian-nguyen/inertia-deploy-test.git",
	  "description": "",
	  "homepage": "https://gitlab.com/brian-nguyen/inertia-deploy-test",
	  "git_http_url": "https://gitlab.com/brian-nguyen/inertia-deploy-test.git",
	  "git_ssh_url": "git@gitlab.com:brian-nguyen/inertia-deploy-test.git",
	  "visibility_level": 20
	}
}`)
