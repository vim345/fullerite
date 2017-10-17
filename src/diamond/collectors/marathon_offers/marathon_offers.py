# coding=utf-8

"""

Collects metrics about marathon offers, specifically what offers are being rejected.
By default queries localhost:5052.
Will check if it's running on the master, and no-op if it's not.
"""

import diamond.collector
import json
import socket
import requests

from requests.auth import HTTPBasicAuth


class MarathonOffersCollector(diamond.collector.Collector):

    QUEUE_PATH = "/v2/queue"
    LEADER_PATH = "/v2/leader"

    def __init__(self, *args, **kwargs):
        super(MarathonOffersCollector, self).__init__(*args, **kwargs)
        self.hostname = socket.gethostname()

    def get_default_config_help(self):
        config_help = super(MarathonOffersCollector, self).get_default_config_help()
        config_help.update({
            'host': 'hostname',
            'port': 'port',
            'username': 'username for marathon api',
            'password': 'password for marathon api',
        })
        return config_help

    def get_default_config(self):
        config = super(MarathonOffersCollector, self).get_default_config()
        config.update({
            'host': 'localhost',
            'port': 5052,
        })
        return config

    def collect(self):
        if not self._am_i_leader():
            return
        queue = self._get(self.QUEUE_PATH)
        for queue_item in queue['queue']:
            summary = queue_item['processedOffersSummary']
            app_name = queue_item['app']['id'][1:].split('.')[0:2]
            self._publish('count', queue_item['count'], app_name)
            for key in ['processedOffersCount', 'unusedOffersCount']:
                self._publish('processedOffersSummary.%s' % key, summary[key], app_name)
            for key in ['rejectSummaryLastOffers', 'rejectSummaryLaunchAttempt']:
                for reason_dict in summary[key]:
                    reason = reason_dict['reason']
                    for t in ['processed', 'declined']:
                        self._publish(
                            'processedOffersSummary.%s.%s.%s' % (key, reason, t),
                            reason_dict[t],
                            app_name,
                        )

    def _publish(self, name, value, app_name):
        self.dimensions = {'app': app_name}
        self.publish('zzz-bentley.%s' % name, value)

    def _get_metrics(self):
        try:
            queue = self._get(QUEUE_PATH)
        except Exception as err:
            self.log.error('Unable to read queue from leader: %s' % err)
            return {}

    def _am_i_leader(self):
        try:
            leader = self._get(self.LEADER_PATH)['leader'].split(':')[0]
            return leader == self.hostname
        except Exception as err:
            self.log.error('Unable to read leader json response: %s' % err)
            return False

    def _get(self, path):
        auth = HTTPBasicAuth(self.config['user'], self.config['password'])
        url = "http://%s:%s/%s" % (self.config['host'],
                                   self.config['port'],
                                   path)
        return requests.get(url, auth=auth).json()
