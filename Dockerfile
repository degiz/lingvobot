FROM debian:jessie

RUN apt-get update -q
RUN apt-get install -y redis-server && apt-get install -y ca-certificates

RUN \
 groupadd -r lingvobot && useradd -r -g lingvobot lingvobot && \
 mkdir /srv/data && \
 chown -R lingvobot:lingvobot /srv

USER lingvobot

ADD nouns.txt /srv/data/nouns.txt
ADD config.json /srv/data/config.json
ADD lingvobot /srv/lingvobot
ADD exec.sh /srv/exec.sh

WORKDIR /srv
ENTRYPOINT ["/srv/exec.sh"]