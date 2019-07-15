# N-Factor Vault - Monoservice

### backend/cmd/monoservice

This is the _monoservice_ package. Monoservice refers to "one service that includes all the services", it's the opposite of microservices. 

At this point, this is the main and the only executable/binary of the **n-factor-vault**. Hence, this is where the backend code starts.

## Development 

### **Setting up**

Setup the dev database and other pre-requisites using:

```
make dev-init
```

You can start the HTTP server and the corresponding database server using:

```
make dev-run
```

You can run the available tests using:
```
make dev-go-test
```

You can cleanup the locally created files (such as DB stuff, binaries) and return to a clean slate using:
```
make dev-clean
```

_Note_: While you run these _make_ commands, you might notice some errors in the terminal that are followed by keyword `(ignored)`. Those errors are to be expected under certain scenarios and can be ignored. E.g. a make command trying to stop the DB server but DB server is already stopped will result in an ignorable error.

### **Usage**
Once the server is running, it exposes a REST API that can be used to interact with the application. Sample HTTP requests for some available endpoints is in the example directory, and also listed below:

_Note 1_: Replace <TOKEN> with JWT auth token.

* **Signup**: Create a new user

    ```curl -v localhost:8080/v1/signup -d '{"name":"Jon Doe", "email":"jon@email.com", "password":"jon has a secret"}'```

* **Login**: # Login the user and returns a JWT auth token

    ```curl localhost:8080/v1/login -d '{"email":"jon@email.com", "password":"jon has a secret"}'```

* **Create Vault**: # Creates a new vault for the authenticated user.

    ```curl -v localhost:8080/v1/vault -d '{"name":"Twitter", "description":"Test vault"}' -H 'Authorization: Bearer <TOKEN>'```

* **Get Vaults**: Fetch all the vaults the authenticated user is a part of.

    ```curl -v localhost:8080/v1/vaults -H 'Authorization: Bearer <TOKEN>'```

* **Add User to Vault**: Associates a user to a vault

    ```curl localhost:8080/v1/vault/<vault_id>/user -d '{"user_id":"<user_id>"}' -H 'Authorization: Bearer <TOKEN>'```