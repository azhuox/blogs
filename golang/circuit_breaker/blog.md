Application Layer Circuit Breakers


What is included in this blog:

1. An introduction of circuit breakers.
2. A demonstration of an application layer circuit breaker.

# What Is A Circuit Breaker

## Introduction


## States

Each circuit breaker has to follow the following specifications:

- A circuit breaker is `closed` when a request is successful or failed under a certain threshold.
- A circuit breaker becomes `open` when the total number of failed requests reaches a certain threshold. Any incoming request will fail fast without even hitting the downstream service it is protecting. This fail-fast mechanism assumes that the downstream service is on fire and stops upstream services from putting more stress to the downstream service.
- A circuit breakers becomes `half-open` when its previous state is `open` and the sleep window (a period of time) has been passed. It then makes a request to check whether the downstream service has recovered. If yes, the circuit breaker will become `closed`. Otherwise it will switch to `open`.



# Application Layer Circuit Breakers

## Overview

[image]

## Demo


# What Is Next

# Reference








