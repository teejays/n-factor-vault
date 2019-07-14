# This creates a new user for Jon
curl -v localhost:8080/v1/signup -d '{"name":"Jon Doe", "email":"jon@email.com", "password":"jon has a secret"}'

# This creates a new user for Jane
curl -v localhost:8080/v1/signup -d '{"name":"Jane Does", "email":"jane@email.com", "password":"jane has a secret"}'
