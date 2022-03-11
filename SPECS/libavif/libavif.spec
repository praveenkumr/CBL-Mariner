Vendor:         Microsoft Corporation
Distribution:   Mariner
# Build with aom
%bcond_without aom
# Build SVT-AV1
%ifarch x86_64
%bcond_without svt
%endif

Name:           libavif
Version:        0.9.0
Release:        2%{?dist}
Summary:        Library for encoding and decoding .avif files

License:        BSD
URL:            https://github.com/AOMediaCodec/libavif
Source0:        %{url}/archive/v%{version}/%{name}-%{version}.tar.gz

BuildRequires:  cmake
BuildRequires:  gcc-c++
BuildRequires:  nasm
%if %{with aom}
BuildRequires:  pkgconfig(aom)
%endif
BuildRequires:  pkgconfig(dav1d)
BuildRequires:  pkgconfig(libjpeg)
BuildRequires:  pkgconfig(libpng)
BuildRequires:  pkgconfig(rav1e)
%{?with_svt:BuildRequires:  pkgconfig(SvtAv1Enc)}
BuildRequires:  pkgconfig(zlib)

%description
This library aims to be a friendly, portable C implementation of the AV1 Image
File Format, as described here:

https://aomediacodec.github.io/av1-avif/

%package devel
Summary:        Development files for libavif
Requires:       %{name}%{?_isa} = %{version}-%{release}

%description devel
This package holds the development files for libavif.

%package tools
Summary:        Tools to encode and decode AVIF files

%description tools
This library aims to be a friendly, portable C implementation of the AV1 Image
File Format, as described here:

https://aomediacodec.github.io/av1-avif/

This package holds the commandline tools to encode and decode AVIF files.

%package     -n avif-pixbuf-loader
Summary:        AVIF image loader for GTK+ applications
BuildRequires:  pkgconfig(gdk-pixbuf-2.0)
Requires:       gdk-pixbuf2

%description -n avif-pixbuf-loader
Avif-pixbuf-loader contains a plugin to load AVIF images in GTK+ applications.

%prep
%autosetup -p1

%build
%cmake \
    -DCMAKE_BUILD_TYPE=RelWithDebInfo \
    %{?with_aom:-DAVIF_CODEC_AOM=1} \
    -DAVIF_CODEC_DAV1D=1 \
    -DAVIF_CODEC_RAV1E=1 \
    %{?with_svt:-DAVIF_CODEC_SVT=1} \
    -DAVIF_BUILD_APPS=1 \
    -DAVIF_BUILD_GDK_PIXBUF=1
%cmake_build

%install
%cmake_install

%files
%license LICENSE
# Do not glob the soname
%{_libdir}/libavif.so.10*

%files devel
%{_libdir}/libavif.so
%{_includedir}/avif/
%{_libdir}/cmake/libavif/
%{_libdir}/pkgconfig/libavif.pc

%files tools
%doc CHANGELOG.md README.md
%{_bindir}/avifdec
%{_bindir}/avifenc

%files -n avif-pixbuf-loader
%{_libdir}/gdk-pixbuf-2.0/*/loaders/libpixbufloader-avif.so

%changelog
* Sun Jun 13 13:40:21 CEST 2021 Robert-André Mauchin <zebob.m@gmail.com> - 0.9.0-2
- Rebuilt for aom v3.1.1

* Mon Mar 15 2021 Andreas Schneider <asn@redhat.com> - 0.9.0-1
- Update to version 0.9.0

* Sat Feb 20 2021 Andreas Schneider <asn@redhat.com> - 0.8.4-4
- Build release with debug info

* Tue Jan 26 2021 Fedora Release Engineering <releng@fedoraproject.org> - 0.8.4-3
- Rebuilt for https://fedoraproject.org/wiki/Fedora_34_Mass_Rebuild

* Mon Dec 14 2020 Robert-André Mauchin <zebob.m@gmail.com> - 0.8.4-2
- Rebuild for dav1d SONAME bump

* Wed Dec 09 05:52:07 CET 2020 Robert-André Mauchin <zebob.m@gmail.com> - 0.8.4-1
- Update to version 0.8.4

* Mon Oct 19 2020 Andreas Schneider <asn@redhat.com> - 0.8.2-1
- Update to version 0.8.2
  https://github.com/AOMediaCodec/libavif/blob/master/CHANGELOG.md

* Thu Aug 06 22:14:02 CEST 2020 Robert-André Mauchin <zebob.m@gmail.com> - 0.8.1-1
- Update to 0.8.1

* Wed Aug 05 21:17:23 CEST 2020 Robert-André Mauchin <zebob.m@gmail.com> - 0.8.0-1
- Update to 0.8.0

* Sat Aug 01 2020 Fedora Release Engineering <releng@fedoraproject.org> - 0.7.3-3
- Second attempt - Rebuilt for
  https://fedoraproject.org/wiki/Fedora_33_Mass_Rebuild

* Tue Jul 28 2020 Fedora Release Engineering <releng@fedoraproject.org> - 0.7.3-2
- Rebuilt for https://fedoraproject.org/wiki/Fedora_33_Mass_Rebuild

* Fri May 22 2020 Igor Raits <ignatenkobrain@fedoraproject.org> - 0.7.3-1
- Update to 0.7.3

* Wed Apr 29 2020 Andreas Schneider <asn@redhat.com> - 0.7.2-1
- Update to version 0.7.2
  * https://github.com/AOMediaCodec/libavif/blob/master/CHANGELOG.md

* Wed Apr 29 2020 Andreas Schneider <asn@redhat.com> - 0.7.1-1
- Update to version 0.7.1

* Wed Mar 04 2020 Andreas Schneider <asn@redhat.com> - 0.5.7-1
- Update to version 0.5.7

* Wed Mar 04 2020 Andreas Schneider <asn@redhat.com> - 0.5.3-2
- Fix License

* Sun Feb 16 2020 Andreas Schneider <asn@redhat.com> - 0.5.3-1
- Initial version
