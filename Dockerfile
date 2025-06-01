# syntax=docker/dockerfile:1

# -✂- this stage is used to compile the application -------------------------------------------------------------------
FROM docker.io/library/golang:1.24-alpine AS compile

# can be passed with any prefix (like `v1.2.3@FOO`), e.g.: `docker build --build-arg "APP_VERSION=v1.2.3@FOO" .`
ARG APP_VERSION="undefined@docker"

# copy the source code
COPY . /src
WORKDIR /src

RUN set -x \
    # build the app itself
    && go generate -skip readme ./... \
    && CGO_ENABLED=0 go build \
      -trimpath \
      -ldflags "-s -w -X gh.tarampamp.am/describe-commit/internal/version.version=${APP_VERSION}" \
      -o ./describe-commit \
      ./cmd/describe-commit/ \
    && ./describe-commit --help

# -✂- and this is the final stage -------------------------------------------------------------------------------------
FROM docker.io/library/alpine:3.22 AS runtime

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

# the root user is used to avoid permission issues with mounted volumes. however, if you prefer, you can use a
# non-root user with the following command: `docker run -u "$(id -u):$(id -g)" -v "$(pwd):/src:ro" -w "/src" ...`
USER 0:0

ENTRYPOINT ["/bin/describe-commit"]
