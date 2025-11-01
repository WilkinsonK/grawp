FROM alpine/curl:latest AS jar_getter
ARG PapermcEndpoint
RUN set -eux && curl -o /server.jar -OJ https://fill-data.papermc.io/v1/objects/${PapermcEndpoint}

FROM alpine:latest AS base_image
RUN set -eux && apk upgrade --no-cache
RUN set -eux && apk add openjdk21-jre

FROM scratch
ENV PAPERMC_MEMINI=2G
ENV PAPERMC_MEMMAX=2G
EXPOSE 25565 25575
COPY --from=base_image / /
COPY --from=jar_getter server.jar /usr/bin/
WORKDIR /opt/
CMD [ "sh", "-c", "/usr/bin/java -Xms${PAPERMC_MEMINI} -Xmx${PAPERMC_MEMMAX} -XX:+AlwaysPreTouch -XX:+DisableExplicitGC -XX:+ParallelRefProcEnabled -XX:+PerfDisableSharedMem -XX:+UnlockExperimentalVMOptions -XX:+UseG1GC -XX:G1HeapRegionSize=8M -XX:G1HeapWastePercent=5 -XX:G1MaxNewSizePercent=40 -XX:G1MixedGCCountTarget=4 -XX:G1MixedGCLiveThresholdPercent=90 -XX:G1NewSizePercent=30 -XX:G1RSetUpdatingPauseTimePercent=5 -XX:G1ReservePercent=20 -XX:InitiatingHeapOccupancyPercent=15 -XX:MaxGCPauseMillis=200 -XX:MaxTenuringThreshold=1 -XX:SurvivorRatio=32 -jar /usr/bin/server.jar nogui" ]
