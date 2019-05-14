FROM openjdk:8-alpine

ARG HDP_VERSION_ARG=3.1.3.0-32
ARG SPARK_VERSION_ARG=2.4.2.3.1.3.0-32
ARG HADOOP_VERSION_ARG=3.1.1.3.1.3.0-32

ENV BASE_IMAGE      openjdk:8-alpine
ENV HDP_VERSION     $HDP_VERSION_ARG
ENV SPARK_VERSION   $SPARK_VERSION_ARG
ENV HADOOP_VERSION  $HADOOP_VERSION_ARG

ENV SPARK_HOME      /opt/spark
ENV SPARK_CONF_DIR  $SPARK_HOME/conf
ENV PATH            $PATH:$SPARK_HOME/bin

RUN set -ex && \
    apk upgrade --no-cache && \
    apk --update add --no-cache bash tini libstdc++ glib gcompat libc6-compat linux-pam krb5 krb5-libs nss openssl wget sed curl && \
    rm /bin/sh && \
    ln -sv /bin/bash /bin/sh && \
    echo "auth required pam_wheel.so use_uid" >> /etc/pam.d/su && \
    chgrp root /etc/passwd && chmod ug+rw /etc/passwd

### install spark
#RUN wget -O /spark-2.4.2.3.1.3.0-27-bin-3.1.1.3.1.3.0-27.tgz http://s3.amazonaws.com/devci.hortonworks.com/aiyer/4035001/HDP/centos7/3.x/BUILDS/3.1.3.0-27/tars/spark2/spark-2.4.2.3.1.3.0-27-bin-3.1.1.3.1.3.0-27.tgz && \
COPY spark-${SPARK_VERSION}-bin-${HADOOP_VERSION}.tgz /
RUN tar -xzf /spark-${SPARK_VERSION}-bin-${HADOOP_VERSION}.tgz -C /opt/ && \
    ln -s /opt/spark-${SPARK_VERSION}-bin-${HADOOP_VERSION} $SPARK_HOME && \
    rm -f /spark-${SPARK_VERSION}-bin-${HADOOP_VERSION}.tgz && \
    mkdir -p $SPARK_HOME/work-dir && \
    mkdir -p $SPARK_HOME/spark-warehouse

COPY conf/* $SPARK_CONF_DIR/
COPY entrypoint.sh /opt/
COPY Dockerfile /my_docker/

COPY conf/log4j.properties /my_docker/conf/
ENV SPARK_CLASSPATH="/my_docker/conf"

RUN chmod +x /opt/*.sh

WORKDIR $SPARK_HOME/work-dir
ENTRYPOINT [ "/opt/entrypoint.sh" ]
