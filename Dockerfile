# The image published as ghcr.io/hosttracker/ht-cli and, under the same tags,
# as docker.io/hosttracker/ht-cli.
#
# goreleaser builds it from the release binaries: the build context holds
# one binary per platform under <os>/<arch>/, which TARGETPLATFORM selects.
# Nothing is compiled here, so the arm64 image needs no emulation.
#
# distroless/static carries the CA bundle the API calls need, and nothing
# else: no shell, no package manager, and a non-root user by default.
FROM gcr.io/distroless/static:nonroot

ARG TARGETPLATFORM
COPY ${TARGETPLATFORM}/ht-cli /usr/bin/ht-cli

ENTRYPOINT ["/usr/bin/ht-cli"]
