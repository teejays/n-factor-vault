# This returns all the vaults that the user is a part of
# Replace <TOKEN> with JWT token obtained after login.
curl -v localhost:8080/v1/vaults -H 'Authorization: Bearer <TOKEN>'