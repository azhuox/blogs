# Robust

## Kubernetes Layer

### Readness Probe & Liveness Probe

What we have:

```
livenessProbe:
  httpGet:
    path: /healthz
    port: 11001
    scheme: HTTP
  initialDelaySeconds: 0
  periodSeconds: 10
  timeoutSeconds: 1
  successThreshold: 1
  failureThreshold: 3
readinessProbe:
  httpGet:
    path: /healthz
    port: 11001
    scheme: HTTP
  initialDelaySeconds: 0
  periodSeconds: 10
  timeoutSeconds: 1
  successThreshold: 1
  failureThreshold: 3
```

What does it mean:

Take Liveness as example:

- The `httpGet` field means the liveness check is performed by calling `http//127.0.0.1:11001/healthz`. Any status code greater than or equal to 200 and less than 400 indicates success. Any other code indicates failure.
- The `initialDelaySeconds: 0` field means delay 0 second before performing the first check.
- The `periodSeconds: 10` field means perform the check every 10 seconds.
- The `timeoutSeconds: 1` field means the timeout of each liveness check is 1 second.
- The `successThreshold: 1` field means the liveness probe is considered successful as long as at least once check succeeds.
- The `failureThreshold: 3` means kubernetes will try 3 more times before giving up when the live check is failed. Giving up in case of liveness probe means restarting the Pod. In case of readiness probe the Pod will be marked Unready.

### Gracefully shut down


Example:

response Status: 200 OK
response Body: {"target":99.9,"pointList":[{"timeStamp":"2019-06-10T16:48:20Z","value":100},{"timeStamp":"2019-06-10T16:48:40Z","value":100},{"timeStamp":"2019-06-10T16:48:50Z","value":100},{"timeStamp":"2019-06-10T16:49:00Z","value":100},{"timeStamp":"2019-06-10T16:49:10Z","value":100},{"timeStamp":"2019-06-10T16:49:20Z","value":100},{"timeStamp":"2019-06-10T16:49:30Z","value":100},{"timeStamp":"2019-06-10T16:50:00Z","value":100},{"timeStamp":"2019-06-10T16:50:10Z","value":100},{"timeStamp":"2019-06-10T16:50:20Z","value":100},{"timeStamp":"2019-06-10T16:50:30Z","value":100},{"timeStamp":"2019-06-10T16:50:40Z","value":100},{"timeStamp":"2019-06-10T16:50:50Z","value":100},{"timeStamp":"2019-06-10T16:51:00Z","value":100},{"timeStamp":"2019-06-10T16:51:10Z","value":100},{"timeStamp":"2019-06-10T16:51:20Z","value":100},{"timeStamp":"2019-06-10T16:51:30Z","value":100},{"timeStamp":"2019-06-10T16:51:40Z","value":100},{"timeStamp":"2019-06-10T16:51:50Z","value":100},{"timeStamp":"2019-06-10T16:52:00Z","value":100},{"timeStamp":"2019-06-10T16:52:10Z","value":100},{"timeStamp":"2019-06-10T16:52:20Z","value":100},{"timeStamp":"2019-06-10T16:52:30Z","value":100},{"timeStamp":"2019-06-10T16:52:40Z","value":100},{"timeStamp":"2019-06-10T16:52:50Z","value":100},{"timeStamp":"2019-06-10T16:53:00Z","value":100}]}
response Status: 503 Service Unavailable
response Body: upstream connect error or disconnect/reset before headers
response Status: 503 Service Unavailable
response Body: upstream connect error or disconnect/reset before headers
response Status: 503 Service Unavailable
response Body: upstream connect error or disconnect/reset before headers
response Status: 503 Service Unavailable
response Body: upstream connect error or disconnect/reset before headers
response Status: 503 Service Unavailable
response Body: upstream connect error or disconnect/reset before headers
response Status: 503 Service Unavailable
response Body: upstream connect error or disconnect/reset before headers



## Application Layer