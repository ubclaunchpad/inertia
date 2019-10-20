ARG VERSION
FROM debian:${VERSION}

RUN apt-get update && apt-get install -y sudo && rm -rf /var/lib/apt/lists/*
RUN apt-get update && apt-get install -y openssh-server
RUN mkdir /var/run/sshd
RUN echo 'root:inertia' | chpasswd

RUN sed -i 's/PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config
RUN echo "AuthorizedKeysFile     %h/.ssh/authorized_keys" >> /etc/ssh/sshd_config

# SSH login fix. Otherwise user is kicked off after login
RUN sed 's@session\s*required\s*pam_loginuid.so@session optional pam_loginuid.so@g' -i /etc/pam.d/sshd

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
CMD ["/usr/sbin/sshd", "-D"]
