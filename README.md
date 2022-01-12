# GOCOMPRESS
A go cli implementation of the Lz77 algorithm. **This is just a quick and dirty implementation for learning purposes expect bugs and ineffeciencies**

### Building
Assuming you have go installed just run ``go build gocomp.go`` there are no dependencies 

### Usage:
To compress a file:
```
gocomp -i inputfile -o outputfile
```
To decompress a file:
```
gocomp -d -i inputfile -o outputfile
```


### Known issues / limitations :
- It seems that some files only the first few chunk gets compressed and the rest filled with null bytes
- Quite slow
