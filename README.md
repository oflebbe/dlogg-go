# dlogg-go

Support for the reading logs from the UVR1611 and UVR61-3 Control Units 
for thermal solar engines.

Tested with the USB Datalogger on an UVR1611 in 1DL mode.

The protocol is reengineered from the source code of the https://github.com/fb/dlogg-linux repository.
Redone because it is too inflexible to be used as an backend for further processing.

The project should work on any operating system.

# Installation
```
go get github.com/oflebbe/dlogg-go@latest
```


# Acknoledgements
Acknowledgement to H. Römer for creating and maintaing dlogg-linux.



