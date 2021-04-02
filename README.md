# kosyncsrv
a tiny koreader sync server rewritten by golang according to
[koreader-sync](https://github.com/myelsukov/koreader-sync),
it uses the sqlite3 file as the database by default, tables will be auto created while the programm runs

# build and run
if you are using the newer go version wiht modle
```
go mod init kosyncsrv
go build
```
run:
```
kosyncsrv [-h] [-t 127.0.0.1] [-p 8080] [-ssl -c "./cert.pem" -k "./cert.key"]
```



