# Update Summary: Vulnerability Remediation and CI Workflow Modernization

## Scope
This update includes:
1. Remediation of all currently fixable vulnerabilities detected in the local container image scan.
2. CI workflow modernization to explicitly use Node.js 24 for workers.
3. Upgrade of GitHub Actions in workflows to their latest release versions.

---

## 1) Security Remediation

### Go dependency remediation
Updated vulnerable dependency in `go.mod`:

- `github.com/sirupsen/logrus`
  - **from:** `v1.9.0`
  - **to:** `v1.9.1`
  - **reason:** resolves `CVE-2025-65637` (High)

### Module graph refresh
Ran module maintenance to apply and lock dependency updates:

- `go mod tidy`

### Validation
- `go test ./...` ✅ passed
- `docker build -t comp2unraid:local .` ✅ succeeded
- `docker scout quickview comp2unraid:local` and `docker scout cves comp2unraid:local` ✅ completed

### Security scan results (after patch)
- **Previous:** `1 High`, `1 Medium`
- **Current:** `0 High`, `1 Medium`

### Remaining vulnerability
One non-fixable vulnerability remains in base OS package:

- Package: `busybox@1.37.0-r30`
- CVE: `CVE-2025-60876`
- Severity: Medium
- Status: **No fixed version available** at scan time

---

## 2) Workflow Update: Node.js 24 for Workers

Updated workflows to ensure Node.js 24 is configured explicitly:

- `.github/workflows/go.yml`
- `.github/workflows/docker-publish.yml`

Added setup step:

- `actions/setup-node` with `node-version: "24"`

Added workflow-level enforcement variable in both workflows:

- `FORCE_JAVASCRIPT_ACTIONS_TO_NODE24: "true"`

Rationale: workflow execution baseline is aligned with Node.js 24 requirement; Node.js versions below 24 are considered obsolete for this CI policy.

---

## 3) GitHub Actions Upgrades to Latest Releases

### `.github/workflows/go.yml`
- `actions/checkout` → `v6.0.2`
- `actions/setup-go` → `v6.4.0`
- `actions/setup-node` → `v6.3.0`

### `.github/workflows/docker-publish.yml`
- `actions/checkout` → `v6.0.2`
- `actions/setup-node` → `v6.3.0`
- `sigstore/cosign-installer` → `v4.1.1`
- `docker/setup-qemu-action` → `v4.0.0`
- `docker/setup-buildx-action` → `v4.0.0`
- `docker/login-action` → `v4.1.0`
- `docker/metadata-action` → `v6.0.0`
- `docker/build-push-action` → `v7.1.0`

Additional cleanup:
- Replaced older pinned commit-style action references with current release tags to match requested “latest release versions” policy.

---

## Outcome

- All currently **fixable** vulnerabilities identified in this pass were remediated.
- CI workflows now explicitly target **Node.js 24** worker setup.
- Workflow actions were upgraded to latest releases across Go and Docker pipelines.
- One Medium vulnerability remains (`busybox`, no upstream fix yet) and should be revisited when a patched base package/image becomes available.