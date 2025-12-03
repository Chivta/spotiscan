FROM node:20-alpine
WORKDIR /build
COPY ./package*.json ./
RUN npm ci
COPY ./ ./
RUN npm run build
RUN mkdir -p /app && cp -r dist /app/dist
CMD ["sleep", "infinity"]