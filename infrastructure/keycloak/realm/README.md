The `onboarder` realm is the authentication/control-plane realm.

The three `target-*` realms are disposable development targets used to test
promotion and realm isolation. Each contains a least-privilege service account:

```text
client ID: onboarder-target-admin
secret: target-admin-secret
```

Never use these example credentials in production. Create a separate
service-account client and secret in every managed realm.
