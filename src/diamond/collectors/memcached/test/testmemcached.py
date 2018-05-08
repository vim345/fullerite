#!/usr/bin/python
# coding=utf-8
################################################################################

from test import CollectorTestCase
from test import get_collector_config
from test import unittest
from mock import Mock
from mock import patch

from diamond.collector import Collector
from memcached import MemcachedCollector

################################################################################


class TestMemcachedCollector(CollectorTestCase):
    def setUp(self):
        config = get_collector_config('MemcachedCollector', {
            'interval': 10,
            'hosts': ['localhost:11211'],
        })

        self.collector = MemcachedCollector(config, None)

    def test_import(self):
        self.assertTrue(MemcachedCollector)

    @patch.object(Collector, 'publish')
    def test_should_work_with_real_data(self, publish_mock):
        patch_raw_stats = patch.object(
            MemcachedCollector,
            'get_raw_stats',
            Mock(return_value=self.getFixture(
                'stats').getvalue()))

        patch_raw_stats.start()
        self.collector.collect()
        patch_raw_stats.stop()

        metrics = {
            'memcache.reclaimed': 0.000000,
            'memcache.expired_unfetched': 0.000000,
            'memcache.hash_is_expanding': 0.000000,
            'memcache.cas_hits': 0.000000,
            'memcache.uptime': 25763,
            'memcache.touch_hits': 0.000000,
            'memcache.delete_misses': 0.000000,
            'memcache.listen_disabled_num': 0.000000,
            'memcache.cas_misses': 0.000000,
            'memcache.decr_hits': 0.000000,
            'memcache.cmd_touch': 0.000000,
            'memcache.incr_hits': 0.000000,
            'memcache.auth_cmds': 0.000000,
            'memcache.limit_maxbytes': 67108864.000000,
            'memcache.bytes_written': 0.000000,
            'memcache.incr_misses': 0.000000,
            'memcache.rusage_system': 0.195071,
            'memcache.total_items': 0.000000,
            'memcache.cmd_get': 0.000000,
            'memcache.curr_connections': 10.000000,
            'memcache.touch_misses': 0.000000,
            'memcache.threads': 4.000000,
            'memcache.total_connections': 11,
            'memcache.cmd_set': 0.000000,
            'memcache.curr_items': 0.000000,
            'memcache.conn_yields': 0.000000,
            'memcache.get_misses': 0.000000,
            'memcache.reserved_fds': 20.000000,
            'memcache.bytes_read': 7,
            'memcache.hash_bytes': 524288.000000,
            'memcache.evicted_unfetched': 0.000000,
            'memcache.cas_badval': 0.000000,
            'memcache.cmd_flush': 0.000000,
            'memcache.evictions': 0.000000,
            'memcache.bytes': 0.000000,
            'memcache.connection_structures': 11.000000,
            'memcache.hash_power_level': 16.000000,
            'memcache.auth_errors': 0.000000,
            'memcache.rusage_user': 0.231516,
            'memcache.delete_hits': 0.000000,
            'memcache.decr_misses': 0.000000,
            'memcache.get_hits': 0.000000,
            'memcache.repcached_qi_free': 0.000000,
        }

        self.setDocExample(collector=self.collector.__class__.__name__,
                           metrics=metrics,
                           defaultpath=self.collector.config['path'])
        self.assertPublishedMany(publish_mock, metrics)

################################################################################
if __name__ == "__main__":
    unittest.main()
