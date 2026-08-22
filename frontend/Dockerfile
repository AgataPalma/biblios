ARG NODE_IMAGE=node:24.19.0-alpine3.24@sha256:d32cdf619f63fe0471182d08996dd516c6275bb5fd31ae06e55a570bd9e1ad43
ARG NGINX_IMAGE=nginx:1.29.4-alpine3.23@sha256:4870c12cd2ca986de501a804b4f506ad3875a0b1874940ba0a2c7f763f1855b2

FROM ${NODE_IMAGE} AS dependencies

WORKDIR /app

COPY --chown=node:node package.json package-lock.json ./
RUN chown node:node /app

USER node

RUN npm ci \
    && sha256sum package.json package-lock.json \
        | sha256sum \
        | awk '{ print $1 }' > node_modules/.biblios-dependencies.sha256


FROM dependencies AS development

COPY --chown=node:node . .
COPY --chmod=755 docker/dev-entrypoint.sh /usr/local/bin/frontend-dev-entrypoint

EXPOSE 5173

ENTRYPOINT ["frontend-dev-entrypoint"]
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]


FROM dependencies AS build

ARG VITE_API_URL=/api/v1
ENV VITE_API_URL=${VITE_API_URL}

COPY --chown=node:node . .

RUN npm run build


FROM ${NGINX_IMAGE} AS production

ARG APP_VERSION=unversioned
ARG VCS_REF=unknown
ARG BUILD_DATE=unknown

LABEL org.opencontainers.image.title="Biblios frontend" \
      org.opencontainers.image.description="Production web runtime for the Biblios single-page application" \
      org.opencontainers.image.version="${APP_VERSION}" \
      org.opencontainers.image.revision="${VCS_REF}" \
      org.opencontainers.image.created="${BUILD_DATE}"

COPY docker/nginx.conf /etc/nginx/nginx.conf
COPY --from=build /app/dist /usr/share/nginx/html

RUN chown -R nginx:nginx /usr/share/nginx/html

USER nginx

EXPOSE 8080

HEALTHCHECK --interval=10s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/healthz >/dev/null || exit 1

ENTRYPOINT []
CMD ["nginx", "-g", "daemon off;"]
