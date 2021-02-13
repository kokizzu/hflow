
.PHONY: start build test stop ca-export exec

build :
	@-rm -f ./bin/hflow
	@go build -o ./bin/hflow

test :
	@go test -cover -count=1 ./...

bench:
	@go test -run=XXX -bench=. -benchtime=5s ./cert/

install: build
	@-sudo rm -f /usr/local/bin/hflow
	@sudo cp ./bin/hflow /usr/local/bin/hflow && sudo chmod +X /usr/local/bin/hflow 

URL_PATTERN=""
STATUS_PATTERN=""

start : stop build 
	@./bin/hflow -v=3 -l=100 -u=${URL_PATTERN} -s=${STATUS_PATTERN} 2> ./bin/hflow.log 1> ./bin/hflow.capture &

stop :
	-@pkill hflow

ca-export : stop build 
	@-mkdir ./bin 2> /dev/null
	@./bin/hflow -ca > ./bin/hflow-ca-export.pem

exec : stub start
	@echo "hello"
#__________________________________________________________________________________________________________________________
#
# Stub server related targets
#__________________________________________________________________________________________________________________________

.PHONY: stub stub-stop 

stub-stop: 
	-@pkill stub

stub : stub-stop
	@-rm ./cmd/stub/stub
	@go build -o ./cmd/stub/stub ./cmd/stub/
	@./cmd/stub/stub -v=2 2> ./bin/stub.log &
	@rm ./cmd/stub/stub
	@echo "add '127.0.0.1 stub-server' to your host file"

#__________________________________________________________________________________________________________________________
#
# PKI related targets. If updating pki files for use in hflow, ensure variables in pem.go files are updated with output
#__________________________________________________________________________________________________________________________

.PHONY: ca cert

ca :
	@./scripts/gen_ca.sh

cert : # ex: `make cert DOMAIN="stub-server"`. ca certs must be present, if not, run `make ca`
	@./scripts/gen_cert.sh ${DOMAIN}