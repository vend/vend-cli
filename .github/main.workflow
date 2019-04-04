workflow "Build and Test" {
  on = "pull_request"
  resolves = ["Build"]
}

action "Fmt" {
  uses = "leosunmo/go-github-actions/fmt@6_fmt_ignore_dirs"
  secrets = ["GITHUB_TOKEN"]
  env = {
    GO_WORKING_DIR = "./"
    GO_IGNORE_DIRS = "./vendor"
  }
}

action "Lint" {
  uses = "sjkaliski/go-github-actions/lint@v0.3.0"
  secrets = ["GITHUB_TOKEN"]
  needs = "Fmt"
  env = {
    GO_WORKING_DIR = "./"
    GO_LINT_PATHS = "./commands/... ./"
  }
}

action "Build" {
  uses = "cedrickring/golang-action/go1.12@1.2.0"
  needs = "Lint"
  args = "GOOS=darwin GOARCH=amd64 go build"
}

workflow "Publish" {
  on = "push"
  resolves = ["Publish Release"]
}
# Filter for master branch
action "Master" {
  uses = "actions/bin/filter@master"
  args = "branch master"
}

action "Merged" {
  needs = "Master"
  uses = "actions/bin/filter@master"
  args = "merged true"
}

action "Version Tag" {
  needs = "Merged"
  uses = "actions/bin/filter@master"
  args = "tag v*"
}

action "Create Archive" {
  needs = "Version Tag"
  uses = "lubusIN/actions/archive@master"
  env = {
    ZIP_FILENAME = "vend-cli"
  }
}

action "Publish Release" {
    needs = "Create Archive"
    uses = "./publish-release/"
    env = {
        ARTIFACT = "vend-cli.zip"
    }
}