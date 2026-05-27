FROM node:22-alpine AS build
ARG REACT_APP_API_URL
ENV REACT_APP_API_URL=$REACT_APP_API_URL
WORKDIR /app
COPY package*.json ./
COPY public/* ./public/
RUN npm install -D tailwindcss postcss autoprefixer
RUN npm install
COPY . .
RUN npx tailwindcss init -p
RUN npm run build

FROM nginx:alpine
COPY --from=build /app/build /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]