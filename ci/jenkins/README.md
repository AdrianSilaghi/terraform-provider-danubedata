# Terraform provider Jenkins migration

These pipelines move the Terraform provider from its offline self-hosted
GitHub Actions runner to the existing serialized Jenkins host builder. Jenkins
now owns pull-request and `main` build, test, and lint execution. GitHub Actions
remains only for tag/manual releases until the Jenkins release path is proven.

## Source and job contract

Jenkins Job DSL owns job creation, triggers, folder credentials, and the fixed
SCM definition. Every job loads its Jenkinsfile from protected `main`; build
parameters select only the separately fetched and verified source revision.
The globally configured DanubeData `TRUSTED_PIPELINE_REF=master` is unrelated
and is intentionally not consumed by these provider pipelines.

| Exact job | Trusted pipeline | Initial mode |
| --- | --- | --- |
| `terraform-provider-danubedata-pr/build` | `ci/jenkins/Jenkinsfile.build` | Webhook, same-repository collaborator PRs only |
| `terraform-provider-danubedata-build/main` | `ci/jenkins/Jenkinsfile.build` | Protected `main` push webhook |
| `terraform-provider-danubedata-acceptance/main` | `ci/jenkins/Jenkinsfile.acceptance` | Manual/scheduled, disabled by default |
| `terraform-provider-danubedata-release/github` | `ci/jenkins/Jenkinsfile.release` | Credential-free snapshot only; publishing disabled |

The PR webhook supplies `SOURCE_REF=refs/pull/<number>/merge`, the immutable
head `SOURCE_SHA`, `CHANGE_ID`, `CHANGE_TARGET=main`, the event repository, the
PR head repository, and GitHub author association. The pipeline requires a
two-parent synthetic merge, checks its second parent against the supplied head,
and requires its base parent to be trusted `main` history. It accepts only the
canonical head repository and `OWNER`, `MEMBER`, or `COLLABORATOR`; public fork
PRs need disposable isolation or a separate GitHub-hosted workflow.

The main webhook supplies `refs/heads/main` and its immutable tip. Jenkins
fetches the canonical repository over public HTTPS and requires the fetched
tip, checkout, and `SOURCE_SHA` to be identical. No job accepts an arbitrary
repository URL, branch, wildcard PR refspec, or Jenkinsfile from the revision
being tested.

## Phase 1 build and status behavior

The trusted Jenkinsfile pins the Linux/amd64 Go 1.24.12 and golangci-lint 1.64.5
images by manifest digest. Pull requests cannot choose or build a privileged
tool image. Each run uses ephemeral in-container Go and lint caches, then runs:

1. `go mod verify`
2. `go build -v ./...`
3. race-enabled tests with atomic `coverage.out`
4. pinned `golangci-lint run --timeout=5m`

`coverage.out` is archived in Jenkins. Codecov upload is deferred and is not a
gate. A future Jenkins upload must remain best-effort and must not expose a
token to pull-request code.

After source verification, the trusted Jenkinsfile briefly binds the
folder-scoped `terraform-provider-danubedata-github-status-token` and publishes
through a fixed GitHub REST endpoint. The credential scope contains only an
absolute `/usr/bin/jq` to `/usr/bin/curl` pipeline; it does not execute checked-out
code or read a workspace-controlled payload file. HTTP or JSON failures fail the
build. It publishes:

- `ci/jenkins/pr-merge` on the verified synthetic merge commit for visibility only and never as a required branch-protection context;
- `ci/jenkins/pr` on the verified second-parent head commit (authoritative);
- `ci/jenkins/main` on the exact protected-main tip.

The merge visibility result is sent first and the authoritative head result
last. Strict up-to-date branch protection is required so a previously green
head cannot remain mergeable after `main` advances. If authoritative publication
fails after a green response may have been committed, the pipeline corrects the
head to failure and rethrows the original failure. The provider status
credential must be restricted to this repository with only **Commit statuses: read and write** permission. Separate credentials with the same folder-local ID
belong in the provider PR and build folders; acceptance and release jobs do not
receive it.

Status publication is fail-closed: if the merge visibility status cannot be
published, Jenkins marks the required head status failed and aborts the build.

