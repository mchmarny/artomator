FROM node:18-alpine
COPY tests/app/ ./
WORKDIR /app/
RUN npm install
CMD ["node", "index.js"]
EXPOSE 8080