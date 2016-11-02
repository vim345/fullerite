#!/usr/bin/python
# coding=utf-8
################################################################################

from test import CollectorTestCase
from test import get_collector_config
from test import unittest
from mock import Mock
from mock import patch

from diamond.collector import Collector
from uwsgi import UwsgiCollector
import httplib

################################################################################


class TestHTTPResponse(httplib.HTTPResponse):
    def __init__(self):
        pass

    def read(self):
        pass


class TestUwsgiCollector(CollectorTestCase):
    def setUp(self, config=None):
        if config is None:
            config = get_collector_config('UwsgiCollector', {
                'interval': '10',
                'url': 'http://www.example.com:80/'
            })
        else:
            config = get_collector_config('UwsgiCollector', config)

        self.collector = UwsgiCollector(config, configfile=config['collectors']['UwsgiCollector'])

        self.HTTPResponse = TestHTTPResponse()

        httplib.HTTPConnection.request = Mock(return_value=True)
        httplib.HTTPConnection.getresponse = Mock(
            return_value=self.HTTPResponse)

    def test_import(self):
        self.assertTrue(UwsgiCollector)

    @patch.object(Collector, 'publish')
    def test_should_work_with_synthetic_data(self, publish_mock):
        self.setUp()

        patch_read = patch.object(
            TestHTTPResponse,
            'read',
            Mock(return_value=self.getFixture(
                'status-json-fake-1').getvalue()))

        patch_read.start()
        self.collector.collect()
        patch_read.stop()

        self.assertPublishedMany(publish_mock, {
            'IdleWorkers': 0,
            'BusyWorkers': 0,
            'SigWorkers': 0,
            'CheapWorkers': 0,
            'PauseWorkers': 0,
            'UnknownStateWorkers': 0,
        })

        patch_read = patch.object(
            TestHTTPResponse,
            'read',
            Mock(return_value=self.getFixture(
                'status-json-fake-2').getvalue()))

        patch_read.start()
        self.collector.collect()
        patch_read.stop()

        self.assertPublishedMany(publish_mock, {
            'IdleWorkers': 1,
            'BusyWorkers': 3,
            'SigWorkers': 1,
            'CheapWorkers': 1,
            'PauseWorkers': 1,
            'UnknownStateWorkers': 0,
        })

        patch_read = patch.object(
            TestHTTPResponse,
            'read',
            Mock(return_value=self.getFixture(
                'status-json-fake-3').getvalue()))

        patch_read.start()
        self.collector.collect()
        patch_read.stop()

        self.assertPublishedMany(publish_mock, {
            'IdleWorkers': 0,
            'BusyWorkers': 1,
            'SigWorkers': 0,
            'CheapWorkers': 0,
            'PauseWorkers': 0,
            'UnknownStateWorkers': 1,
        })

        patch_read = patch.object(TestHTTPResponse,
            'read',
            Mock(return_value=self.getFixture(
                'status-json-fake-4').getvalue()))

        patch_read.start()
        try:
            self.collector.collect()
            self.fail('Should throw an exception')
        except ValueError:
            pass
        patch_read.stop()

    @patch.object(Collector, 'publish')
    def test_should_work_with_real_data(self, publish_mock):
        self.setUp()

        patch_read = patch.object(
            TestHTTPResponse,
            'read',
            Mock(return_value=self.getFixture(
                'status-json-live-1').getvalue()))

        patch_read.start()
        self.collector.collect()
        patch_read.stop()

        self.assertPublishedMany(publish_mock, {
            'IdleWorkers': 2,
            'BusyWorkers': 0,
            'SigWorkers': 0,
            'CheapWorkers': 0,
            'PauseWorkers': 0,
            'UnknownStateWorkers': 0,
        })

        patch_read = patch.object(
            TestHTTPResponse,
            'read',
            Mock(return_value=self.getFixture(
                'status-json-live-2').getvalue()))

        patch_read.start()
        self.collector.collect()
        patch_read.stop()

        self.assertPublishedMany(publish_mock, {
            'IdleWorkers': 1,
            'BusyWorkers': 1,
            'SigWorkers': 0,
            'CheapWorkers': 0,
            'PauseWorkers': 0,
            'UnknownStateWorkers': 0,
        })

    @patch.object(Collector, 'publish')
    def test_should_work_with_multiple_servers(self, publish_mock):
        self.setUp(config={
            'urls': [
                'nickname1 http://localhost:2081/',
                'nickname2 http://localhost:2081/',
            ],
        })

        patch_read = patch.object(
            TestHTTPResponse,
            'read',
            Mock(return_value=self.getFixture(
                'status-json-live-1').getvalue()))

        patch_read.start()
        self.collector.collect()
        patch_read.stop()

        self.assertPublishedMany(publish_mock, {})

        patch_read = patch.object(
            TestHTTPResponse,
            'read',
            Mock(return_value=self.getFixture(
                'status-json-live-4').getvalue()))

        patch_read.start()
        self.collector.collect()
        patch_read.stop()

        metrics = {
            'nickname1.IdleWorkers': 0,
            'nickname1.BusyWorkers': 2,
            'nickname1.SigWorkers': 0,
            'nickname1.CheapWorkers': 0,
            'nickname1.PauseWorkers': 0,
            'nickname1.UnknownStateWorkers': 0,

            'nickname2.IdleWorkers': 0,
            'nickname2.BusyWorkers': 2,
            'nickname2.SigWorkers': 0,
            'nickname2.CheapWorkers': 0,
            'nickname2.PauseWorkers': 0,
            'nickname2.UnknownStateWorkers': 0,
        }

        self.assertPublishedMany(publish_mock, metrics)

    @patch.object(Collector, 'publish')
    def test_sig0(self, publish_mock):
        self.setUp(config={
            'urls': 'vhost tcp://localhost:1789',
        })

        with patch.object(self.collector, 'read_pure_tcp', return_value=self.getFixture(
            'status-json-live-3').getvalue()):
            self.collector.collect()

        self.assertPublishedMany(publish_mock, {
            'vhost.IdleWorkers': 1,
            'vhost.BusyWorkers': 0,
            'vhost.SigWorkers': 1,
            'vhost.CheapWorkers': 0,
            'vhost.PauseWorkers': 0,
            'vhost.UnknownStateWorkers': 0,
        })

    @patch.object(Collector, 'publish')
    def test_issue_533(self, publish_mock):
        self.setUp(config={
            'urls': 'localhost http://localhost:80/server-status?auto,',
        })

        expected_urls = {'localhost': 'http://localhost:80/server-status?auto'}

        self.assertEqual(self.collector.urls, expected_urls)

    @patch.object(Collector, 'publish')
    def test_url_with_port(self, publish_mock):
        self.setUp(config={
            'urls': 'localhost http://localhost:80/server-status?auto',
        })

        expected_urls = {'localhost': 'http://localhost:80/server-status?auto'}

        self.assertEqual(self.collector.urls, expected_urls)

    @patch.object(Collector, 'publish')
    def test_url_without_port(self, publish_mock):
        self.setUp(config={
            'urls': 'localhost http://localhost/server-status?auto',
        })

        expected_urls = {'localhost': 'http://localhost/server-status?auto'}

        self.assertEqual(self.collector.urls, expected_urls)

    @patch.object(Collector, 'publish')
    def test_url_without_nickname(self, publish_mock):
        self.setUp(config={
            'urls': 'http://localhost/server-status?auto',
        })

        expected_urls = {'': 'http://localhost/server-status?auto'}

        self.assertEqual(self.collector.urls, expected_urls)

    @patch.object(Collector, 'publish')
    def test_issue_538(self, publish_mock):
        self.setUp(config={
            'enabled': True,
            'path_suffix': "",
            'ttl_multiplier': 2,
            'measure_collector_time': False,
            'byte_unit': 'byte',
            'urls': 'localhost http://localhost:80/server-status?auto',
        })

        expected_urls = {'localhost': 'http://localhost:80/server-status?auto'}

        self.assertEqual(self.collector.urls, expected_urls)

################################################################################
if __name__ == "__main__":
    unittest.main()
