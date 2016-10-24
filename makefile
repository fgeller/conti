ARTIFACT = contd

build:
		CGO_ENABLED=0 go build -o ${ARTIFACT} .

image: TAG ?= latest
image:
		GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ${ARTIFACT} -a .
		docker build -t ${ARTIFACT}:$(TAG) .

clean:
		rm -f ${ARTIFACT}
