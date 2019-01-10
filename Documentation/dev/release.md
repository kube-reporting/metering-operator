# Releasing

Minor releases:

Let's say we're currently at version 0.7.0 and we want to release 0.8.0, the next minor version.

```
git checkout -b release-0.8
echo 0.8.0 > VERSION
git add VERSION
git commit -m 'VERSION: Bump to version 0.8.0'
make metering-manifests
git add manifests
git commit -m 'manifests: Regenerate manifests for 0.8.0'
```

Next create push your changes (requires access to create branches on origin):

```
git push origin release-0.8
```

This will trigger a new build in Jenkins for the new branch and it will validate things still work before we tag the release and build the ***REMOVED***nal images that the manifests are using.

For example, after pushing your local `release-0.8` to origin, the following new jenkins job will be created:

https://jenkins.prod.coreos.systems/job/metering/job/operator-metering-main/job/release-0.8/


## Veri***REMOVED***cation

Wait for the build to ***REMOVED***nish and pass.
If it fails for trivial reasons, eg test flakes, just re-run the build.

If it fails due to some real issue in the release, make the corrections like any other bug ***REMOVED***x and submit a pull-request against master.
Once any issues have been resolved against master, cherry-pick the change into the release branch using `git cherry-pick` and create a PR for those changes against the release branch.

### Manual testing

After we have a passing build in Jenkins, do any extra veri***REMOVED***cation like you would for any other PR.
Images built will be named after the branch (ex: `release-0.8`).
You will need to override the helm operator image tag, as well as the image tag for each other component if you wish to test these before the ***REMOVED***nal release is tagged and built.

For example:

```
unset METERING_CR_FILE
./hack/custom-metering-operator-install-wrapper.sh release-0.8 ./hack/openshift-install.sh
```

### Pulling in changes into a release branch

If a release branch has issues and needs a bug ***REMOVED***x or something ***REMOVED*** added to it before it can be released we have two potions

- Git merge
- Git cherry-pick

If the change we need merged shortly after the release branch was made and the master branch doesn't contain major changes, than we can do a git merge:

```
git checkout release-0.8
git fetch origin master
git merge origin/master
git push origin release-0.8
```

If there's been a lot of work on master and we want to extract individual changes, than we will use `git cherry-pick` to pull individual changes into the release branch.

For example if master has a commit `12345` that is a merge commit containing a change you need, you would cherry-pick it into your release branch like so (passing the correct value to -m):

```
git checkout release-0.8
git cherry-pick 12345 -x -m 1
git push origin release-0.8
```

## Tagging the release

After the team is con***REMOVED***dent that the release is ready and has no outstanding issues that are blocking it, then use `git tag -s` to tag the release, sign the tag, and provide any information about the release in the description.

First fetch the latest tags:

```
git fetch origin --tags
```

To see what has changed between a release, use git log:

```
git log --oneline release-0.7..release-0.8
```

See previous release tags for examples of a good message to put into the git tag:

```
git tag -l -n50
```

Then create the tag, write your message, and enter your GPG passphrase when prompted to sign the git tag:

```
git tag -s 0.8.0
```

After the tag is created and signed, push it to Github:

```
git push origin 0.8.0
```

This will not automatically trigger a Jenkins build, so to trigger it manually go to the jobs page in the `tags` tab:

https://jenkins.prod.coreos.systems/job/metering/job/operator-metering-main/view/tags/

Then select your version in the list of versions, it should take you to a job page.

Your URL should look like https://jenkins.prod.coreos.systems/job/metering/job/operator-metering-main/view/tags/job/0.8.0/

Finally, click `Build Now` or `Build With Parameters` on the left of the page, leave the defaults and click the `Build` button.

Wait for the build to ***REMOVED***nish, and once it's done the release images and been pushed and the tag exists on Github.


## Making the Release on Github

The last step for ***REMOVED***nalizing the release is to create a Github Release from our tag.

Go to https://github.com/operator-framework/operator-metering/releases and click the "Draft a new release" button at the top right.

In the text box, enter your release tag, eg: `0.8.0`.

Set the title `Metering $version` to eg: `Metering 0.8.0`.

You can leave the description empty, as it should just use our git tag's message.

Check  "this is a pre-release".

Click "Publish release".

# Last steps

If you've gotten this far, you've made a release.
The last and ***REMOVED***nal steps are to merge the release branch into master and to update `VERSION` in master to the next iteration version.

Make sure your local master is up to date:

```
git checkout master
git pull origin master
```

Next create a new branch for our changes, and merge the release branch into it:

```
git checkout -b bump_release
git fetch origin release-0.8
git merge origin/release-0.8
```

This last bit is *EXTREMELY IMPORTANT*.
This part is what ensures that we don't overwrite the 0.8.0 images in future builds:

Bump the `VERSION` ***REMOVED***le to have the `-latest` suf***REMOVED***x:

```
echo 0.8.0-latest > VERSION
git add VERSION
git commit -m 'VERSION: Bump to version 0.8.0-latest'
make metering-manifests
git add manifests
git commit -m 'manifests: Regenerate manifests for 0.8.0-latest'
```

Finally push your changes (replace `chancez` with your forks remote repository name):

```
git push chancez bump_release
```

And lastly create a pull-request to have the PR reviewed before it gets merged into master.
