# claude-runner

A Go service that runs [Claude Code](https://docs.anthropic.com/en/docs/claude-code) (`claude -p`) remotely over NATS and HTTP. Use it for plain prompts, pull request reviews, and GitHub issue tasks where Claude implements the work and opens a PR.

## Install

```bash
# Server
go install github.com/flarexio/claude-runner/cmd/claude-runner@latest

# Client
go install github.com/flarexio/claude-runner/cmd/claude-runner-client@latest
```

Prerequisites: [Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code) installed and authenticated, Go 1.25+.

## Configure

Drop a `config.yaml` under `~/.flarex/claude-runner/`. Start from [`config.example.yaml`](./config.example.yaml) — every knob is commented inline.

For NATS transport, also drop two files in the same directory:

- `id` — plain text edge node ID
- `user.creds` — NATS credentials

The most security-relevant flag is `issue.bypassPermissions`: setting it to `true` passes `--dangerously-skip-permissions` to `claude`. Only enable that when the trigger is gated to trusted members (the `agent-ready` label gate in the [Issue Mode action](#issue-mode-action) is the canonical example) and the runner's GitHub token is restricted to non-destructive operations.

For service architecture, lifecycle, and the issue agent-task protocol, see [AGENTS.md](./AGENTS.md).

## Run

```bash
claude-runner            # NATS only
claude-runner --http     # also enable HTTP on :8080
```

### Docker

```bash
# NATS
docker run -d \
  -v ~/.claude:/root/.claude \
  -v ~/.flarex/claude-runner:/root/.flarex/claude-runner \
  flarexio/claude-runner

# HTTP
docker run -d \
  -v ~/.claude:/root/.claude \
  -v ~/.flarex/claude-runner:/root/.flarex/claude-runner \
  -p 8080:8080 \
  flarexio/claude-runner --http
```

## Call it

Two operations, two endpoints. Clients pick one; the server does no event-based dispatching.

| Operation | HTTP | NATS subject | Behavior |
| --- | --- | --- | --- |
| `Run` | `POST /api/run` | `<topic>.run` | Synchronous prompt / PR review |
| `RunIssue` | `POST /api/run-issue` | `<topic>.run-issue` | Sync claim → background execute → returns `accepted` |

### POST /api/run

```json
{
  "prompt": "Review this code for bugs",
  "repo": "git@github.com:user/repo.git",
  "ref": "feature/my-change",
  "base_ref": "main",
  "event": "pull_request",
  "pr_number": 2
}
```

Response is `{id, output, error}`. When `repo` is set the runner clones it into `workDir/<run-id>` and removes the clone after the run; when `repo` is empty the runner runs in `workDir` directly.

### POST /api/run-issue

```json
{
  "repo": "owner/repo",
  "issue_number": 42
}
```

Validates and claims synchronously, then runs Claude in the background. The HTTP response returns after the claim phase with `output: "...accepted; claude-runner is processing in the background."`. Final success/failure is posted as a comment on the GitHub issue. The server must have a GitHub token configured.

### CLI client

```bash
# Plain prompt
claude-runner-client \
  --transport http --endpoint http://localhost:8080 \
  --prompt "Review this code"

# Issue task
claude-runner-client \
  --transport http --endpoint http://localhost:8080 \
  --event issue \
  --repo https://github.com/user/repo.git \
  --issue-number 42
```

`--prompt` is ignored when `--event=issue`; the runner builds the prompt from the issue body. `--output-file` writes Claude's output to disk in addition to stdout.

## GitHub Action

Add `NATS_CREDS` (content of `user.creds`) and `EDGE_ID` to your repository's **Settings → Secrets → Actions**.

### CI / PR review

The same step works for `pull_request`, `push`, and `workflow_dispatch`. On a PR, `base-ref` and `pr-number` are populated and claude-runner generates a PR diff for Claude to review against.

```yaml
on:
  pull_request:
  push:
    branches: [main]
  workflow_dispatch:

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: flarexio/claude-runner@v1
        with:
          prompt: |
            Review all changed files in this repository.
            Provide a concise summary of findings.
          repo: ${{ github.server_url }}/${{ github.repository }}.git
          ref: ${{ github.head_ref || github.ref_name }}
          base-ref: ${{ github.base_ref }}
          event: ${{ github.event_name }}
          pr-number: ${{ github.event.pull_request.number || '' }}
          nats-creds-content: ${{ secrets.NATS_CREDS }}
          edge-id: ${{ secrets.EDGE_ID }}
          output-file: claude-output.md
```

### Issue mode <a id="issue-mode-action"></a>

The `if:` gate restricts who can trigger an agent run — only repository members with write access can apply labels.

```yaml
on:
  issues:
    types: [labeled]

jobs:
  agent:
    if: github.event.label.name == 'agent-ready'
    runs-on: ubuntu-latest
    steps:
      - uses: flarexio/claude-runner@v1
        with:
          event: issue
          repo: ${{ github.server_url }}/${{ github.repository }}.git
          issue-number: ${{ github.event.issue.number }}
          nats-creds-content: ${{ secrets.NATS_CREDS }}
          edge-id: ${{ secrets.EDGE_ID }}
```

The runner only acts on issues that pass validation: open, body contains `<!-- agent-task:v1 -->`, required labels (`type:agent-task`, `agent:claude-code`, `agent-ready`, `agent-approved`), no excluded labels (`claimed-by-claude`, `agent-blocked`, `security-sensitive`). Failed issue runs preserve the workspace under `workDir` for inspection — see [AGENTS.md](./AGENTS.md) for the on-disk layout.

## One-shot CLI: `claude-runner run-issue`

Runs the full GitHub issue agent-task protocol — claim, run Claude, post results — in a single foreground invocation. Designed to be invocable directly by AI agents as a skill.

```bash
claude-runner run-issue --repo owner/repo --issue-number 42
```

Useful flags: `--path` (config directory), `--ref` (branch), `--github-token` (overrides config token; env: `GITHUB_TOKEN`).

## License

MIT
