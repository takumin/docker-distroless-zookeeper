#
# Build Arguments
#

ARG ZOOKEEPER_VERSION="3.8.0"
ARG ZOOKEEPER_GPG_KEY="BBE7232D7991050B54C8EA0ADC08637CA615D22C"

#
# Build Container
#

FROM docker.io/library/golang:alpine as entrypoint

WORKDIR /go/src/github.com/takumin/docker-distroless-zookeeper
COPY . .
RUN CGO_ENABLED=0 go build -o /docker-entrypoint

FROM docker.io/library/debian:11-slim as artifact

# hadolint ignore=DL3008
RUN export DEBIAN_FRONTEND=noninteractive; \
    apt-get update -yqq; \
    apt-get install --no-install-recommends -yqq \
        openjdk-17-jre-headless \
        ca-certificates \
        dirmngr \
        gnupg \
        wget; \
    apt-get clean; \
    rm -rf /var/lib/apt/lists/*

ARG ZOOKEEPER_VERSION
ENV ZOOKEEPER_VERSION ${ZOOKEEPER_VERSION}

ARG ZOOKEEPER_GPG_KEY
ENV ZOOKEEPER_GPG_KEY ${ZOOKEEPER_GPG_KEY}

# hadolint ignore=SC2155
RUN export GNUPGHOME="$(mktemp -d)"; \
    export DISTRO_NAME="apache-zookeeper-${ZOOKEEPER_VERSION}-bin"; \
    wget -q "https://dlcdn.apache.org/zookeeper/zookeeper-${ZOOKEEPER_VERSION}/${DISTRO_NAME}.tar.gz"; \
    wget -q "https://dlcdn.apache.org/zookeeper/zookeeper-${ZOOKEEPER_VERSION}/${DISTRO_NAME}.tar.gz.asc"; \
    gpg -q --keyserver "hkps://keys.openpgp.org" --recv-key "${ZOOKEEPER_GPG_KEY}" || \
    gpg -q --keyserver "hkps://keyserver.ubuntu.com" --recv-key "${ZOOKEEPER_GPG_KEY}" || \
    gpg -q --keyserver "hkps://keyserver.pgp.com" --recv-key "${ZOOKEEPER_GPG_KEY}" || \
    gpg -q --keyserver "hkps://pgp.mit.edu" --recv-key "${ZOOKEEPER_GPG_KEY}" || \
    gpg -q --batch --verify "${DISTRO_NAME}.tar.gz.asc" "${DISTRO_NAME}.tar.gz"; \
    mkdir "/zookeeper"; \
    tar -xf "${DISTRO_NAME}.tar.gz" -C "/zookeeper" --strip-components 1; \
    rm -fr "${GNUPGHOME}" "${DISTRO_NAME}.tar.gz" "${DISTRO_NAME}.tar.gz.asc"; \
    mv "/zookeeper/conf/zoo_sample.cfg" "/zookeeper/conf/zoo.cfg"

WORKDIR "/zookeeper"

#
# Service Container
#

FROM gcr.io/distroless/java17-debian11:nonroot as service

COPY --chown=nonroot:nonroot --from=entrypoint /docker-entrypoint /docker-entrypoint
COPY --chown=nonroot:nonroot --from=artifact /zookeeper/conf /zookeeper/conf
COPY --chown=nonroot:nonroot --from=artifact /zookeeper/lib /zookeeper/lib

ARG ZOOKEEPER_VERSION
ENV ZOOKEEPER_VERSION ${ZOOKEEPER_VERSION}

USER nonroot
WORKDIR /zookeeper
ENTRYPOINT ["/docker-entrypoint"]
EXPOSE 2181 2888 3888 8080
