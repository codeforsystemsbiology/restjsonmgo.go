include $(GOROOT)/src/Make.inc

TARG=restmgo
GOFILES=\
	mongodb.go\
	restjson.go\
	main.go\

include $(GOROOT)/src/Make.cmd
