FROM alpine/curl:latest AS jar_getter
ARG FabricInstallerVersion
ARG FabricLoaderVersion
ARG MinecraftVersion
RUN set -eux && curl -o /server.jar -OJ https://meta.fabricmc.net/v2/versions/loader/${MinecraftVersion}/${FabricLoaderVersion}/${FabricInstallerVersion}/server/jar

FROM alpine:latest AS base_image
RUN set -eux && apk upgrade --no-cache
RUN set -eux && apk add openjdk21-jre

FROM scratch
ENV FABRIC_MEMINI=2G
ENV FABRIC_MEMMAX=2G
EXPOSE 25565 25575
COPY --from=base_image / /
COPY --from=jar_getter server.jar /usr/bin/
WORKDIR /opt/
CMD [ "sh", "-c", "/usr/bin/java -Xms${FABRIC_MEMINI} -Xmx${FABRIC_MEMMAX} -jar /usr/bin/server.jar nogui" ]
