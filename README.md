# F8A

Otto Frontend Application Manager

Stack:
- go
- fyne
- sqlite
- viper

## Requirements

- Go
- fyne CLI: `go install fyne.io/tools/cmd/fyne@latest`
- Build dependencies (Debian/Ubuntu): `sudo apt install libgl1-mesa-dev xorg-dev libxkbcommon-dev`

## Build and package

```
$ make build      # build/f8a
$ make package    # f8a.tar.xz, the distribution tarball
```

`make package` runs `fyne package` and then fixes the generated tarball:

- the `.desktop` file's `Name=` is set to the window title (`utils.AppTitle`),
  since fyne uses `--name` for both the file names and the displayed name;
- the tarball's `Makefile` is replaced with `packaging/Makefile`, which
  installs the icon into the hicolor theme so it resolves by name.

The window's X11 WM_CLASS is set to `f8a` in `main.go` (Fyne would otherwise
use the window title), matching `f8a.desktop` so desktop environments show the
icon on the running window.

`make user-install` packages and installs into `~/.local` in one step.

## Install

Create `/etc/f8a.json` with the following content:

```
{
  "app": {
    "homePath": "/home/USER/.f8a"
  },
  "shell": {
    "init": ""
  }
}
```

`shell.init` is optional. Commands (git, yarn, scp, ...) run in a non-interactive
`bash -c` shell that inherits the environment of the desktop session, so tools
set up by your shell init files may not be found. Put the snippet needed to make
them available here; it is prepended to every command. For example, with
yarn installed through nvm:

```
"shell": {
  "init": ". $HOME/.nvm/nvm.sh"
}
```


```
$ tar xvf f8a.tar.xz
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
