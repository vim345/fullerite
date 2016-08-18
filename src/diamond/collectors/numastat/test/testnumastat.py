#!/usr/bin/python
# coding=utf-8
################################################################################

from test import CollectorTestCase
from test import get_collector_config
from test import unittest
from mock import Mock
from mock import patch

try:
    from cStringIO import StringIO
except ImportError:
    from StringIO import StringIO

from diamond.collector import Collector
from numastat import NumastatCollector

################################################################################

def find_metric(metric_list, metric_name):
    return filter(lambda metric:metric["name"].find(metric_name) > -1, metric_list)


class TestNumastatCollector(CollectorTestCase):
    def setUp(self):
        config = get_collector_config('NumastatCollector', {
        })

        self.collector = NumastatCollector(config, None)

    def test_import(self):
        self.assertTrue(NumastatCollector)

    @patch.object(Collector, 'publish')
    def test_should_work_with_synthetic_data(self, publish_mock):
        stats = """numa_hit 7550016961
        numa_miss 65266974
        numa_foreign 31195725
        interleave_hit 34088
        local_node 7549989122
        other_node 65294813"""

        patch_open = patch('__builtin__.open', Mock(return_value=StringIO(stats)))

        patch_open.start()
        self.collector.collect()
        patch_open.stop()

        self.assertPublishedMany(publish_mock, {
            'numastat.numa_hit': 7550016961,
            'numastat.numa_miss': 65266974,
            'numastat.numa_foreign': 31195725,
            'numastat.interleave_hit': 34088,
            'numastat.local_node': 7549989122,
            'numastat.other_node': 65294813
        })

    @patch.object(Collector, 'publish')
    def test_should_work_with_real_data(self, publish_mock):
        NumastatCollector.NODE = self.getFixturePath('.')
        self.collector.collect()

        self.assertPublishedMany(publish_mock, {
            'numastat.numa_hit': 7553087993,
            'numastat.numa_miss': 65266974,
            'numastat.numa_foreign': 31195725,
            'numastat.interleave_hit': 34088,
            'numastat.local_node': 7553060154,
            'numastat.other_node': 65294813
        })

class TestNumastatCollectorDimensions(CollectorTestCase):
    def setUp(self):
        config = get_collector_config('NumastatCollector', {
        })

        self.collector = NumastatCollector(config, None)

    @patch.object(Collector, 'flush')
    def test_core_dimension_core_metrics(self, publish_mock):
        stats = """numa_hit 7550016961
        numa_miss 65266974
        numa_foreign 31195725
        interleave_hit 34088
        local_node 7549989122
        other_node 65294813"""

        patch_open = patch('__builtin__.open', Mock(return_value=StringIO(stats)))

        patch_open.start()
        self.collector.collect()
        patch_open.stop()


        for metric_name in ['numastat.numa_hit','numastat.numa_miss','numastat.numa_foreign','numastat.interleave_hit','numastat.local_node','numastat.other_node']:
            metrics = find_metric(self.collector.payload, metric_name)

            self.assertEqual(len(metrics), 1)
            self.assertTrue(metrics[0]['dimensions'].has_key('node'))

################################################################################
if __name__ == "__main__":
    unittest.main()
