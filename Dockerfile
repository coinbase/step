FROM golang@sha256:62b42efa7bbd7efe429c43e4a1901f26fe3728b4603cb802248fff0a898b4825

# Install Zip
RUN apt-get update && apt-get upgrade -y && apt-get install -y zip

# Install Dep
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

WORKDIR /go/src/github.com/coinbase/step

COPY Gopkg.lock Gopkg.toml ./

RUN dep ensure -vendor-only

COPY . .

RUN go build && go install

# Use to deploy Lambda
RUN ./scripts/build_lambda_zip
RUN step json -lambda "%lambda%" > state_machine.json

CMD ["step"]
