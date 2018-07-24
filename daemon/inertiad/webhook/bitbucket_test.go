package webhook

// Bitbucket Push Event
// see https://confluence.atlassian.com/bitbucket/event-payloads-740262817.html
var bitbucketPushRawJSON = []byte(`
{
	"push": {
	  "changes": [
		{
		  "forced": false,
		  "old": {
			"type": "branch",
			"name": "master",
			"links": {
			  "commits": {
				"href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/commits/master"
			  },
			  "self": {
				"href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/refs/branches/master"
			  },
			  "html": {
				"href": "https://bitbucket.org/brian-nguyen/inertia-deploy-test/branch/master"
			  }
			},
			"target": {
			  "hash": "88d0d4becc199c8c2005faebca0fe3c446c88f50",
			  "links": {
				"self": {
				  "href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/commit/88d0d4becc199c8c2005faebca0fe3c446c88f50"
				},
				"html": {
				  "href": "https://bitbucket.org/brian-nguyen/inertia-deploy-test/commits/88d0d4becc199c8c2005faebca0fe3c446c88f50"
				}
			  },
			  "author": {
				"raw": "brian-nguyen <briannguyen992@gmail.com>",
				"type": "author",
				"user": {
				  "username": "brian-nguyen",
				  "display_name": "Brian Nguyen",
				  "account_id": "557058:e73ba7e8-353b-4015-b7cd-77828b57dcad",
				  "links": {
					"self": {
					  "href": "https://api.bitbucket.org/2.0/users/brian-nguyen"
					},
					"html": {
					  "href": "https://bitbucket.org/brian-nguyen/"
					},
					"avatar": {
					  "href": "https://bitbucket.org/account/brian-nguyen/avatar/"
					}
				  },
				  "type": "user",
				  "uuid": "{0d8c0652-f421-44cc-a58b-2f1c8c09fe9f}"
				}
			  },
			  "summary": {
				"raw": "Local github test\n",
				"markup": "markdown",
				"html": "<p>Local github test</p>",
				"type": "rendered"
			  },
			  "parents": [
				{
				  "type": "commit",
				  "hash": "082d134b0f50a4a4fac3ce73292b68697decd97e",
				  "links": {
					"self": {
					  "href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/commit/082d134b0f50a4a4fac3ce73292b68697decd97e"
					},
					"html": {
					  "href": "https://bitbucket.org/brian-nguyen/inertia-deploy-test/commits/082d134b0f50a4a4fac3ce73292b68697decd97e"
					}
				  }
				}
			  ],
			  "date": "2018-07-07T06:03:34+00:00",
			  "message": "Local github test\n",
			  "type": "commit"
			}
		  },
		  "links": {
			"commits": {
			  "href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/commits?include=f7da6e2506829ef3ee8e3f1a2bfae534a5ab5dfa&exclude=88d0d4becc199c8c2005faebca0fe3c446c88f50"
			},
			"html": {
			  "href": "https://bitbucket.org/brian-nguyen/inertia-deploy-test/branches/compare/f7da6e2506829ef3ee8e3f1a2bfae534a5ab5dfa..88d0d4becc199c8c2005faebca0fe3c446c88f50"
			},
			"diff": {
			  "href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/diff/f7da6e2506829ef3ee8e3f1a2bfae534a5ab5dfa..88d0d4becc199c8c2005faebca0fe3c446c88f50"
			}
		  },
		  "truncated": false,
		  "commits": [
			{
			  "hash": "f7da6e2506829ef3ee8e3f1a2bfae534a5ab5dfa",
			  "links": {
				"self": {
				  "href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/commit/f7da6e2506829ef3ee8e3f1a2bfae534a5ab5dfa"
				},
				"comments": {
				  "href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/commit/f7da6e2506829ef3ee8e3f1a2bfae534a5ab5dfa/comments"
				},
				"patch": {
				  "href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/patch/f7da6e2506829ef3ee8e3f1a2bfae534a5ab5dfa"
				},
				"html": {
				  "href": "https://bitbucket.org/brian-nguyen/inertia-deploy-test/commits/f7da6e2506829ef3ee8e3f1a2bfae534a5ab5dfa"
				},
				"diff": {
				  "href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/diff/f7da6e2506829ef3ee8e3f1a2bfae534a5ab5dfa"
				},
				"approve": {
				  "href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/commit/f7da6e2506829ef3ee8e3f1a2bfae534a5ab5dfa/approve"
				},
				"statuses": {
				  "href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/commit/f7da6e2506829ef3ee8e3f1a2bfae534a5ab5dfa/statuses"
				}
			  },
			  "author": {
				"raw": "brian-nguyen <briannguyen992@gmail.com>",
				"type": "author",
				"user": {
				  "username": "brian-nguyen",
				  "display_name": "Brian Nguyen",
				  "account_id": "557058:e73ba7e8-353b-4015-b7cd-77828b57dcad",
				  "links": {
					"self": {
					  "href": "https://api.bitbucket.org/2.0/users/brian-nguyen"
					},
					"html": {
					  "href": "https://bitbucket.org/brian-nguyen/"
					},
					"avatar": {
					  "href": "https://bitbucket.org/account/brian-nguyen/avatar/"
					}
				  },
				  "type": "user",
				  "uuid": "{0d8c0652-f421-44cc-a58b-2f1c8c09fe9f}"
				}
			  },
			  "summary": {
				"raw": "Local parse test\n",
				"markup": "markdown",
				"html": "<p>Local parse test</p>",
				"type": "rendered"
			  },
			  "parents": [
				{
				  "type": "commit",
				  "hash": "a29ac9c00794f1325182d6afdbafad2558b64c85",
				  "links": {
					"self": {
					  "href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/commit/a29ac9c00794f1325182d6afdbafad2558b64c85"
					},
					"html": {
					  "href": "https://bitbucket.org/brian-nguyen/inertia-deploy-test/commits/a29ac9c00794f1325182d6afdbafad2558b64c85"
					}
				  }
				}
			  ],
			  "date": "2018-07-07T20:56:43+00:00",
			  "message": "Local parse test\n",
			  "type": "commit"
			},
			{
			  "hash": "a29ac9c00794f1325182d6afdbafad2558b64c85",
			  "links": {
				"self": {
				  "href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/commit/a29ac9c00794f1325182d6afdbafad2558b64c85"
				},
				"comments": {
				  "href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/commit/a29ac9c00794f1325182d6afdbafad2558b64c85/comments"
				},
				"patch": {
				  "href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/patch/a29ac9c00794f1325182d6afdbafad2558b64c85"
				},
				"html": {
				  "href": "https://bitbucket.org/brian-nguyen/inertia-deploy-test/commits/a29ac9c00794f1325182d6afdbafad2558b64c85"
				},
				"diff": {
				  "href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/diff/a29ac9c00794f1325182d6afdbafad2558b64c85"
				},
				"approve": {
				  "href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/commit/a29ac9c00794f1325182d6afdbafad2558b64c85/approve"
				},
				"statuses": {
				  "href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/commit/a29ac9c00794f1325182d6afdbafad2558b64c85/statuses"
				}
			  },
			  "author": {
				"raw": "brian-nguyen <briannguyen992@gmail.com>",
				"type": "author",
				"user": {
				  "username": "brian-nguyen",
				  "display_name": "Brian Nguyen",
				  "account_id": "557058:e73ba7e8-353b-4015-b7cd-77828b57dcad",
				  "links": {
					"self": {
					  "href": "https://api.bitbucket.org/2.0/users/brian-nguyen"
					},
					"html": {
					  "href": "https://bitbucket.org/brian-nguyen/"
					},
					"avatar": {
					  "href": "https://bitbucket.org/account/brian-nguyen/avatar/"
					}
				  },
				  "type": "user",
				  "uuid": "{0d8c0652-f421-44cc-a58b-2f1c8c09fe9f}"
				}
			  },
			  "summary": {
				"raw": "Github remote test\n",
				"markup": "markdown",
				"html": "<p>Github remote test</p>",
				"type": "rendered"
			  },
			  "parents": [
				{
				  "type": "commit",
				  "hash": "782fc00feb08df381c7a7d94f52d32cf46fb4065",
				  "links": {
					"self": {
					  "href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/commit/782fc00feb08df381c7a7d94f52d32cf46fb4065"
					},
					"html": {
					  "href": "https://bitbucket.org/brian-nguyen/inertia-deploy-test/commits/782fc00feb08df381c7a7d94f52d32cf46fb4065"
					}
				  }
				}
			  ],
			  "date": "2018-07-07T06:16:38+00:00",
			  "message": "Github remote test\n",
			  "type": "commit"
			},
			{
			  "hash": "782fc00feb08df381c7a7d94f52d32cf46fb4065",
			  "links": {
				"self": {
				  "href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/commit/782fc00feb08df381c7a7d94f52d32cf46fb4065"
				},
				"comments": {
				  "href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/commit/782fc00feb08df381c7a7d94f52d32cf46fb4065/comments"
				},
				"patch": {
				  "href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/patch/782fc00feb08df381c7a7d94f52d32cf46fb4065"
				},
				"html": {
				  "href": "https://bitbucket.org/brian-nguyen/inertia-deploy-test/commits/782fc00feb08df381c7a7d94f52d32cf46fb4065"
				},
				"diff": {
				  "href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/diff/782fc00feb08df381c7a7d94f52d32cf46fb4065"
				},
				"approve": {
				  "href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/commit/782fc00feb08df381c7a7d94f52d32cf46fb4065/approve"
				},
				"statuses": {
				  "href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/commit/782fc00feb08df381c7a7d94f52d32cf46fb4065/statuses"
				}
			  },
			  "author": {
				"raw": "brian-nguyen <briannguyen992@gmail.com>",
				"type": "author",
				"user": {
				  "username": "brian-nguyen",
				  "display_name": "Brian Nguyen",
				  "account_id": "557058:e73ba7e8-353b-4015-b7cd-77828b57dcad",
				  "links": {
					"self": {
					  "href": "https://api.bitbucket.org/2.0/users/brian-nguyen"
					},
					"html": {
					  "href": "https://bitbucket.org/brian-nguyen/"
					},
					"avatar": {
					  "href": "https://bitbucket.org/account/brian-nguyen/avatar/"
					}
				  },
				  "type": "user",
				  "uuid": "{0d8c0652-f421-44cc-a58b-2f1c8c09fe9f}"
				}
			  },
			  "summary": {
				"raw": "Enforce JSON for Github\n",
				"markup": "markdown",
				"html": "<p>Enforce JSON for Github</p>",
				"type": "rendered"
			  },
			  "parents": [
				{
				  "type": "commit",
				  "hash": "88d0d4becc199c8c2005faebca0fe3c446c88f50",
				  "links": {
					"self": {
					  "href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/commit/88d0d4becc199c8c2005faebca0fe3c446c88f50"
					},
					"html": {
					  "href": "https://bitbucket.org/brian-nguyen/inertia-deploy-test/commits/88d0d4becc199c8c2005faebca0fe3c446c88f50"
					}
				  }
				}
			  ],
			  "date": "2018-07-07T06:15:45+00:00",
			  "message": "Enforce JSON for Github\n",
			  "type": "commit"
			}
		  ],
		  "created": false,
		  "closed": false,
		  "new": {
			"type": "branch",
			"name": "master",
			"links": {
			  "commits": {
				"href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/commits/master"
			  },
			  "self": {
				"href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/refs/branches/master"
			  },
			  "html": {
				"href": "https://bitbucket.org/brian-nguyen/inertia-deploy-test/branch/master"
			  }
			},
			"target": {
			  "hash": "f7da6e2506829ef3ee8e3f1a2bfae534a5ab5dfa",
			  "links": {
				"self": {
				  "href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/commit/f7da6e2506829ef3ee8e3f1a2bfae534a5ab5dfa"
				},
				"html": {
				  "href": "https://bitbucket.org/brian-nguyen/inertia-deploy-test/commits/f7da6e2506829ef3ee8e3f1a2bfae534a5ab5dfa"
				}
			  },
			  "author": {
				"raw": "brian-nguyen <briannguyen992@gmail.com>",
				"type": "author",
				"user": {
				  "username": "brian-nguyen",
				  "display_name": "Brian Nguyen",
				  "account_id": "557058:e73ba7e8-353b-4015-b7cd-77828b57dcad",
				  "links": {
					"self": {
					  "href": "https://api.bitbucket.org/2.0/users/brian-nguyen"
					},
					"html": {
					  "href": "https://bitbucket.org/brian-nguyen/"
					},
					"avatar": {
					  "href": "https://bitbucket.org/account/brian-nguyen/avatar/"
					}
				  },
				  "type": "user",
				  "uuid": "{0d8c0652-f421-44cc-a58b-2f1c8c09fe9f}"
				}
			  },
			  "summary": {
				"raw": "Local parse test\n",
				"markup": "markdown",
				"html": "<p>Local parse test</p>",
				"type": "rendered"
			  },
			  "parents": [
				{
				  "type": "commit",
				  "hash": "a29ac9c00794f1325182d6afdbafad2558b64c85",
				  "links": {
					"self": {
					  "href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test/commit/a29ac9c00794f1325182d6afdbafad2558b64c85"
					},
					"html": {
					  "href": "https://bitbucket.org/brian-nguyen/inertia-deploy-test/commits/a29ac9c00794f1325182d6afdbafad2558b64c85"
					}
				  }
				}
			  ],
			  "date": "2018-07-07T20:56:43+00:00",
			  "message": "Local parse test\n",
			  "type": "commit"
			}
		  }
		}
	  ]
	},
	"repository": {
	  "scm": "git",
	  "website": "",
	  "name": "inertia-deploy-test",
	  "links": {
		"self": {
		  "href": "https://api.bitbucket.org/2.0/repositories/brian-nguyen/inertia-deploy-test"
		},
		"html": {
		  "href": "https://bitbucket.org/brian-nguyen/inertia-deploy-test"
		},
		"avatar": {
		  "href": "https://bytebucket.org/ravatar/%7Be49b6fff-e6cd-4d32-bd4d-cda24f99384b%7D?ts=default"
		}
	  },
	  "full_name": "brian-nguyen/inertia-deploy-test",
	  "owner": {
		"username": "brian-nguyen",
		"display_name": "Brian Nguyen",
		"account_id": "557058:e73ba7e8-353b-4015-b7cd-77828b57dcad",
		"links": {
		  "self": {
			"href": "https://api.bitbucket.org/2.0/users/brian-nguyen"
		  },
		  "html": {
			"href": "https://bitbucket.org/brian-nguyen/"
		  },
		  "avatar": {
			"href": "https://bitbucket.org/account/brian-nguyen/avatar/"
		  }
		},
		"type": "user",
		"uuid": "{0d8c0652-f421-44cc-a58b-2f1c8c09fe9f}"
	  },
	  "type": "repository",
	  "is_private": false,
	  "uuid": "{e49b6fff-e6cd-4d32-bd4d-cda24f99384b}"
	},
	"actor": {
	  "username": "brian-nguyen",
	  "display_name": "Brian Nguyen",
	  "account_id": "557058:e73ba7e8-353b-4015-b7cd-77828b57dcad",
	  "links": {
		"self": {
		  "href": "https://api.bitbucket.org/2.0/users/brian-nguyen"
		},
		"html": {
		  "href": "https://bitbucket.org/brian-nguyen/"
		},
		"avatar": {
		  "href": "https://bitbucket.org/account/brian-nguyen/avatar/"
		}
	  },
	  "type": "user",
	  "uuid": "{0d8c0652-f421-44cc-a58b-2f1c8c09fe9f}"
	}
}`)
