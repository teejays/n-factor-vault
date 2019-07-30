# N-Factor Vault

### backend/db/data

This directory will be a volume mounted on the Postgres DB docker container. This directory is where the Postgres database, run using `docker-compose`, will keep it's data. The reason the DB data is kept on host is so that subsequent initializations of docker container should not ideally remove the data. However, since this database is only meant for local development, we are not including the files in here in Git, hence `.gitignore` includes everything but this `README` file.