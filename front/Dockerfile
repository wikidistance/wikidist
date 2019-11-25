FROM node:13
WORKDIR /usr/app
COPY package.json .
RUN yarn install
COPY . .
EXPOSE 8080
CMD ["yarn","serve"]