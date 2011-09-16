include $(GOROOT)/src/Make.inc

TARG=golem
GOFILES=\
	mongodb.go\
	restjson.go\
	main.go\

include $(GOROOT)/src/Make.cmd
