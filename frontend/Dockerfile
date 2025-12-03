FROM node:20-alpine
WORKDIR /build
COPY ./package*.json ./
RUN npm ci
COPY ./ ./
RUN npm run build
RUN cp -r dist /app/dist
CMD ["sleep", "infinity"]