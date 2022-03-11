Vendor:         Microsoft Corporation
Distribution:   Mariner
%{?!python3_pkgversion:%global python3_pkgversion 3}
%global sover 25

Name:           openexr
Version:        2.5.5
Release:        1%{?dist}
Summary:        Provides the specification and reference implementation of the EXR file format

License:        BSD
URL:            https://www.openexr.com/
Source0:        https://github.com/AcademySoftwareFoundation/%{name}/archive/v%{version}/%{name}-%{version}.tar.gz

Patch0:         openexr-gcc11.patch

BuildRequires:  cmake gcc gcc-c++
BuildRequires:  boost-devel
BuildRequires:  python%{python3_pkgversion}-devel
BuildRequires:  zlib-devel

Obsoletes:      OpenEXR < 2.5.3
Provides:       OpenEXR = %{version}-%{release}

%description
OpenEXR is an open-source high-dynamic-range floating-point image file format
for high-quality image processing and storage. This document presents a brief
overview of OpenEXR and explains concepts that are specific to this format.

This package containes the binaries for OpenEXR.


%package libs
Summary:        OpenEXR Libraries
Provides:       OpenEXR-libs = %{version}-%{release}
Obsoletes:      OpenEXR-libs < 2.5.3

%description libs
OpenEXR is an open-source high-dynamic-range floating-point image file format
for high-quality image processing and storage. This document presents a brief
overview of OpenEXR and explains concepts that are specific to this format.

OpenEXR Features:

* High dynamic range and color precision.  Support for 16-bit floating-point,
* 32-bit floating-point, and 32-bit integer pixels.
* Multiple image compression algorithms, both lossless and lossy. Some of
  the included codecs can achieve 2:1 lossless compression ratios on images
  with film grain.  The lossy codecs have been tuned for visual quality and
  decoding performance.
* Extensibility. New compression codecs and image types can easily be added
  by extending the C++ classes included in the OpenEXR software distribution.
  New image attributes (strings, vectors, integers, etc.) can be added to
  OpenEXR image headers without affecting backward compatibility with existing
  OpenEXR applications.
* Support for stereoscopic image workflows and a generalization
  to multi-views.
* Flexible support for deep data: pixels can store a variable-length list
  of samples and, thus, it is possible to store multiple values at different
  depths for each pixel. Hard surfaces and volumetric data representations are
  accommodated.
* Multipart: ability to encode separate, but related, images in one file.
  This allows for access to individual parts without the need to read other
  parts in the file.
* Versioning: OpenEXR source allows for user configurable C++
  namespaces to provide protection when using multiple versions of the library
  in the same process space.

The IlmBase Library:

Also a part of OpenEXR, the IlmBase library is a basic, light-weight, and
efficient representation of 2D and 3D vectors and matrices and other simple but
useful mathematical objects, functions, and data types common in computer
graphics applications, including the “half” 16-bit floating-point type.


%package -n python%{python3_pkgversion}-openexr
Summary:        Python 3 binding for OpenEXR
%{?python_provide:%python_provide python3-openexr}

%description -n python%{python3_pkgversion}-openexr
%{summary}.


%package devel
Requires:       %{name}-libs%{?_isa} = %{version}-%{release}

Provides:       OpenEXR-devel = %{version}-%{release}
Provides:       OpenEXR-devel%{?_isa} = %{version}-%{release}
Obsoletes:      OpenEXR-devel < 2.5.3

Provides:       ilmbase-devel = %{version}-%{release}
Provides:       ilmbase-devel%{?_isa} = %{version}-%{release}
Obsoletes:      ilmbase-devel < 2.5.3

Summary:        Development files for %{name}

%description devel
%{summary}.


%prep
%autosetup -p1


%build
%cmake
%cmake_build


%install
%cmake_install


%check
# Test 4 currently fails on aarch64 and sometimes times out on armv7hl
# https://github.com/AcademySoftwareFoundation/openexr/issues/876
%ifnarch armv7hl aarch64 s390x
%ctest
%endif


%files
%{_bindir}/*

%files libs
%doc CHANGES.md CONTRIBUTING.md GOVERNANCE.md SECURITY.md CODE_OF_CONDUCT.md CONTRIBUTORS.md README.md
%license LICENSE.md
%{_libdir}/*.so.%{sover}*
%exclude %{_libdir}/libPy*

%files -n python%{python3_pkgversion}-openexr
# Is it OK to dump the libraries in site-packages?
%{_libdir}/libPy*.so.%{sover}*
%{python3_sitearch}/*.so

%files devel
%{_docdir}/OpenEXR/
%{_includedir}/OpenEXR/
%{_libdir}/*.so
%{_libdir}/cmake/IlmBase/
%{_libdir}/pkgconfig/IlmBase.pc
%{_libdir}/cmake/PyIlmBase/
%{_libdir}/pkgconfig/PyIlmBase.pc
%{_libdir}/cmake/OpenEXR/
%{_libdir}/pkgconfig/OpenEXR.pc


%changelog
* Mon Mar 15 2021 Richard Shaw <hobbes1069@gmail.com> - 2.5.5-1
- Update to 2.5.5.

* Tue Jan 26 2021 Fedora Release Engineering <releng@fedoraproject.org> - 2.5.4-4
- Rebuilt for https://fedoraproject.org/wiki/Fedora_34_Mass_Rebuild

* Fri Jan 22 2021 Jonathan Wakely <jwakely@redhat.com> - 2.5.4-3
- Rebuilt for Boost 1.75

* Mon Jan 18 2021 Richard Shaw <hobbes1069@gmail.com> - 2.5.4-2
- Fix Provides/Obsoletes of OpenEXR package.

* Wed Jan 06 2021 Richard Shaw <hobbes1069@gmail.com> - 2.5.4-1
- Update to 2.5.4.

* Wed Dec  9 2020 Richard Shaw <hobbes1069@gmail.com> - 2.5.3-1
- Repackaged due to massive changes in build system and inclusion of IlmBase.
