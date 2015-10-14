# fullerite

[![Build Status](https://travis-ci.org/baris/fullerite.svg?branch=master)](https://travis-ci.org/baris/fullerite)


*Fullerite is a metrics collection tool*. It is different than other collection tools (e.g. diamond, collectd) in that it supports multidimensional metrics from its core. It is also meant to innately support easy concurrency. Collectors and handler are sufficiently isolated to avoid having one misbehaving component affect the rest of the system. Generally, an instance of fullerite runs as an upstart daemon on a box collecting the configured metrics and reports them via different handlers to endpoints such as graphite, signalfx, or datadog. 

A summary of interesting features of fullerite include:
 * Fully compatible with diamond collectors
 * Written in Go for easy reliable concurrency
 * Configurable set of handlers and collectors
 * Native support for dimensionalized metrics
 * Internal metrics to track handler performance

Fullerite is also able to run [Diamond](https://github.com/python-diamond/Diamond) collectors natively. This means you don't need to port your python code over to Go. We'll do the heavy lifting for you.

## success story
  * Running on 1,000s of machines
  * Running on AWS and real hardware all over the world
  * Running 8-12 collectors and 1-2 handlers at the same time
  * Emitting over 5,000 metrics per flush interval on average per box
  * Well over 10 million metrics per minute

## how it works
Fullerite works by spawning a separate goroutines for each collector and handler then acting as the conduit between the two. Each collector and handler can be individually configured with a nested JSON map in the configuration. But sane defaults are provided. 

The `fullerite_diamond_server` is a process that starts each diamond collector in python as a separate process. The listening collector in go must also be configured on. Doing this each diamond collector will connect to the server and then start piping metrics to the collector. The server handles the transient connections and other such issues by spawning a new goroutine for each of the connecting collectors. 

![Alt text](/fullerite_arch.tiff?raw=true "Optional Title")

## supported collectors
 * [fullerite collectors](src/fullerite/collector)
 * [diamond collectors](src/diamond/collectors)

## supported handlers
 * [Graphite](http://graphite.wikidot.com/)
 * [Kairos](https://github.com/kairosdb/kairosdb)
 * [SignalFx](https://www.signalfx.com)
 * [Datadog](https://www.datadoghq.com)
