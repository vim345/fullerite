# coding=utf-8

"""

Collects metrics from hacheck via its /status endpoint.
"""
import json
import urllib2

import diamond.collector


class HacheckCollector(diamond.collector.Collector):

    METRICS_BLACKLIST = set(['hacheck.uptime'])

    def get_default_config_help(self):
        config_help = super(HacheckCollector, self).get_default_config_help()
        config_help.update({
            'host': 'Hostname (default: localhost)',
            'port': 'Port (default: 6666)',
        })
        return config_help

    def get_default_config(self):
        config = super(HacheckCollector, self).get_default_config()
        config.update({
            'host': 'localhost',
            'port': 6666,
        })
        return config

    def collect(self):
        metrics = self.get_metrics()

        for k, v in metrics.iteritems():
            if k in self.METRICS_BLACKLIST:
                continue 

            self.publish(k, v)

    def get_metrics(self):
        url = 'http://{host}:{port}/status'.format(
            host=self.config['host'],
            port=self.config['port'],
        )

        try:
            response = urllib2.urlopen(url)
        except urllib2.HTTPError as e:
            self.log.error('Failed to get response from hacheck: {}'.format(e))
            return {}

        try:
            response_dict = json.load(response)
        except ValueError as e:
            self.log.error('Could not parse JSON from hacheck: {}'.format(e))
            return {}

        return self._flatten_dict(response_dict, prefix='hacheck')

    def _flatten_dict(self, to_flatten, prefix=''):
        new_dict = {}
        for k, v in to_flatten.iteritems():
            if len(prefix) > 0:
                k = prefix + '.' + k

            if isinstance(v, dict):
                new_dict.update(self._flatten_dict(v, prefix=k))
            else:
                new_dict[k] = v

        return new_dict
