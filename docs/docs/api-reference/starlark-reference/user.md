---
title: User
sidebar_label: User
---

The `User` constructor creates a `User` object that represents a `uid` and `gid` that the service starts with(see `ServiceConfig`[service-config] object)

```python
user = User(uid=0)

# or

user = User(uid=0, gid=0)
```

Note that the `gid` is completely optional.

<!--------------- ONLY LINKS BELOW THIS POINT ---------------------->
[service-config]: ./service-config.md