#!/bin/sh
set -e
set -o noglob

# --- helper functions for logs ---
info()
{
	echo '[INFO] ' "$@"
}
warn()
{
	echo '[WARN] ' "$@" >&2
}
fatal()
{
	echo '[ERROR] ' "$@" >&2
	exit 1
}

# --- fatal if no systemd ---
verify_system() {
	if [ -x /bin/systemctl ] || type systemctl > /dev/null 2>&1; then
		HAS_SYSTEMD=true
		return
	fi
	if [ -x /sbin/openrc-run ]; then
		HAS_OPENRC=true
		return
	fi

	fatal 'Process supervisor not found, please uninstall it manually by referring to the doc: https://docs-llmos.1block.ai'
}

# --- use sudo if we are not already root ---
check_sudo() {
    if [ $(id -u) -ne 0 ]; then
		if command -v sudo >/dev/null 2>&1; then
			info "running as non-root, will use sudo for installation."
			SUDO=sudo
	  	else
			fatal "This script must be run as root. Please use sudo or run as root."
		fi
    else
		SUDO=
    fi
}

# --- define needed environment variables ---
setup_env() {
    SYSTEM_NAME=llmos

    # --- use systemd directory if defined or create default ---
    if [ -n "${INSTALL_LLMOS_SYSTEMD_DIR}" ]; then
        SYSTEMD_DIR="${INSTALL_LLMOS_SYSTEMD_DIR}"
    else
        SYSTEMD_DIR=/etc/systemd/system
    fi

    # --- set related files from system name ---
    SERVICE_LLMOS=${SYSTEM_NAME}.service

    # --- use service or environment location depending on systemd/openrc ---
	if [ "${HAS_SYSTEMD}" = true ]; then
		FILE_LLMOS_SERVICE=${SYSTEMD_DIR}/${SERVICE_LLMOS}
		FILE_LLMOS_ENV=${SYSTEMD_DIR}/${SERVICE_LLMOS}.env
    elif [ "${HAS_OPENRC}" = true ]; then
		$SUDO mkdir -p /etc/llmos
		FILE_LLMOS_SERVICE=/etc/init.d/${SYSTEM_NAME}
		FILE_LLMOS_ENV=/etc/llmos/${SYSTEM_NAME}.env
    fi

    # check k3s or rke2 exist

    KUBE_UNINSTALL=
    if [ -f /etc/systemd/system/k3s.service ]; then
        KUBE_UNINSTALL=k3s-uninstall.sh
    elif [ -f /etc/systemd/system/k3s-agent.service ]; then
        KUBE_UNINSTALL=k3s-agent-uninstall.sh
    elif [ -f /etc/systemd/system/rke2.service ]; then
        KUBE_UNINSTALL=rke2-uninstall.sh
    elif [ -f /etc/systemd/system/rke2-agent.service ]; then
        KUBE_UNINSTALL=rke2-agent-uninstall.sh
    else
    	warn "Kubernetes runtime not found, skipping uninstall k8s runtime."
    fi
}

uninstall_llmos() {
	if [ -n "${KUBE_UNINSTALL}" ]; then
		info "Uninstalling k8s runtime by ${KUBE_UNINSTALL}"
		$SUDO ${KUBE_UNINSTALL}

		$SUDO rm -rf /var/lib/rancher /etc/rancher
	fi

	if [ -f "${FILE_LLMOS_SERVICE}" ]; then
		info "Uninstalling ${SYSTEM_NAME} service"
		if [ "${HAS_SYSTEMD}" = true ]; then
			$SUDO systemctl stop ${SERVICE_LLMOS}
			$SUDO systemctl disable ${SERVICE_LLMOS}
			$SUDO rm -f "${FILE_LLMOS_SERVICE}"
			$SUDO rm -f "${FILE_LLMOS_ENV}"
		elif [ "${HAS_OPENRC}" = true ]; then
			$SUDO rm -f "${FILE_LLMOS_SERVICE}"
		fi

		$SUDO rm -rf /val/lib/llmos /var/lib/rook/*llmos /etc/llmos
	fi


	info "Uninstalled ${SYSTEM_NAME} service"
}

# --- run the un-install process --
{
    verify_system
	setup_env
    check_sudo
    uninstall_llmos
}