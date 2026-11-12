## The Problem

* Well, when I first started working with telemetry, I relied heavily on Alloy. It was nice, simple to configure (besides the eldritch syntax of the config file), and it worked well. But I noticed a problem: it was heavily tied to Grafana. That's not a bad thing, but nowadays the market offers many observability tools, such as Datadog, Splunk, and others. I thought I was good at what I do, but reading about observability tools that were not Grafana made me realize that this could leave me without a chance to apply for jobs in the field. That sounded hellish, so I decided to change my approach. The OpenTelemetry Collector is also much easier to understand and configure.

## Underlying Concept / How it Works
* Well, if you have ever worked with Alloy, you might find it fairly similar. However, as already stated, the OpenTelemetry Collector is much easier to understand and configure, and it is also much more flexible with other observability tools. It is not "locked to Grafana" like Alloy. Let me explain what the OpenTelemetry Collector does. As always, I will divide it into phases so it is easier to understand:
* **Phase 1** -> Just like with Alloy, the first thing we do is explicitly define a receiver. This tells our Collector that it should listen on a specific port, such as the idiomatic `:4318` port for OTLP HTTP. We do this to let the Collector know that it should receive telemetry data from the OTel API.
* **Phase 2** -> Here comes the interesting part: we need to define a processor that tells our Collector how it should process the telemetry data. That is where the term "processor" comes from. Think of processors as instructions that tell the Collector how to handle the data, including how to distribute it to other vendors and when to stop or time out. Here, we define the `batch` processor and the `memory_limiter` processor, which batch the telemetry data and limit the Collector's memory usage, respectively.
* **Phase 3** -> Next, we define the exporters. The syntax is simple: we declare the exporter name, such as `otlphttp/loki`, and then provide its settings, such as the endpoint, TLS settings, headers, and so on.
* **Phase 4** -> Unlike Alloy, the OpenTelemetry Collector lets us define multiple pipelines. Pipelines tell the Collector where to route each type of data. For example, we define the `logs` pipeline to route logs to Loki. Inside it, we specify the `receivers`, `processors`, and `exporters` that we want to use. We repeat the same process for the `traces` pipeline and, in the future, the `metrics` pipeline.

## Workflow / Lifecycle (Diagram)
//

## Example Code and Explanation
```yaml
receivers:
  otlp:
    protocols:
      http:
        endpoint: 0.0.0.0:4318

processors:
  memory_limiter:
    check_interval: 1s
    limit_percentage: 80
    spike_limit_percentage: 25
  batch:
    send_batch_size: 8192
    timeout: 2s
    send_batch_max_size: 16384

exporters:
  otlphttp/loki:
    endpoint: http://loki:3100/otlp
    tls:
      insecure: true
    retry_on_failure:
      enabled: true
      initial_interval: 3s
      max_interval: 10s
  otlphttp/tempo:
    endpoint: http://tempo:4318
    tls:
      insecure: true
    retry_on_failure:
      enabled: true
      initial_interval: 3s
      max_interval: 10s

service:
  pipelines:
    logs:
      receivers: [otlp]
      processors: [memory_limiter, batch]
      exporters: [otlphttp/loki]
    traces:
      receivers: [otlp]
      processors: [memory_limiter, batch]
      exporters: [otlphttp/tempo]
```
This configuration describes the complete route that telemetry takes through the Collector:
1. **Receiver** - `otlp` tells the Collector to accept OTLP data over HTTP. The Collector listens on `0.0.0.0:4318`, so it accepts connections on port `4318` from the network interfaces available inside its container or host.
2. **Processors** - processors prepare data before it is exported.
   * `memory_limiter` prevents the Collector from using unlimited memory. It checks memory every second, starts limiting data at 80% usage, and reserves 25% for sudden memory spikes.
   * `batch` groups telemetry together before sending it. This reduces the number of outgoing requests. A batch is sent when it reaches `8192` items or when the `2s` timeout expires, and it cannot grow beyond `16384` items.
3. **Exporters** - exporters describe where the processed data goes. The name `otlphttp/loki` means "use the OTLP over HTTP exporter, configured for Loki." The part before `/` is the exporter type; the part after `/` is simply a name that lets the configuration refer to that particular instance.
   * The Loki exporter sends logs to `http://loki:3100/otlp`. As mentioned earlier, you should not guess the endpoint. Check the documentation for the vendor you are using and find the correct endpoint for OTLP ingestion.
   * The Tempo exporter sends traces to `http://tempo:4318`.
   * `tls.insecure: true` is appropriate for this local Docker network, where these connections do not use TLS. A production deployment should configure TLS instead of copying this setting blindly.
   * `retry_on_failure` makes the Collector retry temporary exporter failures, starting with a 3-second delay and increasing the delay up to 10 seconds.
4. **Pipelines** - pipelines connect receivers, processors, and exporters for a specific telemetry signal. Both pipelines read from the OTLP receiver and use the same processors, but they send the data to different backends:
   * `logs` sends logs to Loki.
   * `traces` sends traces to Tempo.

The flow is therefore:

```text
OTel -> OTLP receiver -> memory limiter -> batch processor
        -> Loki  (logs)
        -> Tempo (traces)
```

The Collector does not decide whether data is a log or a trace based on the exporter name. The `logs` and `traces` pipeline types select the signal first, and each pipeline then forwards that signal to its configured exporter.

## What I Still Don't Understand
//