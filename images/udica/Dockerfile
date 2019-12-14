FROM fedora:latest

# Can't use this yet as it cannot find the udica package
#FROM registry.access.redhat.com/ubi8/ubi:latest

USER root

# Update image
RUN dnf update --disableplugin=subscription-manager -y && \
    rm -rf /var/cache/yum

# Install dependencies
RUN dnf install --disableplugin=subscription-manager -y \
            python3 \
            python3-setools \
            systemd-devel \
            policycoreutils \
            policycoreutils-python-utils \
            udica \
    && rm -rf /var/cache/yum

ENTRYPOINT ["/usr/bin/udica"]
