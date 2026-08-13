#!/usr/bin/env bash

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "${ROOT}"

fail() {
    printf 'FAIL: %s\n' "$*" >&2
    exit 1
}

assert_contains() {
    local file="$1"
    local pattern="$2"
    grep -Eq -- "${pattern}" "${file}" || fail "${file} is missing /${pattern}/"
}

assert_not_contains() {
    local file="$1"
    local pattern="$2"
    if grep -Eq -- "${pattern}" "${file}"; then
        fail "${file} unexpectedly contains /${pattern}/"
    fi
}

assert_before() {
    local file="$1"
    local first_pattern="$2"
    local second_pattern="$3"
    local first_line second_line
    first_line="$(grep -nE -- "${first_pattern}" "${file}" | head -1 | cut -d: -f1)"
    second_line="$(grep -nE -- "${second_pattern}" "${file}" | head -1 | cut -d: -f1)"
    [ -n "${first_line}" ] && [ -n "${second_line}" ] && [ "${first_line}" -lt "${second_line}" ] || \
        fail "${file} does not place /${first_pattern}/ before /${second_pattern}/"
}

for file in \
    .github/CODEOWNERS \
    ci/jenkins/README.md \
    ci/jenkins/Jenkinsfile.build \
    ci/jenkins/Jenkinsfile.acceptance \
    ci/jenkins/Jenkinsfile.release; do
    [ -f "${file}" ] || fail "missing ${file}"
done

BUILD=ci/jenkins/Jenkinsfile.build
ACCEPTANCE=ci/jenkins/Jenkinsfile.acceptance
RELEASE=ci/jenkins/Jenkinsfile.release
DOC=ci/jenkins/README.md
CODEOWNERS=.github/CODEOWNERS

# Protected workflow surfaces require owner review once repository rules are configured.
assert_contains "${CODEOWNERS}" '^/ci/jenkins/[[:space:]]+@AdrianSilaghi$'
assert_contains "${CODEOWNERS}" '^/tests/jenkins/[[:space:]]+@AdrianSilaghi$'
assert_contains "${CODEOWNERS}" '^/\.goreleaser\.yml[[:space:]]+@AdrianSilaghi$'
assert_contains "${CODEOWNERS}" '^/\.github/CODEOWNERS[[:space:]]+@AdrianSilaghi$'

# Both PR and protected-main builds use one pipeline loaded from protected main.
assert_contains "${BUILD}" 'agent none'
assert_contains "${BUILD}" 'skipDefaultCheckout\(true\)'
assert_contains "${BUILD}" "JOB_NAME == 'terraform-provider-danubedata-pr/build'"
assert_contains "${BUILD}" "JOB_NAME == 'terraform-provider-danubedata-build/main'"
assert_contains "${BUILD}" "SOURCE_REPOSITORY.*AdrianSilaghi/terraform-provider-danubedata"
assert_contains "${BUILD}" "HEAD_REPOSITORY"
assert_contains "${BUILD}" "AUTHOR_ASSOCIATION"
assert_contains "${BUILD}" "\['OWNER', 'MEMBER', 'COLLABORATOR'\]"
assert_contains "${BUILD}" 'CHANGE_TARGET.*main'
assert_contains "${BUILD}" 'sourceRef = "refs/pull/\$\{changeId\}/merge"'
assert_contains "${BUILD}" 'refs/heads/main:refs/remotes/origin/main'
assert_contains "${BUILD}" 'EXPECTED_PR_HEAD_SHA = sourceSha'
assert_contains "${BUILD}" 'git show -s --format=%P HEAD'
assert_contains "${BUILD}" 'mergeParents.size\(\) != 2'
assert_contains "${BUILD}" 'String baseParent = mergeParents\[0\]'
assert_contains "${BUILD}" 'String headParent = mergeParents\[1\]'
assert_contains "${BUILD}" 'headParent.equalsIgnoreCase\(env.EXPECTED_PR_HEAD_SHA\)'
assert_contains "${BUILD}" 'git merge-base --is-ancestor.*baseParent.*VERIFIED_BASE_REF'
assert_not_contains "${BUILD}" "refs/pull/[^[:space:]\"']*/head|refs/pull/\\*|checkout scm"
assert_contains "${BUILD}" "sourceRef != 'refs/heads/main'"
assert_contains "${BUILD}" 'fetchedRefSha.equalsIgnoreCase\(env.EXPECTED_SOURCE_SHA\)'
assert_contains "${BUILD}" 'checkedOutSha.equalsIgnoreCase\(fetchedRefSha\)'

