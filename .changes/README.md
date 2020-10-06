# Handling of CHANGELOG entries

* All PRs, instead of updating CHANGELOG.md directly, will create individual files in a directory .changes/$version (The version is the same contained in the VERSION file)
* The files will be named `xxx-section_name.md`, where xxx is the PR number, and `section_name` is one of 
    * features
    * improvements
    * bug-fixes
    * deprecations
    * notes
    * removals

* The changes files must NOT contain the header

* You can update the file `.changes/sections` to add more headers
* To see the full change set for the current version (or an old one), use `./scripts/make-changelog.sh [version]`

