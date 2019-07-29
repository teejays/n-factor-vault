# N-Factor Vault - Backend

This directory holds all of the backend code, and executable applications. 

## Monoservice:

* Prerequisites: 

    - Install Docker: https://docs.docker.com/
    - Install Docker Compose: https://docs.docker.com/compose/install/


* Run:
    
    Monoservice API and Postgres database server can be started using docker-compose. The configuration for docker compose is in the `docker-compose` file. Use the following command to start Monoservice.
        
        make run

    You can also run Go tests using

        make run-test

    Both of these run using docker containers.