# Images and commands are fixed in the trusted Jenkinsfile, never selected by PR code.
assert_contains "${BUILD}" 'golang:1\.24\.12-bookworm@sha256:[0-9a-f]{64}'
assert_contains "${BUILD}" 'golangci/golangci-lint:v1\.64\.5-alpine@sha256:[0-9a-f]{64}'
assert_contains "${BUILD}" 'go mod verify'
assert_contains "${BUILD}" 'go build -v ./\.\.\.'
assert_contains "${BUILD}" 'go test -v -race -coverprofile=coverage\.out -covermode=atomic ./\.\.\.'
assert_contains "${BUILD}" 'golangci-lint run --timeout=5m'
assert_contains "${BUILD}" 'archiveArtifacts.*coverage\.out'
assert_not_contains "${BUILD}" 'Dockerfile|docker build|:latest|codecov-token|CODECOV_TOKEN'
[ "$(grep -Fc -- '--read-only' "${BUILD}")" -eq 2 ] || fail 'both provider build containers must use a read-only root filesystem'
[ "$(grep -Fc -- '--cap-drop ALL' "${BUILD}")" -eq 2 ] || fail 'both provider build containers must drop all capabilities'
[ "$(grep -Fc -- '--security-opt no-new-privileges' "${BUILD}")" -eq 2 ] || fail 'both provider build containers must prevent privilege escalation'
[ "$(grep -Fc -- '--tmpfs /tmp:rw,exec,nosuid,nodev,size=4g' "${BUILD}")" -eq 2 ] || fail 'both provider build containers must use bounded ephemeral scratch space'

# GitHub statuses use a folder-scoped token only around fixed trusted REST code.
assert_contains "${BUILD}" 'void publishGitHubCommitStatusForSha'
assert_contains "${BUILD}" "\['ci/jenkins/pr', 'ci/jenkins/pr-head', 'ci/jenkins/main'\]"
assert_contains "${BUILD}" "\['PENDING', 'SUCCESS', 'FAILURE', 'ERROR'\]"
assert_contains "${BUILD}" "credentialsId: 'terraform-provider-danubedata-github-status-token'"
assert_contains "${BUILD}" 'withCredentials\(\[string\('
assert_contains "${BUILD}" 'set \+x'
assert_contains "${BUILD}" '#!/bin/bash'
assert_contains "${BUILD}" '/usr/bin/jq -cn'
assert_contains "${BUILD}" '/usr/bin/curl --disable'
assert_contains "${BUILD}" 'https://api.github.com/repos/AdrianSilaghi/terraform-provider-danubedata/statuses/\$\{STATUS_TARGET_SHA\}'
assert_contains "${BUILD}" '--fail-with-body'
assert_contains "${BUILD}" 'payload="\$\(/usr/bin/jq -cn'
assert_contains "${BUILD}" '/usr/bin/curl --disable --config -'
# This is a literal contract for the trusted Jenkins shell block.
# shellcheck disable=SC2016
assert_contains "${BUILD}" '--data-binary "\$payload"'
assert_contains "${BUILD}" 'Authorization: Bearer %s'
assert_not_contains "${BUILD}" 'exec[[:space:]]+3<<<|/dev/fd/3'
assert_before "${BUILD}" 'targetSha ==~ /\[0-9a-fA-F\]\{40\}/' "credentialsId: 'terraform-provider-danubedata-github-status-token'"
assert_not_contains "${BUILD}" 'GitHubCommitStatusSetter|ManuallyEnteredRepositorySource|ManuallyEnteredShaSource|ManuallyEnteredCommitContextSource'
assert_not_contains "${BUILD}" 'github-status\.json|status-payload|writeFile.*status|--data-binary @[A-Za-z0-9_./]+'
assert_contains "${BUILD}" "publishGitHubCommitStatusForSha\(env.GITHUB_STATUS_HEAD_SHA, 'ci/jenkins/pr-head'"
assert_contains "${BUILD}" 'publishGitHubCommitStatusForSha\(env.GITHUB_STATUS_SHA, env.GITHUB_STATUS_CONTEXT'
assert_before "${BUILD}" \
    "publishGitHubCommitStatusForSha\(env.GITHUB_STATUS_HEAD_SHA, 'ci/jenkins/pr-head'" \
    'publishGitHubCommitStatusForSha\(env.GITHUB_STATUS_SHA'
