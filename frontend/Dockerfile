FROM node:20-alpine
WORKDIR /build
COPY ./package*.json ./
RUN npm ci
COPY ./ ./
RUN npm run build

# Mount built files at /app/dist for host nginx to serve
CMD ["sleep", "infinity"]