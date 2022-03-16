Vendor:         Microsoft Corporation
Distribution:   Mariner
# Force out of source build
%undefine __cmake_in_source_build

%{!?jobs:%global jobs %(/usr/bin/getconf _NPROCESSORS_ONLN)}

# apt library somajor...
%global libsomajor 6.0
%global libprivsomajor 0.0

# Disable integration tests by default,
# as there is a bunch of failures on non-Debian systems currently.
# Additionally, these tests take a long time to run.
%bcond_without check_integration

Name:           apt
Version:        2.3.14
Release:        1%{?dist}
Summary:        Command-line package manager for Debian packages

License:        GPLv2+
URL:            https://tracker.debian.org/pkg/apt
Source0:        https://salsa.debian.org/apt-team/%{name}/-/archive/%{version}/%{name}-%{version}.tar.gz

BuildRequires:  gcc
BuildRequires:  gcc-c++
BuildRequires:  cmake >= 3.4
BuildRequires:  ninja-build

BuildRequires:  pkgconfig(gnutls) >= 3.4.6
BuildRequires:  pkgconfig(libgcrypt)
BuildRequires:  pkgconfig(liblzma)
BuildRequires:  pkgconfig(zlib)
BuildRequires:  pkgconfig(liblz4)
BuildRequires:  pkgconfig(libzstd)
BuildRequires:  pkgconfig(libseccomp)
BuildRequires:  pkgconfig(libudev)
BuildRequires:  pkgconfig(libxxhash)

# Package manager BRs
BuildRequires:  dpkg-perl
BuildRequires:  dpkg-devel

# These BRs lack pkgconfig() names
BuildRequires:  libdb-devel
BuildRequires:  gtest-devel
BuildRequires:  bzip2-devel

# Misc BRs
BuildRequires:  triehash
BuildRequires:  po4a >= 0.35
#BuildRequires:  docbook-style-xsl, docbook-dtds
BuildRequires:  docbook-style-xsl
BuildRequires:  gettext >= 0.19
BuildRequires:  doxygen
BuildRequires:  graphviz
#BuildRequires:  w3m
#BuildRequires:  %{_bindir}/xsltproc

%if %{with check_integration}
#BuildRequires:  coreutils, moreutils,
BuildRequires:  coreutils
#BuildRequires:  moreutils-parallel
BuildRequires:  fakeroot, lsof, sed
BuildRequires:  tar, wget
#BuildRequires:  tar, wget, stunnel
BuildRequires:  gnupg, gnupg2
#BuildRequires:  perl(File::FcntlLock)
BuildRequires:  perl(Digest::SHA)
#BuildRequires:  debhelper >= 9
# Unbreak running tests in non-interactive terminals
BuildRequires:  expect
%endif

# For ensuring the user is created
Requires(pre):  shadow-utils

# Apt is essentially broken without dpkg
Requires:       dpkg >= 1.17.14

# To ensure matching apt libs are installed
Requires:       %{name}-libs%{?_isa} = %{version}-%{release}

# apt-transport-curl-https is gone...
Provides:       %{name}-transport-https = %{version}-%{release}
Provides:       %{name}-transport-curl-https = %{version}-%{release}

%description
This package provides commandline tools for searching and
managing as well as querying information about packages
as a low-level access to all features of the libapt-pkg library.

These include:
  * apt-get for retrieval of packages and information about them
    from authenticated sources and for installation, upgrade and
    removal of packages together with their dependencies
  * apt-cache for querying available information about installed
    as well as installable packages
  * apt-cdrom to use removable media as a source for packages
  * apt-config as an interface to the configuration settings
  * apt-key as an interface to manage authentication keys

%package libs
Summary:        Runtime libraries for %{name}

%description libs
This package includes the libapt-pkg library.

libapt-pkg provides the common functionality for searching and
managing packages as well as information about packages.
Higher-level package managers can depend upon this library.

This includes:
  * retrieval of information about packages from multiple sources
  * retrieval of packages and all dependent packages
    needed to satisfy a request either through an internal
    solver or by interfacing with an external one
  * authenticating the sources and validating the retrieved data
  * installation and removal of packages in the system
  * providing different transports to retrieve data over cdrom, ftp,
    http, rsh as well as an interface to add more transports like
    debtorrent (apt-transport-debtorrent).

#%package doc
#Summary:        Documentation for APT
#BuildArch:      noarch

