build:
	go build

run: build
	./mio

install_deps:
	go get gopkg.in/go-playground/validator.v8
	go get github.com/paypal/gatt
