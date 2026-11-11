## The Problem
* Since we already learned alloy and how to apply it, we need another tool, a way to collect the logging that OTel been piling up. As example, since OTel is piling up logging at it's API, using an exporter to tell that it needs to send the logging it's currently batching to that port (The port where Alloy sits, which is `:4318`) - but we need to setup out collector (Alloy), to actually know that there is a vendor that been piling up loggings and want to send it's data to it. The main problem is, we cannot just say "OTel, take and send all you get to Grafana" - nuh uh, we gotta tell it to send the datas to a specialized vendor, as example:
* * **Loki** -> Specialized on Logs, it will structure and display logs via the Grafana UI
* * **Tempo** -> Specialized on Traces, it will structure and display tarces by the spans and use them via Grafan UI
* * **Prometheus** -> Specialized on Metrics, it will structure and display metrics via the Grafana UI
* Though, but before they reach there, we gotta distribute them somehow, and here is wher Alloy comes in, it takes datas from OTel and sends to the vendor we precised

## Underlying Concept / How It Works
* I will divide everything on phases, so it will be easier to follow.
* **Phase 1** -> First thing that we need to do, is to use docker compose (as starter, later you'll learn multiple stuffies as Kubernetes or maybe even Cloud) to make Alloy and the vendors you want run normally, without it, you won't be able to run them. You simply need to enter and find documentation about each, then copy and paste it (preferably to understand what you're doing)
* **Phase 2** -> Then, after the compose, we need to create a file named `config.alloy` (gotta have the extension `.alloy`) - after this, starting inside this file, we declare all the upcoming phases (be wary, Tempo needs a configuration file too)
* **Phase 3** -> Inside Alloy's configuration file, we simply declare for it to listen on port `:4318` - why? Lemme use an analogy -> Think of Alloy as a big fortres, then think about OTel as the deliverer with the mail, in the code, we write `	exporter, err := otlploghttp.New(ctx)` - which simply will search for an environmental variable named "`OTEL_EXPORTER_OTLP_ENDPOINT`" - here we will write stuff as `http://alloy:4318` - this simply tells to OTel: "Hey, OTel, do take your mail (the data after all the process) and send them to the fortress's gates (Alloy's port)". But there is a nuance, just because OTel arrived with the mail at the gates, doesn't mean that Alloy must open them (because it's a secure fortress), that's why we manually write inside it's configuration files "`otelcol.receiver.otlp` (with the HTTP port)" to tell it, do open the gates of the port "`:4318`", since a deliverer (OTel) is going to come with a mail (the data).
* **Phase 4** -> Then, just like in the code, we create a batch, that will take the requests every 'x' - but there's a difference:
* * **Batching OTel in code** -> It will createa a batching prcoess that instead of bombarding the OTel API with hundreds of requests containing datas, it will take the piled up datas and send a batch (from where the term comes) of them every 'x' milliseconds/seconds - cutting latency, because it won't send 1000 requests per 1000 data, but a single request with the 1000 datas
* * **Batching Alloy in config** -> This one applies to Alloy, not OTel anymore - this is how Alloy will interact with other vendors (Loki, Tempo) - we simply write `otelcol.processor.batch` and it will simply work as already stated, we do it, because without it, our Alloy will probably get rate-limited for sending too many requests, while with the batch, it will send a "batch" of them every 'x' time to the vendors
* **Phase 5** -> This is the easiest part, here we precise on which ports to send the data, it's not a "guess where" simulator, so you gotta check documentation and see where exactly they expect to take datas, Loki defaults to the old way (endpoint) -> `http://loki:3100/otlp` - while Tempo uses the global port instead `http://tempo:4318` (yes, because 4318 is no predefined to a single process listening, multiple can expect there).

## Example Code and Explanation (DISCLAIMER: Claude was used for structuring and phrasing this part)
* I'll show the whole `config.alloy` first, then break it down block by block. Think of it as a mail route: the mail enters at the gate, goes through the sorting room, and leaves with a courier. Every block is written as `type "label"`, and the `output` lines are the wires between blocks.
* **`config.alloy`:**
```alloy
otelcol.receiver.otlp "default" {
	http {
		endpoint = "0.0.0.0:4318"
	}

	output {
		logs   = [otelcol.processor.batch.default.input]
		traces = [otelcol.processor.batch.default.input]
	}
}

otelcol.processor.batch "default" {
	output {
		logs   = [otelcol.exporter.otlphttp.loki.input]
		traces = [otelcol.exporter.otlphttp.tempo.input]
	}
}

otelcol.exporter.otlphttp "loki" {
	client {
		endpoint = "http://loki:3100/otlp"
	}
}

otelcol.exporter.otlphttp "tempo" {
	client {
		endpoint = "http://tempo:4318"
	}
}
```
* Here is how the data travels:
```plaintext
Go app (OTel exporter)
   |
   |  http://alloy:4318
   v
otelcol.receiver.otlp "default"        <- the gates
   |  logs + traces
   v
otelcol.processor.batch "default"      <- the sorting room
   |
   |_____ logs   -> otelcol.exporter.otlphttp "loki"   -> loki:3100/otlp
   |_____ traces -> otelcol.exporter.otlphttp "tempo"  -> tempo:4318
```
* **Explanation** -> Let's go block by block, in the order the data travels.
* **`otelcol.receiver.otlp "default"`** -> This is the gate we talked about in Phase 3. `http { endpoint = "0.0.0.0:4318" }` tells Alloy to open port `4318` and accept OTLP over HTTP. We write `0.0.0.0` instead of `localhost` because of Docker. `localhost` inside the Alloy container means "only things inside this same container", and our Go app lives in a different one, so it would knock on the gate forever. `0.0.0.0` means "accept from anyone on the network". The `output` block says where to hand the mail after it comes through the gate. We send `logs` and `traces` to the batch processor, so we write the **address** of that block, `otelcol.processor.batch.default.input`. Read it as `<block type>.<label>.input`: the type, the label we gave it, and its mailbox. We didn't wire `metrics`, so if the app ever sends them, Alloy has nowhere to hand them and they get dropped.
* **`otelcol.processor.batch "default"`** -> This is the batching from Phase 4, the Alloy side of it. Instead of making a separate trip to Loki for every single log line, it piles the data up and sends one batch every `x` time. The defaults are fine for starting out (you can tune them later with `timeout` and `send_batch_size`). The important part is its `output`: this is where the distribution happens. Logs go to the Loki exporter and traces go to the Tempo exporter, so one processor splits both kinds of mail to the right vendors.
* **`otelcol.exporter.otlphttp "loki"`** -> This is the courier that delivers to Loki. The label `"loki"` is just a name we picked, and it's why we can have two blocks of the same type (`"loki"` and `"tempo"`) without them clashing. The `client { endpoint }` part is the delivery address. And here's the nuance that got us: **we give it the base URL, not the full path.** The exporter adds the rest itself (`/v1/logs` for logs, `/v1/traces` for traces), so our final address is `http://loki:3100/otlp/v1/logs`. This is also why the old `/loki/api/v1/push` doesn't work here: that's Loki's own native format, not OTLP.
* **`otelcol.exporter.otlphttp "tempo"`** -> Same courier, different destination. `http://tempo:4318` becomes `http://tempo:4318/v1/traces`. You may wonder: "Wait, Alloy listens on `4318` and now Tempo too, no clash?" None, because `alloy:4318` and `tempo:4318` are two different containers, so two different addresses. But remember that Tempo only listens if you tell it to: its own config file (Phase 2) must have the OTLP HTTP receiver enabled, or the courier will arrive at a closed door.
* One last thing to notice is that the Go side does the same trick. When we set `OTEL_EXPORTER_OTLP_ENDPOINT=http://alloy:4318`, the Go exporter also adds `/v1/logs` on its own. So the rule is the same everywhere: **OTLP exporters get a base URL and add the path themselves.** And when you need a new vendor later, you add an exporter block and point an `output` at its `.input`. That's all the wiring there is.