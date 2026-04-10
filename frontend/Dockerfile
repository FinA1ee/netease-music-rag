# ── Stage 1: build React app ──────────────────────────────────────────────────
FROM node:lts-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .
RUN npm run build

# ── Stage 2: serve with nginx ─────────────────────────────────────────────────
FROM nginx:stable-alpine

# Replace the default nginx config with ours
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy the built static files from stage 1
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
