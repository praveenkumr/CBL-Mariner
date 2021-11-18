Summary:        Nmap Network Mapper
Name:           nmap
Version:        7.90
Release:        3%{?dist}
License:        Nmap
Vendor:         Microsoft Corporation
Distribution:   Mariner
Group:          Applications/System
URL:            https://nmap.org/
Source0:        https://nmap.org/dist/%{name}-%{version}.tar.bz2
## https://github.com/nmap/nmap/commit/f6b40614e4a8131394792d590965f8af3c635323.patch
Patch0:         nmap-unix_crash.patch
BuildRequires:  binutils
BuildRequires:  gcc
BuildRequires:  kernel-headers
BuildRequires:  libpcap-devel
BuildRequires:  libssh2-devel
BuildRequires:  lua-devel
BuildRequires:  make
BuildRequires:  openssl-devel
BuildRequires:  zlib-devel

Requires:       lua

%description
Nmap ("Network Mapper") is a free and open source utility for network discovery and security auditing.

%package ncat
Summary:        Nmap replacement for ncat
Requires:       lua
Provides:       nc

%description ncat
Nmap implementation of the ncat tool

%prep
%autosetup -p1

# Remove bundled copies of several libraries. Leave pcre since we dont have a static version.
rm -rf libpcap macosx mswin32 libssh2 libz

%build
%configure
%make_build

%install
%make_install

rm -rf %{buildroot}%{_datadir}/ncat
ln -s ncat %{buildroot}%{_bindir}/nc

%files
%license LICENSE
%exclude %{_mandir}
%{_bindir}/nmap
%{_bindir}/nping
%{_datadir}/nmap

%files ncat
%license LICENSE
%{_bindir}/ncat
%{_bindir}/nc

%changelog
* Thu Nov 18 2021 Pawel Winogrodzki <pawelwi@microsoft.com> - 7.90-3
- Added missing dependencies on 'lua'.
- Updating release to recompile with 'lua' 5.4.3.

* Mon May 17 2021 Suresh Babu Chalamalasetty <schalam@microsoft.com> 7.90-2
- nmap-unix_crash.patch fix for crash with unix sockets.

* Tue Feb 02 2021 Henry Beberman <henry.beberman@microsoft.com> 7.90-1
- Add nmap spec
- License verified
- Original version for CBL-Mariner
