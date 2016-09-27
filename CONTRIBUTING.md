ToDD Contribution Guide
====

If you're thinking about contributing, then before you write any code, please communicate with the ToDD community first. In these early days, ToDD is going through a lot of changes, and I would hate for you to waste your time writing code that may already be in progress, or no longer needed. For any immediate issues, please open a Github issue, and be specific about the problem you're having and/or would like to tackle. For longer conversations, please use the [mailing list](https://groups.google.com/forum/#!forum/todd-dev).

I am still building out the CI pipeline, but I will be fairly strict about contributions with respect to Golang idioms and proper unit/integration testing. Now that ToDD is open sourced, I want to focus on these much more, and as a result I want to make sure contributions help, not hurt this effort. Here are some resources that may help you in this regard:

- [https://golang.org/doc/code.html](https://golang.org/doc/code.html)
- [https://blog.golang.org/package-names](https://blog.golang.org/package-names)

My preferred methodology for contributing is still being worked out, so for the time being, err on the side of caution and start a conversation with me via a github issue first. I'll do my best to be responsive towards that medium.

Please refer to the .travis.yml file in the root of the repository in order to see the build steps being performed. I will not look at a PR until it produces a passing build, so to save time, try to run these build steps yourself on your own machine first.

Also, and this is important - the smaller and more single-purpose your PRs are, the better. I am happy to review just about any PR that passes the build and seems to be a good idea, but I do have a day job and if you want a quick turnaround, keep them small!

If you're wondering what there is to work on, I try to keep the [issues for this project](https://github.com/toddproject/todd/issues) populated with any known issues, so feel free to peruse that list.
