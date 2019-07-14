#INSTALL
cd $GOPATH
go get -d -v ./..
go build ./<name_project>

#RUN ENV

ENV=<ENV> PORT=<PORT> DEBUG=<BOOLEAN> ./<name_project>

#######