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
* [ ] update `contrib/npm` version

Update distribution streams:

* [ ] draft changelog
* [ ] make release and let builds run
* [ ] update https://github.com/ubclaunchpad/homebrew-tap
* [ ] update https://github.com/ubclaunchpad/scoop-bucket
* [ ] run `npm publish` in `contrib/npm`

---

**Draft Changelog**

![Publish (release)](https://github.com/ubclaunchpad/inertia/workflows/Publish%20(release)/badge.svg) ![Publish (latest)](https://github.com/ubclaunchpad/inertia/workflows/Publish%20(latest)/badge.svg)

TODO

## ⚠️ Breaking Changes 

* TODO

## 🎉 Enhancements

* TODO

## ⚒ Fixes 

* TODO

---

Please refer to the [complete diff](https://github.com/ubclaunchpad/inertia/compare/PREV...NEW) for more details.
