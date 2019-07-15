# This logins JON and returns a JWT token which should be used for subsequent authenticated requests
curl localhost:8080/v1/login -d '{"email":"jon@email.com", "password":"jon has a secret"}'
