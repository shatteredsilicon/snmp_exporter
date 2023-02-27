%define debug_package   %{nil}
%define _GOPATH         %{_builddir}/go

%global provider                github
%global provider_tld            com
%global project                 prometheus
%global repo                    snmp_exporter
%global import_path             %{provider}.%{provider_tld}/%{project}/%{repo}

Name:           %{repo}
Summary:        Prometheus SNMP Exporter
Version:        %{_version}
Release:        1%{?dist}
License:        Apache-2.0
Source0:        %{name}-%{_version}.tar.gz

BuildRequires:  golang net-snmp-devel

%description
Prometheus SNMP Exporter

%prep
%setup -q -n %{name}

%build
mkdir -p %{_GOPATH}/bin
export GOPATH=%{_GOPATH}

go build -ldflags="-s -w" -o %{_GOPATH}/bin/snmp_exporter .
go build -ldflags="-s -w" -o %{_GOPATH}/bin/generator ./generator

%install
install -m 0755 -d $RPM_BUILD_ROOT/opt/ss/snmp_exporter/bin
install -m 0755 %{_GOPATH}/bin/snmp_exporter $RPM_BUILD_ROOT/opt/ss/snmp_exporter/bin/
install -m 0755 %{_GOPATH}/bin/generator $RPM_BUILD_ROOT/opt/ss/snmp_exporter/bin/
cp -pa generator/generator.yml $RPM_BUILD_ROOT/opt/ss/snmp_exporter/

%clean
rm -rf $RPM_BUILD_ROOT

%postun
# uninstall
if [ "$1" = "0" ]; then
    rm -rf /opt/ss/snmp_exporter
    echo "Uninstall complete."
fi

%files
%dir /opt/ss/snmp_exporter
%dir /opt/ss/snmp_exporter/bin
/opt/ss/snmp_exporter/*
/opt/ss/snmp_exporter/bin/*
