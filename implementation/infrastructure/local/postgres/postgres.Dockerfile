FROM postgres:16.8
LABEL maintainer="Alexander Kukushkin <akukushkin@microsoft.com>"

RUN export DEBIAN_FRONTEND=noninteractive \
    && echo 'APT::Install-Recommends "0";\nAPT::Install-Suggests "0";' > /etc/apt/apt.conf.d/01norecommend \
    && apt-get update -y \
    && apt-cache depends patroni | sed -n -e 's/.* Depends: \(python3-.\+\)$/\1/p' \
    | grep -Ev '^python3-(sphinx|etcd|consul|kazoo|kubernetes)' \
    | xargs apt-get install -y vim-tiny curl jq locales git python3-pip python3-wheel \
    ## Make sure we have a en_US.UTF-8 locale available
    && localedef -i en_US -c -f UTF-8 -A /usr/share/locale/locale.alias en_US.UTF-8 \
    && pip3 install --break-system-packages setuptools \
    && pip3 install --break-system-packages 'git+https://github.com/patroni/patroni.git#egg=patroni[kubernetes]' \
    && PGHOME=/home/postgres \
    && mkdir -p $PGHOME \
    && chown postgres $PGHOME \
    && sed -i "s|/var/lib/postgresql.*|$PGHOME:/bin/bash|" /etc/passwd \
    # Set permissions for OpenShift
    && chmod 775 $PGHOME \
    && chmod 664 /etc/passwd \
    # Clean up
    && apt-get remove -y git python3-pip python3-wheel \
    && apt-get autoremove -y \
    && apt-get clean -y \
    && rm -rf /var/lib/apt/lists/* /root/.cache

COPY entrypoint.sh /entrypoint.sh

COPY certs/ca.pem /etc/ssl/pg-ca.pem
COPY certs/server.crt /etc/ssl/pg-server-cert.crt
COPY certs/server.key /etc/ssl/private/pg-server-key.key

RUN chmod 664 /etc/ssl/pg-ca.pem
RUN chmod 664 /etc/ssl/pg-server-cert.crt
RUN chown postgres /etc/ssl/private/pg-server-key.key
RUN chmod 600 /etc/ssl/private/pg-server-key.key

ENV PGSSLMODE=verify-ca 
ENV PGSSLKEY=/etc/ssl/private/pg-server-key.key
ENV PGSSLCERT=/etc/ssl/pg-server-cert.crt
ENV PGSSLROOTCERT=/etc/ssl/pg-ca.pem

EXPOSE 5432 8008
ENV LC_ALL=en_US.UTF-8 LANG=en_US.UTF-8 EDITOR=/usr/bin/editor
USER postgres
WORKDIR /home/postgres
CMD ["/bin/bash", "/entrypoint.sh"]