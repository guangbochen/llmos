ARG ELEMENTAL_TOOLKIT
ARG LLMOS_CLI_REPO
ARG VERSION
ARG GOLANG_ARCH
FROM ${ELEMENTAL_TOOLKIT} AS TOOLKIT

FROM ${LLMOS_CLI_REPO}:${VERSION}-${GOLANG_ARCH} AS LLMOS_CLI

FROM ollama/ollama:0.1.29 as ollama

# OS base image of our flavor choice
FROM registry.opensuse.org/opensuse/leap:15.5 as OS
ARG REPO
ARG VERSION
ARG ARCH
ARG GOLANG_ARCH
ENV VERSION=${VERSION}

# install kernel, systemd, dracut, grub2 and other required tools
RUN ARCH=$(uname -m); \
    [[ "${ARCH}" == "aarch64" ]] && ARCH="arm64"; \
    zypper --non-interactive install --no-recommends -- \
      kernel-default \
      device-mapper \
      dracut \
      grub2 \
      grub2-${ARCH}-efi \
      shim \
      haveged \
      systemd \
      NetworkManager \
      openssh-server \
      openssh-clients \
      timezone \
      parted \
      e2fsprogs \
      dosfstools \
      mtools \
      xorriso \
      findutils \
      gptfdisk \
      rsync \
      squashfs \
      lvm2 \
      tar \
      gzip \
      vim \
      which \
      less \
      sudo \
      curl \
      iproute2 \
      podman \
      sed \
      btrfsprogs \
      btrfsmaintenance \
      wget2 \
      snapper && \
    zypper clean --all

# Add system files first
COPY iso/framework/files /

# Add LLMOS installer cli
COPY --from=LLMOS_CLI /llmos /usr/bin/llmos
RUN chmod +x /usr/bin/llmos

# Add the elemental cli
COPY --from=TOOLKIT /usr/bin/elemental /usr/bin/elemental

# Add the ollama bin
COPY --from=ollama /usr/bin/ollama /usr/bin/ollama
COPY build-models /var/lib/llmos/ollama

# Copy the ollama model

# Enable essential services
RUN systemctl enable NetworkManager.service

# Enable /tmp to be on tmpfs
RUN cp /usr/share/systemd/tmp.mount /etc/systemd/system

# Generate initrd with required elemental services
RUN elemental init --debug --force

# Update os-release file with some metadata
RUN echo IMAGE_REPO=\"${REPO}\"         >> /etc/os-release && \
    echo IMAGE_TAG=\"${VERSION}\"           >> /etc/os-release && \
    echo IMAGE=\"${REPO}:${VERSION}\" >> /etc/os-release && \
    echo TIMESTAMP="`date +'%Y%m%d%H%M%S'`" >> /etc/os-release && \
    echo GRUB_ENTRY_NAME=\"LLMOS\" >> /etc/os-release

RUN rm /etc/machine-id || true
RUN rm /var/lib/dbus/machine-id || true
RUN rm /etc/hostname || true

# Good for validation after the build
CMD /bin/bash
