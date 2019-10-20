ARG VERSION
FROM amazonlinux:${VERSION}

ENV container docker

RUN yum install -y sudo; \
    yum install -y openssh-server; \
    yum install -y openssh-clients;

# Set up SSH
RUN mkdir /var/run/sshd
RUN echo 'root:inertia' | chpasswd
RUN sed -i 's/PermitRootLogin forced-commands-only/PermitRootLogin without-password/' /etc/ssh/sshd_config; \
    echo "PubkeyAuthentication  yes" >> /etc/ssh/sshd_config; \
    echo "RSAAuthentication yes" >> /etc/ssh/sshd_config;

# SSH login fix. Otherwise user is kicked off after login
RUN sed 's@session\s*required\s*pam_loginuid.so@session optional pam_loginuid.so@g' -i /etc/pam.d/sshd

# Cent wants this for some reason
RUN ssh-keygen -f /etc/ssh/ssh_host_rsa_key -N '' -t rsa

ENV NOTVISIBLE "in users profile"
RUN echo "export VISIBLE=now" >> /etc/profile

# Copy test key to allow use
RUN mkdir ~/.ssh/ ; touch ~/.ssh/authorized_keys
COPY ./keys/ .
RUN cat id_rsa.pub >> ~/.ssh/authorized_keys

# Copy certs
RUN mkdir ~/.inertia/ ; mkdir ~/.inertia/.ssl/
COPY ./certs/ .
RUN mv daemon.cert ~/.inertia/.ssl ; mv daemon.key ~/.inertia/.ssl

# Copy dockerd configuration
COPY ./vps/daemon.json /etc/docker/daemon.json

EXPOSE 0-9000
CMD ["/usr/sbin/sshd", "-D"]
