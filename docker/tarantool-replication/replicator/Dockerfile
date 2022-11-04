FROM centos:centos7

RUN yum -y install epel-release && \
    yum -y update && \
    yum -y clean all

RUN set -x \
    && yum -y install \
        libstdc++ \
        readline \
        openssl \
        yaml \
        lz4 \
        binutils \
        ncurses \
        libgomp \
        lua \
        curl \
        tar \
        zip \
        unzip \
        libunwind \
        libcurl \
        libicu \
    && yum -y install \
        perl \
        gcc-c++ \
        cmake \
        readline-devel \
        openssl-devel \
        libyaml-devel \
        lz4-devel \
        binutils-devel \
        ncurses-devel \
        lua-devel \
        make \
        git \
        libunwind-devel \
        autoconf \
        automake \
        libtool \
        go \
        wget \
        curl-devel \
        libicu-devel \
    && yum -y install \
        boost \
        boost-devel

RUN rpm --import https://repo.mysql.com/RPM-GPG-KEY-mysql-2022

##
## Install mysql57 client, devel packages
##
RUN yum localinstall -y https://dev.mysql.com/get/mysql80-community-release-el7-1.noarch.rpm \
    && yum --disablerepo=mysql80-community --enablerepo=mysql57-community install -y \
        mysql-community-client \
        mysql-community-devel


RUN git clone https://github.com/tarantool/mysql-tarantool-replication.git mysql_tarantool-replication

RUN cd mysql_tarantool-replication \
    && git submodule update --init --recursive

RUN cd mysql_tarantool-replication \
    && cmake . \
    && make

##
## Copy replicatord binary to new image to minimize size
##
FROM centos:centos7
COPY --from=0 mysql_tarantool-replication/replicatord .

ENTRYPOINT ["/replicatord",  "-l /dev/stdout"]
