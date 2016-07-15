#!/usr/bin/env ruby

require 'json'

metrics = {}


dimensions = {"dim1" => "val1"}

metrics['first'] = {
    "name" => "example",
    "value" => 2.0,
    "dimensions" => dimensions,
    "type"=> "gauge"
}

metrics['second'] = {
    "name" => "counter2.example",
    "value" => 2.0,
    "dimensions" => dimensions,
    "type" => "cumcounter"
}

puts metrics.values.to_json
