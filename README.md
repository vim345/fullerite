# fullerite

[![Build Status](https://travis-ci.org/baris/fullerite.svg?branch=master)](https://travis-ci.org/baris/fullerite)

A metrics collection tool. It is different than other collection tools (e.g. diamond, collectd) in that it supports multidimensional metrics from its core. It is also meant to innately support easy concurrency. Collectors and handler are sufficiently isolated to avoid having one misbehaving component effect the rest of the system.

fullerite is also able to run [Diamond](https://github.com/python-diamond/Diamond) collectors natively. This means you don't need to port your python code over to Go. We'll do the heavy lifting for you.

## supported collectors
 * [fullerite collectors](src/fullerite/collector)
 * [diamond collectors](src/diamond/collectors)

## supported handlers
 * [Graphite](http://graphite.wikidot.com/)
 * [Kairos](https://github.com/kairosdb/kairosdb)
 * [SignalFx](https://www.signalfx.com)
 * [Datadog](https://www.datadoghq.com)

# beatit

A command line tool to test fullerite handlers and metric stores they write to.

    beatit -c test.conf --graphite -l error -t 100 --dps 500 --time 60

Above command runs 100 graphite handlers and tries sending 500 data points per second to each handler for 60 seconds.
