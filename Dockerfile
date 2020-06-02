FROM golang@sha256:ebe7f5d1a2a6b884bc1a45b8c1ff7e26b7b95938a3e8847ea96fc6761fdc2b77

# Install Zip
RUN apt-get update && apt-get upgrade -y && apt-get install -y zip

WORKDIR /go/src/github.com/coinbase/step

ENV GO111MODULE on
ENV GOPATH /go

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build && go install

# builds lambda.zip
RUN ./scripts/build_lambda_zip
RUN shasum -a 256 lambda.zip | awk '{print $1}' > lambda.zip.sha256

RUN mv lambda.zip.sha256 lambda.zip /
RUN step json > /state_machine.json

CMD ["step"]
