#!/usr/bin/python
# coding=utf-8
################################################################################

from test import CollectorTestCase
from test import get_collector_config
from test import unittest
from test import run_only
from mock import patch

from diamond.collector import Collector
from puppetagent import PuppetAgentCollector

################################################################################


def run_only_if_yaml_is_available(func):
    try:
        import yaml
    except ImportError:
        yaml = None
    pred = lambda: yaml is not None
    return run_only(func, pred)


class TestPuppetAgentCollector(CollectorTestCase):
    def setUp(self):
        config = get_collector_config('PuppetAgentCollector', {
            'interval': 10,
            'yaml_path': self.getFixturePath('last_run_summary.yaml')
        })

        self.collector = PuppetAgentCollector(config, None)

    def test_import(self):
        self.assertTrue(PuppetAgentCollector)

    @run_only_if_yaml_is_available
    @patch.object(Collector, 'publish')
    def test(self, publish_mock):

        self.collector.collect()

        metrics = {
            'puppet.changes.total': 1,
            'puppet.events.failure': 0,
            'puppet.events.success': 1,
            'puppet.events.total': 1,
            'puppet.resources.changed': 1,
            'puppet.resources.failed': 0,
            'puppet.resources.failed_to_restart': 0,
            'puppet.resources.out_of_sync': 1,
            'puppet.resources.restarted': 0,
            'puppet.resources.scheduled': 0,
            'puppet.resources.skipped': 6,
            'puppet.resources.total': 439,
            'puppet.time.anchor': 0.009641,
            'puppet.time.augeas': 1.286514,
            'puppet.time.config_retrieval': 8.06442093849182,
            'puppet.time.cron': 0.00089,
            'puppet.time.exec': 9.780635,
            'puppet.time.file': 1.729348,
            'puppet.time.filebucket': 0.000633,
            'puppet.time.firewall': 0.007807,
            'puppet.time.group': 0.013421,
            'puppet.time.last_run': 1377125556,
            'puppet.time.mailalias': 0.000335,
            'puppet.time.mount': 0.002749,
            'puppet.time.package': 1.831337,
            'puppet.time.resources': 0.000371,
            'puppet.time.service': 0.734021,
            'puppet.time.ssh_authorized_key': 0.017625,
            'puppet.time.total': 23.5117989384918,
            'puppet.time.user': 0.02927,
            'puppet.version.config': 1377123965,
        }

        unpublished_metrics = {
            'puppet.version.puppet': '2.7.14',
        }

        self.setDocExample(collector=self.collector.__class__.__name__,
                           metrics=metrics,
                           defaultpath=self.collector.config['path'])

        self.assertPublishedMany(publish_mock, metrics)
        self.assertUnpublishedMany(publish_mock, unpublished_metrics)

################################################################################
if __name__ == "__main__":
    unittest.main()
