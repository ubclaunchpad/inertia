package webhook

var dockerPushRawJSON = []byte(`
{
	"callback_url": "https://registry.hub.docker.com/u/svendowideit/testhook/hook/2141b5bi5i5b02bec211i4eeih0242eg11000a/",
	"push_data": {
	  "pushed_at": 1.417566161e+09,
	  "pusher": "briannguyen",
	  "tag": "latest"
	},
	"repository": {
	  "comment_count": 0,
	  "date_created": 1.417494799e+09,
	  "description": "",
	  "full_description": "Docker Hub based automated build from a GitHub repo",
	  "is_official": false,
	  "is_private": true,
	  "is_trusted": true,
	  "name": "inertia",
	  "namespace": "inertia",
	  "owner": "ubclaunchpad",
	  "repo_name": "ubclaunchpad/inertia",
	  "repo_url": "https://registry.hub.docker.com/u/svendowideit/testhook/",
	  "star_count": 0,
	  "status": "Active"
	}
}`)
