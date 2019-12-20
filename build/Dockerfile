# Step one: build compliance-operator
FROM registry.access.redhat.com/ubi8/go-toolset as builder

WORKDIR /go/src/github.com/JAORMX/selinux-operator

ENV GOFLAGS=-mod=vendor

COPY . .
RUN make

# Step two: containerize compliance-operator
FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

ENV OPERATOR=/usr/local/bin/selinux-operator \
    USER_UID=1001 \
    USER_NAME=selinux-operator

# install operator binary
COPY build/_output/bin/selinux-operator ${OPERATOR}
COPY --from=builder /go/src/github.com/JAORMX/selinux-operator/build/_output/bin/selinux-operator ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
