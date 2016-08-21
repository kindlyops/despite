FROM        node:6.2.2
MAINTAINER  Kindly Ops, LLC <support@kindlyops.com>
ENV location /code
RUN mkdir $location
WORKDIR ${location}
ADD package.json npm-shrinkwrap.json $location/
RUN npm install
