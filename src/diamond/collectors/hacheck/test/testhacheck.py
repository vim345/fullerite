#!/usr/bin/python
# coding=utf-8
################################################################################

from mock import Mock
from mock import patch
from test import CollectorTestCase
from test import get_collector_config
from test import unittest
from urllib2 import HTTPError

from diamond.collector import Collector

from hacheck import HacheckCollector

################################################################################


class TestHacheckCollector(CollectorTestCase):
    
    def setUp(self):
        config = get_collector_config('HacheckCollector', {})
        self.collector = HacheckCollector(config, None)

    @patch.object(Collector, 'publish')
    @patch('urllib2.urlopen')
    def test_works_with_real_data(self, urlopen_mock, publish_mock):
        urlopen_mock.return_value = self.getFixture('metrics')
        self.collector.collect()
        self.assertPublishedMany(
            publish_mock,
            {
                'hacheck.cache.expirations': 2692,
                'hacheck.cache.sets': 2713,
                'hacheck.cache.gets': 28460,
                'hacheck.cache.hits': 25747,
                'hacheck.cache.misses': 2713,
                'hacheck.outbound_request_queue_size': 12
            },
        )

    @patch.object(Collector, 'publish')
    @patch('urllib2.urlopen')
    def test_graceful_failure_on_http_error(self, urlopen_mock, publish_mock):
        urlopen_mock.side_effect = HTTPError(
            Mock(), Mock(), Mock(), Mock(), Mock())
        self.collector.collect()
        self.assertPublishedMany(publish_mock, {})

    @patch.object(Collector, 'publish')
    @patch('urllib2.urlopen')
    def test_graceful_failure_on_json_error(self, urlopen_mock, publish_mock):
        urlopen_mock.return_value = self.getFixture('bad_metrics')
        self.collector.collect()
        self.assertPublishedMany(publish_mock, {})


################################################################################
if __name__ == "__main__":
    unittest.main()
