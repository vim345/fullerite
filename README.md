# fullerite

[![Build Status](https://travis-ci.org/baris/fullerite.svg?branch=master)](https://travis-ci.org/baris/fullerite)

A metrics collection tool. It is different than other collection tools (e.g. diamond, collectd) in that it supports multidimensional metrics from its core. It is also meant to innately support easy concurrency. Collectors and handler are sufficiently isolated to avoid having one misbehaving component effect the rest of the system. 

fullerite is also able to run [Diamond](https://github.com/python-diamond/Diamond) collectors natively. This means you don't need to port your python code over to Go. We'll do the heavy lifting for you.

# supported collectors
 * [fullerite collectors](src/fullerite/collector)
 * [diamond collectors](src/diamond/collectors)

# supported handlers
 * [Graphite](http://graphite.wikidot.com/)
 * [Signalfx](https://www.signalfx.com)
 * [Datadog](https://www.datadoghq.com)


