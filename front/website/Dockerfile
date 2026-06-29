ARG NODE_IMAGE=node:20-alpine
ARG NGINX_IMAGE=nginx:alpine

FROM ${NODE_IMAGE} AS base
WORKDIR /app
COPY website/package.json ./
COPY website/package-lock.json ./
RUN npm install

FROM base AS dev
COPY website/ .
COPY shared/ /shared/
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0", "--port", "5173"]

FROM base AS build
COPY website/ .
COPY shared/ /shared/
# These must be present at build time (Vite bakes them into the bundle).
ARG VITE_API_BASE_URL
ARG VITE_IMAGE_BASE_URL
ARG VITE_SITE_URL
ARG VITE_PRERENDER_API_BASE_URL
ARG VITE_SITEMAP_LASTMOD
ARG VITE_DEFAULT_LANG
ARG PRERENDER_CACHE_BUST
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
ENV VITE_IMAGE_BASE_URL=$VITE_IMAGE_BASE_URL
ENV VITE_SITE_URL=$VITE_SITE_URL
ENV VITE_PRERENDER_API_BASE_URL=$VITE_PRERENDER_API_BASE_URL
ENV VITE_SITEMAP_LASTMOD=$VITE_SITEMAP_LASTMOD
ENV VITE_DEFAULT_LANG=$VITE_DEFAULT_LANG
RUN printf '%s\n' "$PRERENDER_CACHE_BUST" >/tmp/prerender-cache-bust && npm run build

FROM ${NGINX_IMAGE} AS prod
COPY --from=build /app/dist /usr/share/nginx/html
COPY website/nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
