# syntax=docker/dockerfile:1

# -✂- this stage is used to develop and build the application locally -------------------------------------------------
FROM docker.io/library/golang:1.23-alpine AS develop

ENV \
  # use the /var/tmp/go as the GOPATH to reuse the modules cache
  GOPATH="/var/tmp/go" \
  # set path to the Go cache (think about this as a "object files cache")
  GOCACHE="/var/tmp/go/cache"

# install development tools and dependencies
RUN set -x \
    && apk add --no-cache git \
    # renovate: source=github-releases name=golangci/golangci-lint
    && GOLANGCI_LINT_VERSION="1.64.4" \
    && wget -O- -nv "https://cdn.jsdelivr.net/gh/golangci/golangci-lint@v${GOLANGCI_LINT_VERSION}/install.sh" \
      | sh -s -- -b /bin "v${GOLANGCI_LINT_VERSION}"

WORKDIR /src

RUN \
    --mount=type=bind,source=go.mod,target=/src/go.mod \
    --mount=type=bind,source=go.sum,target=/src/go.sum \
    set -x \
    # burn the Go modules cache
    && go mod download -x \
    # allow anyone to read/write the Go cache
    && find /var/tmp/go -type d -exec chmod 0777 {} + \
    && find /var/tmp/go -type f -exec chmod 0666 {} +

# -✂- this stage is used to compile the application -------------------------------------------------------------------
FROM develop AS compile

# can be passed with any prefix (like `v1.2.3@FOO`), e.g.: `docker build --build-arg "APP_VERSION=v1.2.3@FOO" .`
ARG APP_VERSION="undefined@docker"

# copy the source code
COPY . /src

RUN set -x \
    # build the app itself
    && go generate -skip readme ./... \
    && CGO_ENABLED=0 go build \
      -trimpath \
      -ldflags "-s -w -X gh.tarampamp.am/describe-commit/internal/version.version=${APP_VERSION}" \
      -o ./describe-commit \
      ./cmd/describe-commit/ \
    && ./describe-commit --version

# -✂- and this is the final stage -------------------------------------------------------------------------------------
FROM docker.io/library/alpine:3.21 AS runtime

ARG APP_VERSION="undefined@docker"

LABEL \
    # Docs: <https://github.com/opencontainers/image-spec/blob/master/annotations.md>
    org.opencontainers.image.title="describe-commit" \
    org.opencontainers.image.description="Generate a commit message based on the git history using AI" \
    org.opencontainers.image.url="https://github.com/tarampampam/describe-commit" \
    org.opencontainers.image.source="https://github.com/tarampampam/describe-commit" \
    org.opencontainers.image.vendor="tarampampam" \
    org.opencontainers.version="$APP_VERSION" \
    org.opencontainers.image.licenses="MIT"

# install git
RUN apk add --no-cache git

# import compiled application
COPY --from=compile /src/describe-commit /bin/describe-commit

ENTRYPOINT ["/bin/describe-commit"]