#%description doc
#This package contains the user guide and offline guide for various
#APT tools which are provided in a html and a text-only version.

%package devel
Summary:        Development files for APT's libraries
Provides:       libapt-pkg-devel%{?_isa} = %{version}-%{release}
Provides:       libapt-pkg-devel = %{version}-%{release}
Requires:       %{name}-libs%{?_isa} = %{version}-%{release}

%description devel
This package contains the header files and libraries for
developing with APT's libapt-pkg Debian package manipulation
library.

#%package apidoc
#Summary:        Documentation for developing against APT libraries
#Provides:       libapt-pkg-doc = %{version}-%{release}
#Obsoletes:      %{name}-devel-doc < 1.9.7-1
#Provides:       %{name}-devel-doc = %{version}-%{release}
#BuildArch:      noarch

#%description apidoc
#This package contains documentation for development of the APT
#Debian package manipulation program and its libraries.
#
#This includes the source code documentation generated by doxygen
#in html format.

%package utils
Summary:        Package management related utility programs
Requires:       %{name}%{?_isa} = %{version}-%{release}

%description utils
This package contains some less used commandline utilities related
to package management with APT.

  * apt-extracttemplates is used by debconf to prompt for configuration
    questions before installation.
  * apt-ftparchive is used to create Packages and other index files
    needed to publish an archive of Debian packages
  * apt-sortpkgs is a Packages/Sources file normalizer.

%prep
%autosetup -p1

# TODO: fix in dpkg.spec
mkdir -pv /etc/dpkg/origins/
cat <<EOF > /etc/dpkg/origins/debian
Vendor: debian
Vendor-URL: http://www.fedoraproject.org/
Bugs: https://bugzilla.redhat.com
EOF

%build
# This package fails its testsuite when LTO is enabled.  It is not yet clear if
# this is an LTO issue or an issue with the package itself
%define _lto_cflags %{nil}

mkdir build
pushd build
%cmake .. -GNinja -DWITH_TESTS=OFF -DWITH_DOC=OFF
cmake --build . --verbose
popd


%install
pushd build
DESTDIR=%{buildroot} cmake --install .
popd


%find_lang %{name}
%find_lang %{name}-utils
%find_lang libapt-pkg%{libsomajor}

cat libapt*.lang >> %{name}-libs.lang

mkdir -p %{buildroot}%{_localstatedir}/log/apt
touch %{buildroot}%{_localstatedir}/log/apt/{term,history}.log
mkdir -p %{buildroot}%{_sysconfdir}/apt/{apt.conf,preferences,sources.list,trusted.gpg}.d
install -pm 644 doc/examples/apt.conf %{buildroot}%{_sysconfdir}/apt/
touch %{buildroot}%{_sysconfdir}/apt/sources.list
mkdir -p %{buildroot}%{_sysconfdir}/logrotate.d
cat > %{buildroot}%{_sysconfdir}/logrotate.d/apt <<EOF
%{_localstatedir}/log/apt/term.log {
  rotate 12
  monthly
  compress
  missingok
  notifempty
}
%{_localstatedir}/log/apt/history.log {
  rotate 12
  monthly
  compress
  missingok
  notifempty
}
EOF


%check
%ctest
#%if %{with check_integration}
#unbuffer ./test/integration/run-tests -q %{?jobs:-j %{jobs}}
#%endif

# Create the _apt user+group for apt data
%pre
getent group _apt >/dev/null || groupadd -r _apt
getent passwd _apt >/dev/null || \
    useradd -r -g _apt -d %{_sharedstatedir}/apt -s /sbin/nologin \
    -c "APT account for owning persistent & cache data" _apt
exit 0

%ldconfig_scriptlets libs