## Acceptance boundary

Acceptance is not a CI-cutover gate. The job is disabled by default in Job DSL,
serialized, and also fails unless an operator explicitly sets
`RUN_ACCEPTANCE=true`. It then verifies an exact protected-main tip before
binding `terraform-provider-acceptance-api-token`. A missing or empty
`DANUBEDATA_API_TOKEN` is a hard failure, never a successful skip.

Enable this only after provisioning a dedicated quota-limited test tenant with
cleanup, cost alerts, and an agreed schedule. Keep its credential in the
acceptance folder; PR and ordinary main jobs must not inherit it.

## Release boundary

Release requests use a full `refs/tags/v<semver>` ref and immutable target SHA.
The pipeline rejects malformed SemVer (including leading-zero numeric
prerelease identifiers), requires the exact tag object to resolve to that SHA,
and requires the commit to be reachable from protected `main`. It also verifies
the Go module, provider Registry address, GoReleaser binary convention, protocol
manifest, and module checksums.

The default `PUBLISH_RELEASE=false` path runs the pinned GoReleaser image as
`release --snapshot --clean --skip=sign` without credentials. Real publication
is hard-disabled by `REAL_RELEASE_ENABLED=false` in protected pipeline code and
must fail before any secret binding. A future reviewed enablement may use only:

| Credential ID | Type and scope |
| --- | --- |
| `terraform-provider-github-release-token` | Secret text; this repository's release contents only |
| `terraform-provider-gpg-private-key` | Secret file; dedicated provider signing key |
| `terraform-provider-gpg-passphrase` | Secret text; empty only if the dedicated key is intentionally unencrypted |

The future real path imports the key into a per-build temporary `GNUPGHOME` and
removes it unconditionally. GoReleaser receives the passphrase through stdin,
not a command argument.

The existing Terraform Registry webhook on the GitHub repository is part of the
release contract. Jenkins creates the signed GitHub release; it does not replace
or duplicate that webhook. After publication, verify webhook delivery and that
the exact version appears on the public Registry before calling the migration
proven. The current 0.3.4 incident history in `CHANGELOG.md` remains relevant:
do not re-add the Registry manifest as a release asset until its ingestion
behavior is understood.

## Repository rule preconditions

These files define required controls; they do not assert that the GitHub rules are currently configured. Before enabling any Jenkins jobs or webhook triggers,
an operator must verify a `main` branch ruleset that:

- blocks direct pushes and force pushes;
- applies to administrators and grants no routine bypass path;
- requires pull requests, strict up-to-date checking, and makes
  `ci/jenkins/pr` a required status context;
- requires CODEOWNER review through GitHub's **Require review from Code Owners**
  setting, using `.github/CODEOWNERS` for the trusted Jenkinsfiles, their
  validator, `.goreleaser.yml`, and the CODEOWNERS policy itself.

Record the ruleset review as cutover evidence. Do not treat CODEOWNERS alone as
enforcement: GitHub only enforces owner approval when the matching branch rule
requires Code Owner reviews.

## Cutover checklist

1. Run `bash tests/jenkins/validate-pipelines.sh` and review image digests.
2. Apply Job DSL with exact job names, protected-main SCM paths, distinct
   webhook selector tokens, and the new jobs initially disabled.
3. Verify every repository-rule precondition above before enabling jobs or
   triggers.
4. Prove one passing and one intentionally failing same-repository PR; confirm
   `ci/jenkins/pr` targets the head and `ci/jenkins/pr-merge` targets the merge.
5. Prove the exact-tip main build and credential-free release snapshot.
6. Require only `ci/jenkins/pr`; never require `ci/jenkins/pr-merge`.
7. Keep `.github/workflows/release.yml` until the Jenkins release path has
   proven equivalent behavior. The replaced PR/main `test.yml` is retired.
8. Provision and test release credentials, deliberately enable real release in
   protected code, publish one provider version, and verify the GitHub release,
   signature, Registry webhook, and Registry version.
9. Retire only replaced workflow files and remove the old provider runner after
   the full cutover proof; preserve unrelated runners and `.github` content.
