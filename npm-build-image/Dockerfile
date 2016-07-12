FROM        node:6.2.2
MAINTAINER  Kindly Ops, LLC <support@kindlyops.com>
ENV location /node
RUN mkdir $location
WORKDIR ${location}
ADD package.json $location
RUN npm install -g
