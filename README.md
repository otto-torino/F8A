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

## Features

### Deployment Tracking

F8A tracks all deployments with detailed progress information:

- **Real-time Progress**: See which step is currently running during deployment
- **Time Estimates**: View average duration for each step based on historical data
- **Deployment History**: Access last 10 deployments with status, duration, and commit details
- **Status Dashboard**: Compare local vs deployed revisions at a glance
- **Desktop Notifications**: Get notified when deployments complete or fail

### Deployment Steps

Each deployment consists of 7 tracked steps:
1. Build - Run yarn build locally
2. Archive - Create tar file
3. Upload - SCP to remote server
4. Backup - Move current version to previous
5. Extract - Extract new build
6. Activate - Create symlink
7. Cleanup - Remove tar and copy .htaccess if needed
