# Bundler
FROM node:22.7.0-bookworm-slim AS assets

WORKDIR /tmp

COPY *.json .
COPY *.ts .
RUN npm install

COPY ./public ./public/
COPY ./src ./src/
COPY ./index.html ./index.html
RUN npm run build

# Nginx
FROM nginx:1.27-alpine3.20

COPY --from=assets /tmp/dist /mnt/html
COPY public/favicon.ico /mnt/html/favicon.ico
COPY nginx.conf /etc/nginx/nginx.conf

EXPOSE 8080