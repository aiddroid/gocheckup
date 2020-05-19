### GoCheckup
A fast multiple protocol checker written by Golang.
For example, you can make following checks:
HTTP、DNS、FTP、SSH、TCP、UDP and etc.

### Usage
```
gocheckup -c checkup.json
```


### configure
configure in file `checkup.json`
```
{
  "checkers": [{
    "type": "http",
    "endpoint_name": "163",
    "endpoint_url": "https://www.163.com",
    "attempts": 5
  },{
    "type": "http",
    "endpoint_name": "qq",
    "endpoint_url": "https://www.qq.com",
    "attempts": 5
  }],
  "storage":{
    "type": "fs",
    "dir": "/tmp",
    "filename": "gocheckup.log"
  }
}
```