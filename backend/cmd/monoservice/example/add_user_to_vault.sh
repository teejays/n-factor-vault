# This adds a new user to an existing vault
# Replace <TOKEN> with JWT token obtained after login.
# Replace <user_id> with the user_id of the user you are adding to the vault
# Replace <vault_id> with the vault_id of the vault you are adding a user to
curl localhost:8080/v1/vault/<vault_id>/user -d '{"user_id":"<user_id>"}' -H 'Authorization: Bearer <TOKEN>'