%files -f %{name}.lang
%license COPYING*
%doc README.* AUTHORS
%{_bindir}/apt
%{_bindir}/apt-cache
%{_bindir}/apt-cdrom
%{_bindir}/apt-config
%{_bindir}/apt-get
%{_bindir}/apt-key
%{_bindir}/apt-mark
%dir %{_libexecdir}/apt
%{_libexecdir}/apt/apt-helper
%{_libexecdir}/apt/methods
%{_libexecdir}/dpkg/methods/apt
%attr(-,_apt,_apt) %{_sharedstatedir}/apt
%attr(-,_apt,_apt) %{_localstatedir}/cache/apt
%dir %attr(-,_apt,_apt) %{_localstatedir}/log/apt
%ghost %{_localstatedir}/log/apt/history.log
%ghost %{_localstatedir}/log/apt/term.log
%dir %attr(-,_apt,_apt) %{_sysconfdir}/apt/apt.conf.d
%dir %attr(-,_apt,_apt) %{_sysconfdir}/apt/preferences.d
%dir %attr(-,_apt,_apt) %{_sysconfdir}/apt/sources.list.d
%dir %attr(-,_apt,_apt) %{_sysconfdir}/apt/trusted.gpg.d
%config(noreplace) %attr(-,_apt,_apt) %{_sysconfdir}/apt/apt.conf
%ghost %{_sysconfdir}/apt/sources.list
%config(noreplace) %{_sysconfdir}/logrotate.d/apt
%{_datadir}/bash-completion/completions/apt
%{_mandir}/*/*/apt.*
%{_mandir}/*/*/apt-cache.*
%{_mandir}/*/*/apt-cdrom.*
%{_mandir}/*/*/apt-config.*
%{_mandir}/*/*/apt-get.*
%{_mandir}/*/*/apt-key.*
%{_mandir}/*/*/apt-mark.*
%{_mandir}/*/*/apt-patterns.*
%{_mandir}/*/*/apt-secure.*
%{_mandir}/*/*/apt-transport-http.*
%{_mandir}/*/*/apt-transport-https.*
%{_mandir}/*/*/apt-transport-mirror.*
%{_mandir}/*/*/apt_auth.*
%{_mandir}/*/*/apt_preferences.*
%{_mandir}/*/*/sources.list.*
%{_mandir}/*/apt.*
%{_mandir}/*/apt-cache.*
%{_mandir}/*/apt-cdrom.*
%{_mandir}/*/apt-config.*
%{_mandir}/*/apt-get.*
%{_mandir}/*/apt-key.*
%{_mandir}/*/apt-mark.*
%{_mandir}/*/apt-patterns.*
%{_mandir}/*/apt-secure.*
%{_mandir}/*/apt-transport-http.*
%{_mandir}/*/apt-transport-https.*
%{_mandir}/*/apt-transport-mirror.*
%{_mandir}/*/apt_auth.*
%{_mandir}/*/apt_preferences.*
%{_mandir}/*/sources.list.*
%doc %{_docdir}/%{name}

%files libs -f %{name}-libs.lang
%license COPYING*
%{_libdir}/libapt-pkg.so.%{libsomajor}{,.*}
%{_libdir}/libapt-private.so.%{libprivsomajor}{,.*}

#%files doc
#%doc %{_docdir}/%{name}-doc

#%files apidoc
#%doc %{_docdir}/libapt-pkg-doc

%files devel
%{_libdir}/*.so
%{_libdir}/pkgconfig/*.pc
%{_includedir}/*

%files utils -f %{name}-utils.lang
%{_bindir}/apt-extracttemplates
%{_bindir}/apt-ftparchive
%{_bindir}/apt-sortpkgs
%{_libexecdir}/apt/planners
%{_libexecdir}/apt/solvers
%{_mandir}/*/*/apt-extracttemplates.*
%{_mandir}/*/*/apt-ftparchive.*
%{_mandir}/*/*/apt-sortpkgs.*
%{_mandir}/*/apt-extracttemplates.*
%{_mandir}/*/apt-ftparchive.*
%{_mandir}/*/apt-sortpkgs.*
%{_datadir}/doc/apt-utils/examples/apt-ftparchive.conf
%{_datadir}/doc/apt-utils/examples/ftp-archive.conf
#%doc %{_docdir}/%{name}-utils

%changelog
* Sun Jan 16 2022 Sérgio Basto <sergio@serjux.com> - 2.3.14-1
- Update apt to 2.3.14 (#2037920)

* Sat Dec 18 2021 Sérgio Basto <sergio@serjux.com> - 2.3.13-1
- Update apt to 2.3.13 (#2024297)

* Thu Oct 21 2021 Fedora Release Monitoring <release-monitoring@fedoraproject.org> - 2.3.11-1
- Update to 2.3.11 (#2002944)

* Sat Aug 14 2021 Fedora Release Monitoring <release-monitoring@fedoraproject.org> - 2.3.8-1
- Update to 2.3.8 (#1993644)

* Thu Jul 29 2021 Fedora Release Monitoring <release-monitoring@fedoraproject.org> - 2.3.7-1
- Update to 2.3.7 (#1987763)

* Wed Jul 21 2021 Fedora Release Engineering <releng@fedoraproject.org> - 2.3.6-2
- Rebuilt for https://fedoraproject.org/wiki/Fedora_35_Mass_Rebuild

* Wed Jun 09 2021 Fedora Release Monitoring <release-monitoring@fedoraproject.org> - 2.3.6-1
- Update to 2.3.6 (#1969935)

* Mon May 17 2021 Fedora Release Monitoring <release-monitoring@fedoraproject.org> - 2.3.5-1
- Update to 2.3.5 (#1930430)

* Mon Feb 15 2021 Mosaab Alzoubi <moceap[At]hotmail[Dot]com> - 2.1.20-1
- Update to 2.1.20

* Tue Jan 26 2021 Fedora Release Engineering <releng@fedoraproject.org> - 2.1.18-2
- Rebuilt for https://fedoraproject.org/wiki/Fedora_34_Mass_Rebuild

* Wed Jan 13 2021 Fedora Release Monitoring <release-monitoring@fedoraproject.org> - 2.1.18-1
- Update to 2.1.18 (#1906457)

* Mon Nov 23 2020 Fedora Release Monitoring <release-monitoring@fedoraproject.org> - 2.1.12-1
- Update to 2.1.12 (#1900787)

* Wed Oct 21 2020 Fedora Release Monitoring <release-monitoring@fedoraproject.org> - 2.1.11-1
- Update to 2.1.11 (#1890077)

* Tue Aug 11 2020 Fedora Release Monitoring <release-monitoring@fedoraproject.org> - 2.1.10-1
- Update to 2.1.10 (#1868031)

* Mon Aug 10 2020 Fedora Release Monitoring <release-monitoring@fedoraproject.org> - 2.1.9-1
- Update to 2.1.9 (#1867591)

* Tue Aug 04 2020 Fedora Release Monitoring <release-monitoring@fedoraproject.org> - 2.1.8-1
- Update to 2.1.8 (#1865853)

* Sat Aug 01 2020 Fedora Release Engineering <releng@fedoraproject.org> - 2.1.7-4
- Second attempt - Rebuilt for
  https://fedoraproject.org/wiki/Fedora_33_Mass_Rebuild

* Mon Jul 27 2020 Jeff law <law@redhat.com.com> - 2.1.7-3
- Disable LTO for now

* Mon Jul 27 2020 Fedora Release Engineering <releng@fedoraproject.org> - 2.1.7-2
- Rebuilt for https://fedoraproject.org/wiki/Fedora_33_Mass_Rebuild

* Fri Jul 10 2020 Sérgio Basto <sergio@serjux.com> - 2.1.7-1
- Update apt to 2.1.7 (#1854759)

* Wed Jun 03 2020 Fedora Release Monitoring <release-monitoring@fedoraproject.org> - 2.1.6-1
- Update to 2.1.6 (#1831062)

* Tue May 26 2020 Fedora Release Monitoring <release-monitoring@fedoraproject.org> - 2.1.5-1
- Update to 2.1.5 (#1831062)

* Tue May 19 2020 Fedora Release Monitoring <release-monitoring@fedoraproject.org> - 2.1.4-1
- Update to 2.1.4 (#1831062)

* Thu Apr 09 2020 Fedora Release Monitoring <release-monitoring@fedoraproject.org> - 2.0.2-1
- Update to 2.0.2 (#1816610)

* Sat Mar 07 17:26:29 EST 2020 Neal Gompa <ngompa13@gmail.com> - 2.0.0-1
- Update to 2.0.0

* Tue Feb 18 2020 Fedora Release Monitoring <release-monitoring@fedoraproject.org> - 1.9.10-1
- Update to 1.9.10 (#1804170)

* Tue Jan 28 2020 Fedora Release Engineering <releng@fedoraproject.org> - 1.9.7-2
- Rebuilt for https://fedoraproject.org/wiki/Fedora_32_Mass_Rebuild

* Sat Jan 18 09:50:33 EST 2020 Neal Gompa <ngompa13@gmail.com> - 1.9.7-1
- Update to 1.9.7
- Rename apt-devel-doc to apt-apidoc to better reflect the content

* Mon Dec 16 22:10:42 EST 2019 Neal Gompa <ngompa13@gmail.com> - 1.9.4-1
- Switch from apt-rpm to apt from Debian and rebase to v1.9.4
  + This drops rpm support from apt
- Truncate changelog due to complete spec rewrite and replacement of apt implementation
