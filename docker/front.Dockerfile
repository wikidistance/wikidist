FROM node:13

RUN mkdir /front
COPY ./frontend /front

WORKDIR /front

RUN yarn install

CMD ["yarn", "serve"]