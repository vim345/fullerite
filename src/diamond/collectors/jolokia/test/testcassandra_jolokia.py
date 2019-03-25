#!/usr/bin/python
# coding=utf-8
################################################################################

from test import CollectorTestCase
from test import get_collector_config
from test import unittest
from mock import Mock
from mock import patch
from mock import mock_open

from diamond.collector import Collector

from cassandra_jolokia import CassandraJolokiaCollector
import re
import json

################################################################################

def find_metric(metric_list, metric_name):
    return filter(lambda metric:metric["name"].find(metric_name) > -1, metric_list)

def find_by_dimension(metric_list, key, val):
    return filter(lambda metric:metric["dimensions"][key] == val, metric_list)[0]

def list_request(host, port):
    return {'value': {'com.yelp':'bla'}, 'status':200}

def read_host_list():
    return {
        'cass_1': { 'host': '10.0.0.1', 'port': 8999 },
        'cass_2': { 'host': '10.0.0.2', 'port': 8999 }
    }

class TestCassandraJolokiaCollector(CollectorTestCase):
    def setUp(self):
        config = get_collector_config('CassandraJolokiaCollector', {})

        self.collector = CassandraJolokiaCollector(config, None)

    def test_import(self):
        self.assertTrue(CassandraJolokiaCollector)

    @patch.object(Collector, 'flush')
    def test_should_create_dimension(self, publish_mock):
        self.collector.list_request = list_request

        def se(url):
            return self.getFixture("metrics.json")

        patch_urlopen = patch('urllib2.urlopen', Mock(side_effect=se))

        with patch_urlopen:
            self.collector.collect()
        self.assertEquals(len(self.collector.payload), 3827)

        metrics = find_metric(self.collector.payload, "org.apache.cassandra.metrics.ColumnFamily.LiveSSTableCount")
        self.assertNotEqual(len(metrics), 0)
        metric = find_by_dimension(metrics, "type", "compaction_history")
        self.assertEquals(metric["type"], "GAUGE")

        pending_task = find_metric(self.collector.payload,
                                   "org.apache.cassandra.metrics.CommitLog4.2.PendingTasks")
        self.assertNotEqual(len(pending_task), 0)

    def test_patch_host_list_should_filter_out_non_cassandra_hosts(self):
        data = {
            "services": {
                "service1": { "host": "10.0.0.1" },
                "cassandra_1.dc": { "host": "10.0.0.2"}
            }
        }
        with patch("__builtin__.open", mock_open(read_data=json.dumps(data))):
            hosts = self.collector.read_host_list()
        self.assertEqual(
            hosts,
            {
                'cassandra_1': { 'host': '10.0.0.2', 'port': self.collector.config['port'] }
            }
        )

    @patch.object(Collector, 'flush')
    def test_should_have_cassandra_cluster_dimension_in_multi_hosts_mode(self, publish_mock):
        self.collector.list_request = list_request
        self.collector.config['multiple_hosts_mode'] = True
        self.collector.read_host_list = read_host_list

        def se(url):
            return self.getFixture("metrics.json")

        patch_urlopen = patch('urllib2.urlopen', Mock(side_effect=se))

        with patch_urlopen:
            self.collector.collect()
        self.collector.config['multiple_hosts_mode'] = False
        self.assertEquals(len(self.collector.payload), 3827 * 2)

        metrics = find_metric(self.collector.payload, "org.apache.cassandra.metrics.ColumnFamily.LiveSSTableCount")
        self.assertNotEqual(len(metrics), 0)
        metric = find_by_dimension(metrics, "cassandra_cluster", "cass_1")
        self.assertEquals(metric["type"], "GAUGE")
        metric = find_by_dimension(metrics, "cassandra_cluster", "cass_2")
        self.assertEquals(metric["type"], "GAUGE")

    @patch.object(Collector, 'flush')
    def test_should_create_type(self, publish_mock):
        self.collector.list_request = list_request
        def se(url):
            return self.getFixture("metrics.json")

        patch_urlopen = patch('urllib2.urlopen', Mock(side_effect=se))

        with patch_urlopen:
            self.collector.collect()
        self.assertEquals(len(self.collector.payload), 3827)

        metrics = find_metric(self.collector.payload, "org.apache.cassandra.metrics.ColumnFamily.CoordinatorReadLatency.count")
        self.assertNotEqual(len(metrics), 0)
        metric = find_by_dimension(metrics, "keyspace", "OpsCenter")
        self.assertEquals(metric["type"], "CUMCOUNTER")

    @patch.object(Collector, 'flush')
    def test_mbean_blacklisting(self, publish_mock):
        def se(url):
            if url.find("org.apache.cassandra.metrics") > 0:
                return self.getFixture("metrics.json")
            elif url.find("list/org.apache.cassandra.db") > 0:
                return self.getFixture("cas_db.json")
            elif url.find("org.apache.cassandra.db:type=StorageService") > 0:
                return Exception('storage service should be blacklisted')
            elif url.find("list?ifModifiedSince") > 0:
                return self.getFixture("cas_list.json")
            else:
                return self.getFixture("storage_proc.json")
        patch_urlopen = patch('urllib2.urlopen', Mock(side_effect=se))
        self.collector.config['mbean_blacklist'] = [
            'org.apache.cassandra.db:type=StorageService'
        ]

        with patch_urlopen:
            self.collector.collect()
        metrics = find_metric(self.collector.payload, "org.apache.cassandra.db.StorageProxy.cascontentiontimeout")
        self.assertNotEqual(len(metrics), 0)


################################################################################
if __name__ == "__main__":
    unittest.main()
