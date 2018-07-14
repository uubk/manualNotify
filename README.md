# manualNotify
Manually send NOTIFYs to specific hosts based on bind 9 log output

## Background
When using `bind-dyndb-ldap`, I came across [Bug 152](https://pagure.io/bind-dyndb-ldap/issue/152), that is, the missing
support of `also-notify`. `also-notify` is normally used to send DNS `NOTIFY` messages to additional hosts that are not
part of the `NS` set of a zone, in my case, my upstream provider uses a non-`NS` host for incoming zone transfers.
Since fixing the bug appears to require substantial code changes in both `bind` and `bind-dyndb-ldap`, I wrote this
workaround...

## Usage
A configuration file looks like this:
```
hostname: my-ns.example.org
resolvconf: /etc/resolv.conf
unit: named
zones:
  - name: example.org
    destination: provider.example.org:53
    issigned: false
```
This will lead to the following:
* The service will wait for named to echo a message of the format `zone example.org/IN/external: sending notifies (serial XXX)`
* Afterwards, it will fetch the `SOA` for `example.org`
* If the `SOA` is equal to `my-ns.example.org.`, it will send a `NOTIFY` to `provider.example.org:53`