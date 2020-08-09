# Request Tracking 

## Elasticsearch 

Moxsoar uses Elasticsearch to perform *request tracking*. Request tracking logs all HTTP requests, including the body
and headers, as Elasticsearch documents so they can be later searched.

Moxsoar will run happily without an ES instance available and simply won't log requests in this instance, which is 
useful when purely using as a mock tool.

## Quickly Running elasticsearch

Elasticsearch is not part of Moxsoar but you can run the two together in one command using 
[docker compose.](../docker-compose.yml)

After starting, you must set the built in passwords for Elasticsearch. The following command sets them interactively.
```bash
docker-compose exec elasticsearch ./bin/elasticsearch-setup-passwords interactive
```

After setting your password in Elasticsearch, pass it (and the Kibana user) as environment variables in the docker
compose file.

 