FROM sasnouskikh/livy-builder:0.1 as build

ENV GITHUB_REPO     jahstreet/incubator-livy
ENV GITHUB_BRANCH   v0.6.0-incubating-kubernetes-support

RUN git clone -b ${GITHUB_BRANCH} https://github.com/${GITHUB_REPO}.git && \
    cd incubator-livy && \
    mvn clean package -DskipTests=true && \
    cp assembly/target/apache-livy-0.7.0-incubating-SNAPSHOT-bin.zip /

FROM busybox

COPY --from=build /apache-livy-0.7.0-incubating-SNAPSHOT-bin.zip /apache-livy-0.7.0-incubating-SNAPSHOT-bin.zip
