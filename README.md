# Step (Beta)

<img align="right" src="./assets/one_small_step_for_gopher.png" alt="One Small Step for Go"/>

Step is a opinionated implementation of the [AWS State Machine language](./STATE_SPEC.md) in [Go](https://golang.org/) used to build and test [AWS Step Functions](https://docs.aws.amazon.com/step-functions/latest/dg/getting-started.html) and [Lambdas](https://docs.aws.amazon.com/lambda/latest/dg/getting-started.html). Step combines the **Structure** of a state machine with the **Code** of a lambda so that the two can be developed, tested and maintained together.

The three core components of Step are:

1. **Library**: tools for building and deploying Step Functions in Go.
2. **Implementation**: of the AWS State Machine specification to test with the code together ([README](./machine)).
3. **Deployer**: to deploy Lambda's and Step Functions securely ([README](./deployer))

### Getting Started

A Step function has two parts:

1. A **State Machine** description in JSON, which outlines the flow of execution.
2. The **Lambda Function** which executes the `Task` states of the step function.

*Step uses only one Lambda per State Machine making it easier to test and maintain.*

**All following examples come from [`coinbase/step-hello-world`](https://github.com/coinbase/step-hello-world)**

Create a State Machine like this:

```go
func StateMachine(lambdaArn string) (machine.StateMachine, error) {
  state_machine, err := machine.FromJSON([]byte(`{
    "Comment": "Hello World",
    "StartAt": "HelloFn",
    "States": {
      "HelloFn": {
        "Type": "Pass",
        "Result": "Hello",
        "ResultPath": "$.Task",
        "Next": "Hello"
      },
      "Hello": {
        "Type": "Task",
        "Comment": "Deploy Step Function",
        "End": true
      }
    }
  }`))

  if err != nil {
      return nil, err
  }

  // Set the "Hello" Task Handler to be executed on that state
  state_machine.SetResourceFunction("Hello", HelloHandler)

  // Set Lambda Arn to call with Task States
  state_machine.SetResource(&lambdaArn)

  return state_machine, nil
}
```

When a `Task` state is reached, the handler assigned to that state will be executed. To know what Handler to execute the state machine must include a `Pass` state before the `Task` that injects the `Task` name into `$.Task`. 

Above you can see we set the Handler to the `"Hello"` task to be `HelloHandler`. All Handlers must implement the type `func(context.Context, interface{}) (interface{}, error)` where the interfaces can be arbitrary Structs.

`HelloHandler` is defined like:

```go
type Hello struct {
  Greeting *string
}

// HelloHandler takes a Hello struct alters its greeting and returns it
func HelloHandler(_ context.Context, hello *Hello) (*Hello, error) {
  if hello.Greeting == "" {
    hello.Greeting = "Hello World"
  }
  return hello, nil
}
```

To build a Step Function we then need an executable that can:

1. Be executed in a Lambda
2. Build the State Machine

```go
func main() {
  var arg, command string
  switch len(os.Args) {
  case 1:
    fmt.Println("Starting Lambda")
    run.Lambda(StateMachine("lambda"))
  case 2:
    command = os.Args[1]
    arg = ""
  case 3:
    command = os.Args[1]
    arg = os.Args[2]
  default:
    printUsage() // Print how to use and exit
  }

  switch command {
  case "json":
    run.JSON(StateMachine(arg))
  case "exec":
    run.Exec(StateMachine(""))(&arg)
  default:
    printUsage() // Print how to use and exit
  }

}
```

1. `./step-hello-world` will run as a Lambda Function
2. `./step-hello-world json` will print out the state machine

As a bonus command `exec` will execute the State Machine, e.g. `./step-hello-world exec` returns:

```bash
{
  "Greeting": "Hello World"
}
```

and running `./step-hello-world exec '{"Greeting": "Hi"}'` returns:

``` bash
{
 "Greeting": "Hi"
}
```

### Testing

A core benefit when using Step and joining the State Machine and Lambda together is that it makes it possible to test your Step Functions execution.

For example, a basic test that ensures the correct output and execution path through the Hello World step function looks like:

```go
func Test_HelloWorld_StateMachine(t *testing.T) {
  state_machine, err := StateMachine("")
  assert.NoError(t, err)

  exec, err := state_machine.Execute(&Hello{})
  assert.NoError(t, err)
  assert.Equal(t, "Hello World", exec.Output["Greeting"])

  assert.Equal(t, state_machine.Path(), []string{
    "HelloFn",
    "Hello",
  })
}
```

### Deploying

There are two ways to get a State Machine into the cloud:

1. **Bootstrap**: Directly upload the Lambda and Step Function to AWS
2. **Deploy**: Using the Step Deployer which is a Step Function included in this library.

The Step executable can perform both of these functions.

*Step does not create the Lambda or Step Function in AWS, it only modifies them. So before either bootstrapping or deploying the resources must already be created.*

First build and install step with:

```bash
go build && go install
```

Bootstrap (directly upload to the Step Function and Lambda):

```bash
# Use AWS credentials or assume-role into AWS
# Build linux zip for lambda
GOOS=linux go build -o lambda
zip lambda.zip lambda

# Tell step to bootstrap this lambda
step bootstrap                        \
  -lambda "coinbase-step-hello-world" \
  -step "coinbase-step-hello-world"   \
  -states "$(./step-hello-world json)"
```

Deploy (via the step-deployer step function):

```bash
GOOS=linux go build -o lambda
zip lambda.zip lambda

# Tell step-deployer to deploy this lambda
step deploy                           \
  -lambda "coinbase-step-hello-world" \
  -step "coinbase-step-hello-world"   \
  -states "$(./step-hello-world json)"
```

### Development State

Step is still very Beta and its API might change quickly.

### More Links

1. https://docs.aws.amazon.com/step-functions/latest/dg/step-functions-dg.pdf
1. https://github.com/vkkis93/serverless-step-functions-offline
1. https://github.com/totherik/step

*CC Renee French for the logo, borrowed from GopherCon 2017*
