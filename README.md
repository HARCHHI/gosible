# Gosible

A simple tool  made by `Golang` can help you deployment app automatically and parallelly.

## Required

`Zero dependency` in client, it will work through pure SSH.

## Build

Require:
  * `golang >= 1.11`

Run:
```
git clone https://github.com/HARCHHI/gosible.git
cd gosible
go build
```

## Usage

Run:
```bash
gosible -f config.yaml
```

You can set all you want to copy and execute in a yaml file.
Gosible has some root keywords in yaml file:

* configs
* devices
* cpoy
* execute

### configs
You can put auth config `privateKey` and proxy setting `proxy` in configs field, it will work on every devices.

```yaml
configs:
  privateKey: /path/to/your/privateKey/file
  proxy:
    user: user
    addr: 127.0.0.1
    port: 22
    password: pwd
```

### devices

A yaml array with all devices you want to configure.

```yaml
devices:
  - user: user1
    addr: 127.0.0.1
    port: 22
    password: pwd
  - user: user1
    addr: 127.0.0.2
    port: 22
    password: pwd
```

### cpoy

Files you want to copy.

```yaml
copy:
  - source:
      - ./file1
      - ./file2
    destination: /home/root/
  - source:
      - ./file3
      - ./file4
    destination: /var/some/dest
```

### execute

Files you want to execute. You can copy any file you want to execute in device; Execute action will work after copy action.

```yaml
execute:
  - /home/vagrant/echo.sh
```

### Completely config yaml

```yaml
configs:
  privateKey: /home/user/.ssh/id_rsa
devices:
  - user: user
    addr: someaddress
    port: 22
copy:
  - source:
      - ./echo.sh
    destination: /home/root
execute:
  - /home/root/echo.sh
```