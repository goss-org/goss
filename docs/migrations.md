# Migration guide

## v4 migration

### Array matchers (e.g. user.groups) no longer allows duplicates

Goss v0.3.X allowed:

```yaml
user:
  root:
    exists: true
    groups:
      - root
      - root
      - root
```

Goss v0.4.x, will fail with the above as group "root" is only in the slice once. However, with goss v0.4.x the array may
contain matchers. The test below is valid for v0.4.x but not valid for v0.3.x

```yaml
user:
  root:
    exists: true
    groups:
      - have-prefix: r
```

## rpm now contains the full EVR version

To enable the ability to compare RPM versions in the future, The version matching of rpm has changed

from:

```console
rpm -q --nosignature --nohdrchk --nodigest --qf '%{VERSION}\n' package_name
```

to:

```console
rpm -q --nosignature --nohdrchk --nodigest --qf '%|EPOCH?{%{EPOCH}:}:{}|%{VERSION}-%{RELEASE}\n' package_name
```

## `file.contains` -> `file.contents`

File contains attribute has been renamed to file.contents

from:

```yaml
file:
  /tmp/foo:
    exists: true
    contains: []
```

to:

```yaml
file:
  /tmp/foo:
    exists: true
    contents: []
```
