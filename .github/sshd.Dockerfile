# docker build -t dummy-sshd-container -f sshd.Dockerfile .
# docker run -d -p 127.0.0.1:2222:22 --name dummy-sshd dummy-sshd-container

FROM ubuntu:latest
RUN apt-get update && apt-get install -y openssh-server
# Configure SSH
RUN mkdir /var/run/sshd

# Add test user password
RUN useradd -m -d /home/test -s /bin/bash test
RUN echo 'test:test' | chpasswd

# Start SSH server
CMD ["/usr/sbin/sshd", "-D"]