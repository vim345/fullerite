#!/usr/bin/python
# coding=utf-8
################################################################################

from test import CollectorTestCase
from test import get_collector_config
from test import unittest
from mock import Mock
from mock import mock_open
from mock import patch
from mock import call

from diamond.collector import Collector
from proc_pid import ProcPidCollector
import subprocess

################################################################################


class TestProcPidCollector(CollectorTestCase):
    TEST_CONFIG = {'pid_paths': {'service1': '/var/run/service1.pid',
                                 'service2': '/var/run/service2.pid'}}

    def setUp(self):
        config = get_collector_config('ProcPidCollector',
                                      self.TEST_CONFIG)

        self.collector = ProcPidCollector(config, None)
        self.collector.config.update(self.TEST_CONFIG)

    def test_import(self):
        self.assertTrue(ProcPidCollector)

    def test_get_config_help(self):
        self.collector.get_default_config_help()

    def test_get_default_config(self):
        ret = self.collector.get_default_config()
        assert ret['ls'] == '/bin/ls'
        assert ret['sudo_cmd'] == '/usr/bin/sudo'

    def test_get_proc_path(self):
        mock_open_file = mock_open(read_data='2312')
        with patch('proc_pid.open', mock_open_file):
            ret = self.collector.get_proc_path('/var/whatever')
            mock_open_file.assert_called_with('/var/whatever')
            assert ret == '/proc/2312'

    @patch('proc_pid.subprocess.Popen')
    def test_get_fds(self, mock_popen):
        mock_proc_ret = ("0\n1\n2\n3\n4\n5\n6", None)
        mock_popen.return_value = Mock(communicate=Mock(return_value=mock_proc_ret))
        ret = self.collector.get_fds('/proc/2312')
        mock_popen.assert_called_with(['/usr/bin/sudo', '/bin/ls', '/proc/2312/fd'],
                                      stdout=subprocess.PIPE,
                                      stderr=subprocess.PIPE)
        assert ret == 7

    def mock_get_fds_side(self, path):
        if '1' in path:
            return 3
        else:
            return 5

    def mock_get_proc_path_side(self, path):
        if 'service1' in path:
            return '/proc/1'
        else:
            return '/proc/2'

    @patch.object(Collector, 'publish')
    @patch('proc_pid.ProcPidCollector.get_fds')
    @patch('proc_pid.ProcPidCollector.get_proc_path')
    def test_collect(self, mock_get_proc_path, mock_get_fds, mock_publish):
        mock_get_fds.side_effect = self.mock_get_fds_side
        mock_get_proc_path.side_effect = self.mock_get_proc_path_side
        self.collector.collect()
        mock_get_proc_path.assert_has_calls([call('/var/run/service1.pid'),
                                             call('/var/run/service2.pid')],
                                            any_order=True)
        mock_get_fds.assert_has_calls([call('/proc/1'),
                                       call('/proc/2')],
                                      any_order=True)
        self.assertPublished(mock_publish, 'proc_pid_stats.fds', [5, 3])



################################################################################
if __name__ == "__main__":
    unittest.main()
