FROM node:22-alpine AS build
WORKDIR /app
ARG VITE_API_URL=https://media-scheduler-backend.onrender.com
ENV VITE_API_URL=${VITE_API_URL}
COPY package*.json ./
RUN npm install
COPY . .
RUN npm run build

FROM node:22-alpine
WORKDIR /app
COPY --from=build /app/package*.json ./
COPY --from=build /app/node_modules ./node_modules
COPY --from=build /app/dist ./dist
RUN npm install -g serve
ENV PORT=4173
EXPOSE 4173
CMD ["sh", "-c", "serve -s dist -l ${PORT:-4173}"]
