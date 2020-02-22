# e5go

This service will call office365 api forever.

# build
- go 1.13

`go build e5go *.go`


# usage

create config file: `~/.config/e5go.yaml`, an example:

```yaml
apis:
  - https://graph.microsoft.com/v1.0/me/calendars
  - https://graph.microsoft.com/v1.0/me/messages
  - https://graph.microsoft.com/v1.0/me/
  - https://graph.microsoft.com/v1.0/me/contacts
client_id: <your client id>
client_secret: <your client secret>
listen: 127.0.0.1:3000
period: 30s
redirect_uri: http://localhost:3000/callback
scope:
  - openid
  - profile
  - offline_access
  - User.Read
  - Mail.ReadWrite
  - Calendars.ReadWrite
  - Contacts.ReadWrite
```

If you don't have a client_id, go to https://aad.portal.azure.com/ and register an application.

