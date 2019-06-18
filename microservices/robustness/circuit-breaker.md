Config Circuit Breaker for Microservices
Author(s): Aaron Zhuo
Approvers: [Approver's Name should be someone off the team]
Implemented in: <Stormbreaker (Envoy); Go; Typescripts>
TL;DR
This document talks about configuring circuit breakers for microservices to prevent them from being hammered when they are not healthy.  
Background
Circuit Breakers
The definition of “circuit breaker design pattern” from Wikipedia: “Circuit breaker is a design pattern used in modern software development. It is used to detect failures and encapsulates the logic of preventing a failure from constantly recurring, during maintenance, temporary external system failure or unexpected system difficulties.”
The above picture shows all the potential states of a circuit breaker:
A circuit breaker is `Closed` (default state) when nothing goes wrong. A circuit breaker does nothing but just let requests go through when its state is closed. Moreover, If a request is successful or failed under a certain threshold, the state will remain the same. 
A circuit breaker becomes `Open` when the failed requests reach a certain threshold. With the state `Open`, all the incoming requests are marked as failed with error `Circuit Open`, without hitting the real service.
A circuit breaker is `Half open` when it makes an attempt to detect whether the system has recovered. If yes, the circuit breaker will become `Closed` or remain at the `Open` state.
Use Case

The above picture shows that a circuit breaker can help in two aspects:
It prevents the service it is protecting from being hammered. For example, the circuit breaker of Service C will capture all the requests and marks them failed without bothering Service C when it is dealing with some kind of outrage.
The fail-fast mechanism helps the whole system get rid of slow time out problems form downstream services to upstream services. Sometimes time out brings more hurt to the system than any other problems. With the fail-fast mechanism, an cascading problem will still happen but in a quick way. Moreover, a circuit breaker can provide a fullback, for example, a downgrading function, to allow you to elegantly deal with downstream service errors. 
Our Problems
The major problem is we have not applied any kind of circuit breaker to protect our microservices.
Related Proposals:
A list of proposals this proposal builds on or supersedes.
Proposal
Overview
The following pictures demonstrate a scenario where Account Group service and a user (a browser) talks to Accounts service through its Go and typescript SDK. All the requests hit Stormbreaker (Envoy Proxy wrapper) and get redirected to Accounts Deployment.
There are two types of circuit breakers we can apply in our microservices.

Envoy Circuit Breaker
Envoy proxy provides a circuit breaker which allows you to control the maximum connects (HTTP/1.1) or the maximum requests (HTTP/2) that each of our microservices can accept.
For example, if a microservice has 3 Pods and each Pod can handle 1000 requests per second. Then it makes sense to set the maximum requests to 6000 and config a auto-scaler which can scale the replica of the service to 6.
Application Layer Circuit Breaker
As shown in the above picture, an application layer circuit breaker lives in the SDKs in provides end-to-end projection. I wrote a demo to demonstrate how to apply a circuit breaker in Acccounts Go SDK using hystrix-go. Here is the major code:

// NewClient returns a new accounts API client object
func NewClient(ctx context.Context, e config.Env, dialOptions ...grpc.DialOption) (Interface, error) {
  address := addresses[e]
  if address == "" {
     return nil, fmt.Errorf("failed to create client with environment %d", e)
  }
  connection, err := vax.NewGRPCConnection(ctx, address, e != config.Local, scopes[e], true, dialOptions...)
  if err != nil {
     return nil, err
  }

  // Config hystrix command
  hystrix.ConfigureCommand("circuit_breaker_demo", hystrix.CommandConfig{
     Timeout:               10000,  // how long to wait for command to complete, in milliseconds.
     MaxConcurrentRequests: 100,       // The maximum concurrent requests this service can send to Accounts service.
     RequestVolumeThreshold: 20,       // The minimum number of requests needed before a circuit can be tripped due to health.
     ErrorPercentThreshold: 50,    // This setting causes circuits to open once the rolling measure of errors exceeds this percent of requests.
     SleepWindow: 3000,           // How long, in milliseconds, to wait after a circuit opens before testing for recovery.
  })


  return &client{accounts_v1.NewAccountsServiceClient(connection)}, nil
}

The usage of the circuit breaker:

// Get returns just the account for the specified input parameters
func (c *client) Get(ctx context.Context, businessID, appID, partnerID string) (*Account, error) {
  var resp *accounts_v1.GetResponse

  err := hystrix.Do("circuit_breaker_demo", func() error {
     err := // Call Accounts service to get the data.
     if err != nil {
        return err
     }
     return nil

  }, func(err error) error{
     return fmt.Errorf("this is a fallback example, err %s", err.Error())
  })
  if err != nil {
     return nil, err
  }

  return resp.Account, nil
}

From the code, you can see that the circuit breaker is triggered by `ErrorPercentThreshold` and it allows you to provide a fullback function when the downstream is down. 

Each language basically has its own circuit breaker implementation. Since we are using GO and typescripts to write most of our SDKs, we can apply this circuit breaker to our TS SDks and hystrix-go to our Go SDKs.
Rationale
[A discussion of alternative approaches and the tradeoffs, advantages, and disadvantages of the specified approach.]
Implementation
Add the support of Envoy circuit breakers in mscli and microservice.yaml.
Update the version of mscli in mission control.
Ask each team to config Envoy circuit breaker for the microservices they are managing based on what they need.
Add circuit breaker support in typescript SDK auto generator.
Add circuit breaker support in Go SDK auto generator, or each team may have to do this manually if they think it is important to do this.
Open issues (if applicable)
[A discussion of issues relating to this proposal for which the author does not know the solution. This section may be omitted if there are none.]
Reference
Circuit Breaker Design Pattern
Circuit Breaker and Retry
Envoy Proxy Home
Circuit Breaking In Envoy Proxy
Automatic Retries in Envoy Proxy
Hystrix in Go
Hystrix-like Circuit Breaker for JavaSctipt.

