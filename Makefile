# Copyright 2016 The Prometheus Authors
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

BUILDDIR	?= /tmp/ssmbuild
VERSION		?=
RELEASE		?= 1

ifeq (0, $(shell hash dpkg 2>/dev/null; echo $$?))
ARCH	:= $(shell dpkg --print-architecture)
else
ARCH	:= $(shell rpm --eval "%{_arch}")
endif

TARBALL_FILE	:= $(BUILDDIR)/tarballs/snmp_exporter-$(VERSION)-$(RELEASE).tar.gz
SRPM_FILE		:= $(BUILDDIR)/results/SRPMS/snmp_exporter-$(VERSION)-$(RELEASE).src.rpm
RPM_FILE		:= $(BUILDDIR)/results/RPMS/snmp_exporter-$(VERSION)-$(RELEASE).$(ARCH).rpm

.PHONY: all
all: srpm rpm

$(TARBALL_FILE):
	mkdir -vp $(shell dirname $(TARBALL_FILE))

	GO111MODULE=on go mod vendor

	tar --exclude-vcs -czf $(TARBALL_FILE) -C $(shell dirname $(CURDIR)) --transform s/^$(shell basename $(CURDIR))/snmp_exporter/ $(shell basename $(CURDIR))

.PHONY: srpm
srpm: $(SRPM_FILE)

$(SRPM_FILE): $(TARBALL_FILE)
	mkdir -vp $(BUILDDIR)/rpmbuild/{SOURCES,SPECS,BUILD,SRPMS,RPMS}
	mkdir -vp $(shell dirname $(SRPM_FILE))

	cp snmp_exporter.spec $(BUILDDIR)/rpmbuild/SPECS/snmp_exporter.spec
	sed -i "s/%{_version}/$(VERSION)/g" "$(BUILDDIR)/rpmbuild/SPECS/snmp_exporter.spec"
	sed -i "s/%{_release}/$(RELEASE)/g" "$(BUILDDIR)/rpmbuild/SPECS/snmp_exporter.spec"
	cp $(TARBALL_FILE) $(BUILDDIR)/rpmbuild/SOURCES/
	rpmbuild -bs --define "debug_package %{nil}" --define "_topdir $(BUILDDIR)/rpmbuild" $(BUILDDIR)/rpmbuild/SPECS/snmp_exporter.spec
	mv $(BUILDDIR)/rpmbuild/SRPMS/$(shell basename $(SRPM_FILE)) $(SRPM_FILE)

.PHONY: rpm
rpm: $(RPM_FILE)

$(RPM_FILE): $(SRPM_FILE)
	mkdir -vp $(BUILDDIR)/mock $(shell dirname $(RPM_FILE))
	mock -r ssm-9-$$(rpm --eval "%{_arch}") --resultdir $(BUILDDIR)/mock --rebuild $(SRPM_FILE)
	mv $(BUILDDIR)/mock/$(shell basename $(RPM_FILE)) $(RPM_FILE)

.PHONY: clean
clean:
	rm -rf $(BUILDDIR)/{tarballs,rpmbuild,mock,results}
