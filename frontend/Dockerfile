FROM node:18.12 as dependencies

USER node

RUN mkdir /home/node/app

WORKDIR /home/node/app

COPY --chown=node:node yarn.lock package.json ./

RUN yarn

FROM dependencies as builder

COPY --chown=node:node . .

ENV API_URL=$API_URL

RUN yarn build

CMD ["yarn", "start"]