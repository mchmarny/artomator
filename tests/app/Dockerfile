# TODO: migrate to cgr.dev/chainguard/node:latest when arm supported 
FROM node:18-alpine
WORKDIR /app/
COPY ./index.js .
COPY ./package.json .
RUN npm install
CMD ["node", "index.js"]
EXPOSE 8080