assert_contains "${BUILD}" "IS_PULL_REQUEST == 'true'.*'ci/jenkins/pr'.*'ci/jenkins/main'"
assert_contains "${BUILD}" "publishGitHubCommitStatus\('PENDING'"
assert_contains "${BUILD}" 'cleanup \{'
assert_contains "${BUILD}" 'currentBuild.currentResult'
assert_contains "${BUILD}" "'ABORTED'.*'ERROR'"
assert_before "${BUILD}" 'archiveArtifacts.*coverage\.out' 'cleanup \{'
assert_before "${BUILD}" 'cleanup \{' '^[[:space:]]+publishGitHubCommitStatus\($'

# Acceptance is deliberately unusable by default and cannot turn a missing token into success.
assert_contains "${ACCEPTANCE}" "JOB_NAME != 'terraform-provider-danubedata-acceptance/main'"
assert_contains "${ACCEPTANCE}" "booleanParam\(name: 'RUN_ACCEPTANCE', defaultValue: false"
assert_contains "${ACCEPTANCE}" 'if \(!params.RUN_ACCEPTANCE\)'
assert_contains "${ACCEPTANCE}" "error\('Acceptance execution is disabled by default"
assert_contains "${ACCEPTANCE}" "credentialsId: 'terraform-provider-acceptance-api-token'"
# This is a literal contract for the Jenkins shell block.
# shellcheck disable=SC2016
assert_contains "${ACCEPTANCE}" '\[ -n "\$DANUBEDATA_API_TOKEN" \]'
assert_contains "${ACCEPTANCE}" 'TF_ACC=1'
assert_contains "${ACCEPTANCE}" 'go test -v -timeout 60m ./internal/resources/\.\.\. ./internal/datasources/\.\.\.'
assert_contains "${ACCEPTANCE}" 'disableConcurrentBuilds\(\)'
assert_not_contains "${ACCEPTANCE}" 'skipping acceptance|\|\| true'
assert_before "${ACCEPTANCE}" 'git merge-base --is-ancestor' 'withCredentials'

