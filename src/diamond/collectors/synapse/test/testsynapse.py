#!/usr/bin/python
# coding=utf-8
################################################################################

from mock import Mock
from mock import patch
from test import CollectorTestCase
from test import get_collector_config
from test import unittest
try:
    from cStringIO import StringIO
except ImportError:
    from StringIO import StringIO

from diamond.collector import Collector

from synapse import SynapseCollector

################################################################################


class TestSynapseCollector(CollectorTestCase):
    
    def setUp(self):
        config = get_collector_config('SynapseCollector', {})
        self.collector = SynapseCollector(config, None)

    @patch.object(Collector, 'publish')
    @patch('synapse.open')
    def test_works_with_real_data(self, open_mock, publish_mock):
        open_mock.return_value = self.getFixture('synapse_conf')
        self.collector.collect()
        self.assertPublishedMany(
            publish_mock,
            {
                'synapse.backend_count': 2,
                'synapse.frontend_count': 1,
            },
        )

    @patch.object(Collector, 'publish')
    @patch('synapse.open')
    def test_graceful_failure_on_json_error(self, open_mock, publish_mock):
        open_mock.return_value = self.getFixture('bad_metrics')
        self.collector.collect()
        self.assertPublishedMany(publish_mock, {})


################################################################################
if __name__ == "__main__":
    unittest.main()
