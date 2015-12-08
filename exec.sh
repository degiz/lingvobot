#!/bin/bash

/usr/bin/redis-server &
/srv/lingvobot --config /srv/data/config.json &> /srv/log