# F8A

Otto Frontend Application Manager

Stack:
- go
- fyne
- sqlite
- viper

## Create package

```
$ fyne package -os linux -icon Icon.png
```

## Install

Create `/etc/f8a.json` with the following content:

```
{
  "app": {
    "homePath": "/home/USER/.f8a"
  }
}
```


```
$ tar xvf f8a.tar.gz
$ cd f8a
$ make user-install
```