# Releases validate a full tag ref and immutable commit before any optional secret binding.
assert_contains "${RELEASE}" "JOB_NAME != 'terraform-provider-danubedata-release/github'"
assert_contains "${RELEASE}" "booleanParam\(name: 'PUBLISH_RELEASE', defaultValue: false"
assert_contains "${RELEASE}" "REAL_RELEASE_ENABLED = 'false'"
assert_contains "${RELEASE}" 'refs/tags/'
assert_contains "${RELEASE}" 'SOURCE_SHA'
assert_contains "${RELEASE}" 'git merge-base --is-ancestor.*VERIFIED_TAG_COMMIT.*VERIFIED_MAIN_REF'
assert_contains "${RELEASE}" 'module github.com/AdrianSilaghi/terraform-provider-danubedata'
assert_contains "${RELEASE}" 'terraform-provider-danubedata'
assert_contains "${RELEASE}" 'terraform-registry-manifest.json'
assert_contains "${RELEASE}" "agent \{ label 'trusted-build' \}"
assert_not_contains "${RELEASE}" 'agent none'
assert_contains "${RELEASE}" 'go mod download'
assert_contains "${RELEASE}" 'go mod tidy'
assert_contains "${RELEASE}" 'git diff --exit-code -- go\.mod go\.sum'
assert_before "${RELEASE}" 'go mod tidy' 'git diff --exit-code -- go\.mod go\.sum'
assert_contains "${RELEASE}" 'go mod verify'
# These are literal contracts for the Jenkins shell blocks.
# shellcheck disable=SC2016
assert_contains "${RELEASE}" 'GO_MOD_CACHE=.\$WORKSPACE_TMP/terraform-provider-go-mod.'
# shellcheck disable=SC2016
assert_contains "${RELEASE}" '-v .\$GO_MOD_CACHE:/tmp/go-mod.'
# shellcheck disable=SC2016
[ "$(grep -Fc -- '-v "$GORELEASER_HOME:/tmp/home"' "${RELEASE}")" -eq 2 ] || fail 'snapshot and real releases must mount a writable temporary HOME'
assert_contains "${RELEASE}" 'goreleaser/goreleaser:v2\.13\.3@sha256:[0-9a-f]{64}'
assert_contains "${RELEASE}" 'release --snapshot --clean --skip=sign'
assert_contains "${RELEASE}" "credentialsId: 'terraform-provider-github-release-token'"
assert_contains "${RELEASE}" "credentialsId: 'terraform-provider-gpg-private-key'"
assert_contains "${RELEASE}" "credentialsId: 'terraform-provider-gpg-passphrase'"
assert_contains "${RELEASE}" 'GNUPGHOME'
assert_contains "${RELEASE}" "trap 'rm -rf"
assert_before "${RELEASE}" 'git merge-base --is-ancestor' 'withCredentials'
assert_before "${RELEASE}" "REAL_RELEASE_ENABLED != 'true'" 'withCredentials'

# Migration/cutover and external registry behavior stay explicit.
assert_contains "${DOC}" 'terraform-provider-danubedata-pr/build'
assert_contains "${DOC}" 'terraform-provider-danubedata-build/main'
assert_contains "${DOC}" 'terraform-provider-danubedata-acceptance/main'
assert_contains "${DOC}" 'terraform-provider-danubedata-release/github'
assert_contains "${DOC}" 'protected.*main'
assert_contains "${DOC}" 'same-repository.*collaborator'
assert_contains "${DOC}" 'ci/jenkins/pr-head.*never.*required'
assert_contains "${DOC}" 'Commit statuses: read and write'
assert_contains "${DOC}" 'Codecov.*best-effort|Codecov.*deferred'
assert_contains "${DOC}" 'Terraform Registry.*webhook'
assert_contains "${DOC}" 'disabled by default'
assert_contains "${DOC}" 'quota-limited test tenant'
assert_contains "${DOC}" 'remains only for tag/manual releases'
assert_contains "${DOC}" 'Before enabling.*jobs'
assert_contains "${DOC}" 'direct pushes'
assert_contains "${DOC}" 'force pushes'
assert_contains "${DOC}" 'bypass'
assert_contains "${DOC}" 'ci/jenkins/pr.*required|required.*ci/jenkins/pr'
assert_contains "${DOC}" 'CODEOWNER review|Code Owner reviews'
assert_contains "${DOC}" 'do not assert.*currently configured|not assert.*currently configured'

# Phase 1 is additive; retirement occurs only after cutover proof.
[ ! -f .github/workflows/test.yml ] || fail 'test.yml must be retired after Jenkins PR/main cutover'
[ -f .github/workflows/release.yml ] || fail 'release.yml must remain until Jenkins release proof'
assert_contains .github/workflows/release.yml 'tags:'
assert_contains .github/workflows/release.yml "- 'v\*'"
assert_contains .github/workflows/release.yml 'workflow_dispatch:'
assert_contains .github/workflows/release.yml 'contents: write'
assert_contains .github/workflows/release.yml 'goreleaser release --clean'
assert_not_contains .github/workflows/release.yml 'pull_request|branches: \[main, master\]'

bash -n tests/jenkins/validate-pipelines.sh
shellcheck tests/jenkins/validate-pipelines.sh

printf 'Terraform provider Jenkins pipeline validation passed.\n'
