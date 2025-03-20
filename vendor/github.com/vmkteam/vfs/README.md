# VFS Service
[![Release](https://img.shields.io/github/release/vmkteam/vfs.svg)](https://github.com/vmkteam/vfs/releases/latest)
[![Build Status](https://github.com/vmkteam/vfs/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/vmkteam/vfs/actions)

## Examples

### Hash upload

`curl --upload-file image.jpg  http://localhost:9999/upload/hash`

### How to run

    createdb vfs
    psql -f docs/vfs.sql vfs
    mkdir testdata
    make run

#### Upload image

    wget -O image.jpg https://media.myshows.me/shows/e/22/e22c3ab75b956c6c1c1fca8182db7efb.jpg
    export AUTHTOKEN=`curl http://localhost:9999/auth-token`    
    curl --upload-file image.jpg  -H  "AuthorizationJWT: ${AUTHTOKEN}" http://localhost:9999/upload/hash
    open http://localhost:9999/media/6/4a/64a9f060983200709061894cc5f69f83.jpg
