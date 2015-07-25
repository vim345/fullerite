# fullerite
A metrics collection tool. It is different than other collection tools (e.g. diamond, collectd) in that it supports multidimensional metrics from its core. It is also meant to innately support easy concurrency. Collectors and handler are sufficiently isolated to avoid having one misbehaving component effect the rest of the system. 

The other big goal is to support diamond collectors natively. This means you don't need to port you python code over to GO. We'll do the heavy lifting. 

# supported handlers
We will support graphite (no dimensions :( :( :( ) and signalfx (yay dimensions :) :) :) ). As dimensions are cool, signalfx is going to get most of the love first.

