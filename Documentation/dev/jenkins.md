# Jenkins

Jenkins is heavily used by Metering for continuous integration.
Its primary duties are:

- Building the metering binary
- Running metering unit tests
- Building docker images for all of the metering components
- Pushing docker images for each component to quay.io
- Running integration and e2e tests
  - Deploys Metering to a real Kubernetes cluster in a new namespace, and runs reports.
  - By default, it will delete the namespace it used to deploy after it runs tests.

## Pull-Requests

Every time you submit a pull-request, Jenkins will perform all of the above steps automatically.
If any step fails, it will generally stop at that point immediately and mark the commit as failing in Github.

### Pull-request debugging

To determine what step failed, you should view the job in Jenkins, and ***REMOVED***nd out what stage failed.

After determining what stage failed, you have a few different options for debugging it.
The ***REMOVED***rst step should always be to look at the console output for the job that failed and look at the logs to see if there's an error.
This should generally work for most issues that happen before the e2e and integration tests run.

#### Debugging e2e/integration tests

For e2e and integration tests, debugging can be a bit more involved.

- If the e2e/integration tests timeout, look at the logs to see if the pods every became ready, or if it got to the point where it runs tests.
- If the pods never became ready, then take a look at the "build artifacts" for that job in Jenkins.
- The build artifacts will contain various log ***REMOVED***les for the job.
- If the pods never became ready, then you should start by looking at the deploy log, and see if anything stands out.
- If the deploy log isn't showing anything useful, or it gets to the point where pods are created, but some don't become ready, then look at the pod logs ***REMOVED***le.

## Using images built by Jenkins

When developing a new feature, it's very common to have a workflow similar to this:

- Make changes to the repo
- Submit a pull-request
- Wait for the build to push the image
- Manually install metering with your PR's image tag to a Kubernetes cluster to see if they work how you expect

This process is useful when trying to debug an issue causing the e2e deployment or tests to fail.

To do this, it's actually fairly simple.
Since we push images for every build, all you need to do is follow the [Dev Installation](developer-guide.md#developer-install) documentation to use a custom image tag.
We publish multiple image tags for each PR build.
If you want to use the latest image for a PR, the image tag is `pr-$PR_NUMBER`, just replace `$PR_NUMBER` with your pull-request number.
