FROM alpine/curl:latest AS jar_getter
ARG VelocityEndpoint
RUN set -eux && curl -o /proxy.jar -OJ https://fill-data.papermc.io/v1/objects/${VelocityEndpoint}

FROM alpine:latest AS base_image
RUN set -eux && apk upgrade --no-cache
RUN set -eux && apk add openjdk21-jre

FROM scratch
ENV VELOCITY_MEMINI=2G
ENV VELOCITY_MEMMAX=2G
EXPOSE 25565 25565
COPY --from=base_image / /
COPY --from=jar_getter proxy.jar /usr/bin/
WORKDIR /opt/
CMD [ "sh", "-c", "/usr/bin/java -Xms${VELOCITY_MEMINI} -Xmx${VELOCITY_MEMMAX} -XX:+UseG1GC -XX:G1HeapRegionSize=4M -XX:+UnlockExperimentalVMOptions -XX:+ParallelRefProcEnabled -XX:+AlwaysPreTouch -XX:MaxInlineLevel=15 -jar /usr/bin/proxy.jar" ]
