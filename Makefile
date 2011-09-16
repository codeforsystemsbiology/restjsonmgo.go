include $(GOROOT)/src/Make.inc

TARG=restjsonmgo
GOFILES=\
	mongodb.go\
	restjson.go\
	main.go\

include $(GOROOT)/src/Make.cmd
