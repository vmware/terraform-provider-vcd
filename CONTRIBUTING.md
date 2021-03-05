# Contributing to terraform-provider-vcd

The **terraform-provider-vcd** project team welcomes contributions from the community. 
Before you start working with any contribution, please read our [Developer Certificate of Origin](https://cla.vmware.com/dco).
All contributions to this repository must be signed as described on that page. Your signature certifies that you wrote the patch or have the right to pass it on as an open-source patch.


## Community

Terraform-provider-vcd contributors can discuss matters here:
https://vmwarecode.slack.com, channel `#vcd-terraform-dev`

## Code Contribution Flow

We use GitHub pull requests to incorporate code changes from external
contributors. Typical contribution flow steps are:

- Fork the terraform-provider-vcd repo into a new repo on GitHub
- Clone the forked repo locally and set the original terraform-provider-vcd repo as the upstream repo
- Open an Issue in terraform-provider-vcd describing what you propose to do (unless the change is so trivial that an issue is not needed)
- Wait for discussion and possible direction hints in the issue thread
- Once you know  which steps to take in your intended contribution, make changes in a topic branch and commit (don't forget to add or modify tests too)
- Update Go modules files `go.mod` and `go.sum` if you're changing dependencies.
- Fetch changes from upstream and resolve any merge conflicts so that your topic branch is up-to-date
- Push all commits to the topic branch in your forked repo
- Submit a pull request to merge topic branch commits to upstream master 

Example:

``` shell
git remote add upstream https://github.com/vmware/terraform-provider-vcd.git
git checkout -b my-new-feature
git add filename1 [filename2 filename3]
git commit --signoff filename1 [filename2 filename3]
git push origin my-new-feature
```

If this process sounds unfamiliar have a look at the
excellent [overview of collaboration via pull requests on
GitHub](https://help.github.com/categories/collaborating-with-issues-and-pull-requests) for more information.

## Coding Style

Our standard for Golang contributions is to match the format of the [standard
Go package library](https://golang.org/pkg).

- Run `go fmt` on all code with latest stable version of Go (`go fmt` results may vary between Go versions).
- All public interfaces, functions, and structs must have complete, grammatically correct Godoc comments that explain their purpose and proper usage.
- Use self-explanatory names for all variables, functions, and interfaces.
- Add comments for non-obvious features of internal implementations but otherwise let the code explain itself.
- Include unit tests for new features and update tests for old ones. Refer to the [testing guide](TESTING.md) for more details.

Go is pretty readable so if you follow these rules most functions
will not need additional comments.

See [**CODING_GUIDELINES**](CODING_GUIDELINES.md) for more advice on how to write code for this project.

### Commit Message Format

We follow the conventions on [How to Write a Git Commit Message](http://chris.beams.io/posts/git-commit/).

Be sure to include any related GitHub
issue references in the commit message.  See [GFM
syntax](https://guides.github.com/features/mastering-markdown/#GitHub-flavored-markdown)
for referencing issues.

### Staying In Sync With Upstream

When your branch gets out of sync with the vmware master branch, use the following to update:

``` shell
git checkout master
git pull upstream master
git push
# At this point, your local copy of the master branch is synchronized
git checkout my-new-feature
git merge master
```
If there are conflicts you'll need to [merge them now](https://stackoverflow.com/questions/161813/how-to-resolve-merge-conflicts-in-git).

### Updating pull requests

If your PR fails to pass CI, or if you get change requests from the reviewers, you need to apply fixes in your local
repository, and then submit the changes.

``` shell
# edit files
git add filename
git commit -v --signoff filename
git push origin my-new-feature
```

Be sure to add a comment to the PR indicating your new changes are ready to review, as GitHub does not generate a
notification when you run `git push`.

### Formatting Commit Messages

We follow the conventions on [How to Write a Git Commit Message](http://chris.beams.io/posts/git-commit/).

Be sure to include any related GitHub issue references in the commit message.  See
[GFM syntax](https://guides.github.com/features/mastering-markdown/#GitHub-flavored-markdown) for referencing issues
and commits.

## Logging Bugs

Anyone can log a bug using the GitHub 'New Issue' button.  Please use
a short title and give as much information as you can about what the
problem is, relevant software versions, and how to reproduce it.  If you
know of a fix or a workaround include that too.

## Final Words

Thanks for helping us make the project better!
