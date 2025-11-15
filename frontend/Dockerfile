# Build frontend
FROM node:20-alpine AS frontend
WORKDIR /build
COPY ./package*.json ./
RUN npm ci
COPY ./ ./
RUN npm run build

# Serve with nginx
FROM nginx:alpine
COPY --from=frontend /build/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]