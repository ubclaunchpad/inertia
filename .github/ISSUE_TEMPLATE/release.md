---
name: Release tracker
about: Only for repository maintainers - describes TODOs for a release
title: 'release: inertia vNUMBER'
labels: ''
assignees: ''

---

**Milestone**

https://github.com/ubclaunchpad/inertia/milestones/NUMBER

**Tasks**

Make a new PR and merge when everything is ready:

* [ ] generate release documentation (`make docs`)
* [ ] generate new tip documentation (`make docs-tip`)

Update distribution streams:

* [ ] draft changelog using template below
* [ ] publish GitHub release with the changleog and let pipelines run
* [ ] ensure release is live in [releases](https://github.com/ubclaunchpad/inertia/releases), [packages](https://github.com/orgs/ubclaunchpad/packages/container/package/inertiad), [`ubclaunchpad/homebrew-tap`](https://github.com/`ubclaunchpad/homebrew-tap), and [`ubclaunchpad/scoop-bucket`](https://github.com/ubclaunchpad/scoop-bucket)

---

**Draft Changelog**

![Publish (release)](https://github.com/ubclaunchpad/inertia/workflows/Publish%20(release)/badge.svg) [![Docs](https://img.shields.io/website?label=docs&up_message=live&url=https%3A%2F%2Finertia.ubclaunchpad.com)](https://inertia.ubclaunchpad.com)

TODO

## ⚠️ Breaking Changes

* TODO

## 🎉 Enhancements

* TODO

## ⚒ Fixes

* TODO

---

Please refer to the [complete diff](https://github.com/ubclaunchpad/inertia/compare/PREV...NEW) for more details.
