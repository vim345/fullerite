#!/usr/bin/python
# coding=utf-8
################################################################################

from test import CollectorTestCase
from test import get_collector_config
from test import unittest
from test import run_only
from mock import Mock
from mock import patch

from diamond.collector import Collector
from proxysqlstat import ProxySQLCollector

################################################################################



class TestProxySQLCollector(CollectorTestCase):
    def setUp(self):
        config = get_collector_config('ProxySQLCollector', {})

        self.collector = ProxySQLCollector(config, None)
        self.collector.config['hosts'] = ['admin:admin@127.0.0.1:6032/']

    def test_import(self):
        self.assertTrue(ProxySQLCollector)

    def _verify_calls(self, actual, expected):
        assert len(actual) == len(expected)
        for call in actual:
            assert call[0] in expected
            expected.remove(call[0])

    @patch.object(ProxySQLCollector, 'connect', Mock(return_value=True))
    @patch.object(ProxySQLCollector, 'disconnect', Mock(return_value=True))
    @patch.object(Collector, 'publish')
    def test_global_status(self, publish_mock):
        with patch.object(
            ProxySQLCollector,
            'get_db_stats',
            Mock(return_value=[
                {'Value': '0', 'Variable_name': 'Active_transactions'},
                {'Value': '1', 'Variable_name': 'Client_Connections_connected'}
            ])
        ):
            self.collector.collect()
            calls = publish_mock.call_args_list
            expected = [('Active_transactions', 0.0), ('Client_Connections_connected', 1.0)]
            self._verify_calls(calls, expected)

    @patch.object(ProxySQLCollector, 'connect', Mock(return_value=True))
    @patch.object(ProxySQLCollector, 'disconnect', Mock(return_value=True))
    @patch.object(Collector, 'publish')
    def test_commands_counters(self, publish_mock):
        with patch.object(
            ProxySQLCollector,
            'get_db_stats',
            Mock(return_value=[
                {'Command': 'BEGIN', 'Total_Time_us': 500, 'Total_cnt': 1000},
                {'Command': 'SELECT', 'Total_Time_us': 25, 'Total_cnt': 50}
            ])
        ):
            self.collector.collect()
            calls = publish_mock.call_args_list
            expected = [
                ('BEGIN.command_total_count', 1000.0),
                ('BEGIN.command_total_time_us', 500.0),
                ('SELECT.command_total_count', 50.0),
                ('SELECT.command_total_time_us', 25.0),
            ]
            self._verify_calls(calls, expected)


################################################################################
if __name__ == "__main__":
    unittest.main()
