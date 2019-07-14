# This creates a new vault for user.
# Replace <TOKEN> with JWT token obtained after login.
curl -v localhost:8080/v1/vault -d '{"name":"Twitter", "description":"Test vault"}' -H 'Authorization: Bearer <TOKEN>'
