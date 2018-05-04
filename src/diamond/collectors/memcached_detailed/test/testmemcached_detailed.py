#!/usr/bin/python
# coding=utf-8
################################################################################

from test import CollectorTestCase
from test import get_collector_config
from test import unittest
from mock import Mock
from mock import patch

from diamond.collector import Collector
from memcached_detailed import MemcachedDetailedCollector

################################################################################


class TestMemcachedDetailedCollector(CollectorTestCase):
    def setUp(self):
        config = get_collector_config('MemcachedDetailedCollector', {
            'interval': 10,
            'hosts': ['localhost:11211'],
        })

        self.collector = MemcachedDetailedCollector(config, None)

    def assertPublishedManyMultiple(self, mock, expected_dict, expected_count=1):
        for key, values in expected_dict.iteritems():
            self.assertPublishedMultiple(mock, key, values, expected_count)

        mock.reset_mock()

    def assertPublishedMultiple(self, mock, key, values, expected_count):
        calls = filter(lambda x: x[0][0] == key, mock.call_args_list)

        actual_count = len(calls)
        message = '%s: actual number of calls %d, expected %d' % (key, actual_count, expected_count)
        self.assertEqual(actual_count, expected_count, message)

        actual_values = [call[0][1] for call in calls]

        for value in values:
            message = '%d not a value for key %s' % (value, key)
            self.assertIn(value, actual_values, message)

    def test_import(self):
        self.assertTrue(MemcachedDetailedCollector)

    @patch.object(Collector, 'publish')
    def test_real_stats(self, publish_mock):
        patch_raw_stats = patch.object(
            MemcachedDetailedCollector,
            'get_raw_stats',
            Mock(return_value=self.getFixture(
                'stats_simple').getvalue()))

        patch_raw_stats.start()
        self.collector.collect()
        patch_raw_stats.stop()


        metrics = {
            'memcache.detailed_get': [10, 12, 100],
            'memcache.detailed_set': [14, 1, 86],
            'memcache.detailed_hit': [0, 50],
            'memcache.detailed_del': [5, 9, 1],
        }

        self.assertPublishedManyMultiple(publish_mock, metrics, 5)

################################################################################
if __name__ == "__main__":
    unittest.main()
