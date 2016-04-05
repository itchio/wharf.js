
var ZIP_RE = /\.zip$/i

function read (file) {
  return new Promise(function (resolve, reject) {
    var reader = new window.FileReader()
    reader.onload = function (evt) {
      resolve(evt.target.result)
    }
    reader.onerror = reject
    reader.readAsArrayBuffer(file)
  })
}

function handleZip (file, listener) {
  var container = {
    entries: []
  }

  read(file).then((bytes) => {
    var zip = new window.JSZip(bytes)
    for (var entryName in zip.files) {
      var entry = zip.files[entryName]
      if (entry.dir) {
        continue
      }
      console.log('zip entry: ', entry)
      var perms = entry.unixPermissions || 0
      console.log('entry perms: ', perms.toString(8))
      if (perms & (2 << 12) > 0) {
        // symlink!
        console.log('symlink target: ', entry.asText())
      }

      var size = (entry._data || {}).uncompressedSize
      if (!size) {
        size = entry.asArrayBuffer().byteLength
      }

      container.entries.push({
        path: entry.name,
        size: size,
        read: function (entry) {
          return new Promise(function (resolve, reject) {
            resolve(entry.asArrayBuffer())
          })
        }.bind(null, entry)
      })
    }
    listener(container)
  })
}

function handleChromeFolder (files, listener) {
  var container = {
    entries: []
  }

  for (var i = 0; i < files.length; i++) {
    var file = files[i]
    container.entries.push({
      path: file.webkitRelativePath.replace(/^[^\/]*\//, ''),
      size: file.size,
      read: function (file) {
        return read(file)
      }.bind(null, file)
    })
  }
  listener(container)
}

function handleDirectoryUpload (input, listener) {
  var container = {
    entries: []
  }

  var iterate = function (entries, path, resolve) {
    var promises = []
    entries.forEach(function (entry) {
      promises.push(new Promise(function (resolve) {
        if ('getFilesAndDirectories' in entry) {
          entry.getFilesAndDirectories().then(function (entries) {
            iterate(entries, entry.path + entry.name + '/', resolve)
          })
        } else {
          if (entry.name) {
            var p = (path + entry.name).replace(/^[\/\\]/, '')
            console.log('directory upload entry:', entry)
            container.entries.push({
              path: p,
              size: entry.size,
              read: function (entry) {
                return read(entry)
              }.bind(null, entry)
            })
          }
          resolve()
        }
      }))
    })
    Promise.all(promises).then(resolve.bind())
  }

  input.getFilesAndDirectories().then(function (entries) {
    new Promise(function (resolve) {
      iterate(entries, '/', resolve)
    }).then(function () {
      listener(container)
    })
  })
}

function onchange (listener, evt) {
  var input = evt.target
  if (input.directory) {
    console.log('Method chosen: directory upload')
    handleDirectoryUpload(input, listener)
  } else {
    var files = input.files
    if (files[0].webkitRelativePath) {
      console.log('Method chosen: webkitfolder')
      handleChromeFolder(files, listener)
    } else if (files.length === 1) {
      var file = files[0]
      if (!ZIP_RE.test(file.name)) {
        throw new Error('Only zip files are supported for now')
      }
      console.log('Method chosen: zip')
      handleZip(file, listener)
    } else {
      throw new Error('Directory upload not supported in your browser, please pick a .zip file.')
    }
  }
}

window.Uppa = function (input, listener) {
  input.onchange = (evt) => onchange(listener, evt)
}
