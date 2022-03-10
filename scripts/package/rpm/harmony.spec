{{VER=2.0.0}}
{{REL=0}}
# SPEC file overview:
# https://docs.fedoraproject.org/en-US/quick-docs/creating-rpm-packages/#con_rpm-spec-file-overview
# Fedora packaging guidelines:
# https://docs.fedoraproject.org/en-US/packaging-guidelines/

Name:		astra
Version:	{{ VER }}
Release:	{{ REL }}
Summary:	astra blockchain validator node program

License:	MIT
URL:		https://astra.one
Source0:	%{name}-%{version}.tar
BuildArch: x86_64
Packager: Leo Chen <leo@hamrony.one>
Requires(pre): shadow-utils
Requires: systemd-rpm-macros jq

%description
Astra is a sharded, fast finality, low fee, PoS public blockchain.
This package contains the validator node program for astra blockchain.

%global debug_package %{nil}

%prep
%setup -q

%build
exit 0

%check
./astra --version
exit 0

%pre
getent group astra >/dev/null || groupadd -r astra
getent passwd astra >/dev/null || \
   useradd -r -g astra -d /home/astra -m -s /sbin/nologin \
   -c "Astra validator node account" astra
mkdir -p /home/astra/.astra/blskeys
mkdir -p /home/astra/.config/rclone
chown -R astra.astra /home/astra
exit 0


%install
install -m 0755 -d ${RPM_BUILD_ROOT}/usr/sbin ${RPM_BUILD_ROOT}/etc/systemd/system ${RPM_BUILD_ROOT}/etc/sysctl.d ${RPM_BUILD_ROOT}/etc/astra
install -m 0755 -d ${RPM_BUILD_ROOT}/home/astra/.config/rclone
install -m 0755 astra ${RPM_BUILD_ROOT}/usr/sbin/
install -m 0755 astra-setup.sh ${RPM_BUILD_ROOT}/usr/sbin/
install -m 0755 astra-rclone.sh ${RPM_BUILD_ROOT}/usr/sbin/
install -m 0644 astra.service ${RPM_BUILD_ROOT}/etc/systemd/system/
install -m 0644 astra-sysctl.conf ${RPM_BUILD_ROOT}/etc/sysctl.d/99-astra.conf
install -m 0644 rclone.conf ${RPM_BUILD_ROOT}/etc/astra/
install -m 0644 astra.conf ${RPM_BUILD_ROOT}/etc/astra/
exit 0

%post
%systemd_user_post %{name}.service
%sysctl_apply %{name}-sysctl.conf
exit 0

%preun
%systemd_user_preun %{name}.service
exit 0

%postun
%systemd_postun_with_restart %{name}.service
exit 0

%files
/usr/sbin/astra
/usr/sbin/astra-setup.sh
/usr/sbin/astra-rclone.sh
/etc/sysctl.d/99-astra.conf
/etc/systemd/system/astra.service
/etc/astra/astra.conf
/etc/astra/rclone.conf
/home/astra/.config/rclone

%config(noreplace) /etc/astra/astra.conf
%config /etc/astra/rclone.conf
%config /etc/sysctl.d/99-astra.conf 
%config /etc/systemd/system/astra.service

%doc
%license



%changelog
* Wed Aug 26 2020 Leo Chen <leo at astra dot one> 2.3.5
   - get version from git tag
   - add %config macro to keep edited config files

* Tue Aug 18 2020 Leo Chen <leo at astra dot one> 2.3.4
   - init version of the astra node program

