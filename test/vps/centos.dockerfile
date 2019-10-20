ARG VERSION
FROM centos:${VERSION}

ENV container docker

# Get systemd working properly to start docker
RUN yum -y install initscripts ; yum clean all
RUN yum -y install initscripts ; yum -y install systemd; yum clean all; \
    (cd /lib/systemd/system/sysinit.target.wants/; for i in *; do [ $i == systemd-tmpfiles-setup.service ] || rm -f $i; done); \
    rm -f /lib/systemd/system/multi-user.target.wants/*;\
    rm -f /etc/systemd/system/*.wants/*;\
    rm -f /lib/systemd/system/local-fs.target.wants/*; \
    rm -f /lib/systemd/system/sockets.target.wants/*udev*; \
    rm -f /lib/systemd/system/sockets.target.wants/*initctl*; \
    rm -f /lib/systemd/system/basic.target.wants/*;\
    rm -f /lib/systemd/system/anaconda.target.wants/*;
RUN yum install -y sudo; \
    yum install -y openssh-server; \
    yum install -y openssh-clients;

# Set up ssh
RUN mkdir /var/run/sshd
RUN echo 'root:inertia' | chpasswd
RUN sed -i 's/PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config; \
    echo "AuthorizedKeysFile     %h/.ssh/authorized_keys" >> /etc/ssh/sshd_config;

# SSH login fix. Otherwise user is kicked off after login
RUN sed 's@session\s*required\s*pam_loginuid.so@session optional pam_loginuid.so@g' -i /etc/pam.d/sshd

# Cent wants this for some reason
RUN ssh-keygen -f /etc/ssh/ssh_host_rsa_key -N '' -t rsa

ENV NOTVISIBLE "in users profile"
RUN echo "export VISIBLE=now" >> /etc/profile

# Copy test key to allow use
RUN mkdir $HOME/.ssh/
COPY ./keys/ .
RUN cat id_rsa.pub >> $HOME/.ssh/authorized_keys

# Copy certs
RUN mkdir ~/.inertia/ ; mkdir ~/.inertia/.ssl/
COPY ./certs/ .
RUN mv daemon.cert ~/.inertia/.ssl ; mv daemon.key ~/.inertia/.ssl

# Copy dockerd configuration
COPY ./vps/daemon.json /etc/docker/daemon.json

EXPOSE 0-9000
CMD ["/usr/sbin/init", "/usr/sbin/sshd", "-D"]
