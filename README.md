# log2file

`log2file` reads from standard input and write to the file, reopening its file
handle if the file is renamed or deleted.

# Build

```
GOOS=linux go build log2file.go
```

# Releases

If you tag a release then Travis CI will automatically add the binary for 64-bit
Linux to it.
