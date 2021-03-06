package main

import (
	"fmt"
	"path"

	"github.com/goreleaser/goreleaser/config"
)

// processEquinoxio create a fake goreleaser config for equinox.io
// and use a similar template.
func processRaw(repo string, exe string, nametpl string) (string, error) {
	if repo == "" {
		return "", fmt.Errorf("must have GitHub owner/repo")
	}
	if exe == "" {
		exe = path.Base(repo)
	}
	if nametpl == "" {
		nametpl = "{{ .Binary }}_v{{ .Version }}_{{ .Os }}_{{ .Arch }}"
	}

	// translate golang template to shell string
	name, err := makeName(nametpl)
	if err != nil {
		return "", err
	}

	project := config.Project{}
	project.Release.GitHub.Owner = path.Dir(repo)
	project.Release.GitHub.Name = path.Base(repo)
	project.Builds = []config.Build{
		{Binary: exe},
	}
	project.Archive.NameTemplate = name
	return makeShell(shellRaw, &project)
}

var shellRaw = `#!/bin/sh
set -e
#  Code generated by godownloader. DO NOT EDIT.
#

usage() {
  this=$1
  cat <<EOF

$this: download binaries for {{ $.Release.GitHub.Owner }}/{{ $.Release.GitHub.Name }}

Usage: $this [-b bindir] [version]
  -b set BINDIR or install directory.  Defaults to ./bin
  [version] is a version number from
  https://github.com/{{ $.Release.GitHub.Owner }}/{{ $.Release.GitHub.Name }}/releases
  if absent, defaults to latest.  Consider setting GITHUB_TOKEN to avoid
  triggering GitHub rate limits.  See the following for more details:
  https://help.github.com/articles/creating-a-personal-access-token-for-the-command-line/

Generated by godownloader
 https://github.com/goreleaser/godownloader

EOF
  exit 2
}
parse_args() {
  #BINDIR is ./bin unless set be ENV
  # over-ridden by flag below

  BINDIR=${BINDIR:-./bin}
  while getopts "b:h?" arg; do
    case "$arg" in
      b) BINDIR="$OPTARG" ;;
      h) usage "$0" ;;
      \?) usage "$0" ;;
    esac
  done
  shift $((OPTIND - 1))
  VERSION=$1
}
adjust_version() {
  if [ -z "${VERSION}" ]; then
    echo "$PREFIX: checking GitHub for latest version"
    VERSION=$(github_last_release "$OWNER/$REPO")
  fi
  # if version starts with 'v', remove it
  VERSION=${VERSION#v}
}
adjust_binary() {
  if [ "$OS" = "windows" ]; then
    NAME="${NAME}.exe"
    BINARY="${BINARY}.exe"
  fi
}
# wrap all destructive operations into a function
# to prevent curl|bash network truncation and disaster
execute() {
  TMPDIR=$(mktmpdir)
  echo "$PREFIX: downloading from ${TARBALL_URL}"
  http_download "${TMPDIR}/${NAME}" "$TARBALL_URL"
  install -d "${BINDIR}"
  install "${TMPDIR}/${NAME}" "${BINDIR}/${BINARY}"
  echo "$PREFIX: installed ${BINDIR}/${BINARY}"
}
` + shellfn + `
OWNER={{ .Release.GitHub.Owner }}
REPO={{ .Release.GitHub.Name }}
BINARY={{ (index .Builds 0).Binary }}
BINDIR=${BINDIR:-./bin}
PREFIX="$OWNER/$REPO"
OS=$(uname_os)
ARCH=$(uname_arch)

# make sure we are on a platform that makes sense
uname_os_check "$OS"
uname_arch_check "$ARCH"

# parse_args, show usage and exit if necessary
parse_args "$@"

# get or adjust version
adjust_version

{{ .Archive.NameTemplate }}

# adjust binary name based on OS
adjust_binary

# compute URL to download
TARBALL_URL=https://github.com/${OWNER}/${REPO}/releases/download/v${VERSION}/${NAME}

# do it
execute
`
