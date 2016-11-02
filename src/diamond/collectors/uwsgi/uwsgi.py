# coding=utf-8

"""
Collect stats from uWSGI stats server
#### Dependencies
 * httplib
 * json
 * urlparse
"""

import collections
import httplib
import json
import re
import socket
import urlparse
import diamond.collector
from subprocess import Popen, PIPE


class UwsgiCollector(diamond.collector.Collector):


    def process_config(self):
        super(UwsgiCollector, self).process_config()
        if 'url' in self.config:
            self.config['urls'].append(self.config['url'])

        self.urls = {}
        if isinstance(self.config['urls'], basestring):
            self.config['urls'] = self.config['urls'].split(',')

        for url in self.config['urls']:
            # Handle the case where there is a trailing comma on the urls list
            if len(url) == 0:
                continue
            if ' ' in url:
                parts = url.split(' ')
                self.urls[parts[0]] = parts[1]
            else:
                self.urls[''] = url

    def get_default_config_help(self):
        config_help = super(UwsgiCollector, self).get_default_config_help()
        config_help.update({
            'urls': "Urls to server-status in auto format, comma seperated,"
            + " Format 'nickname http://host:port/, "
            + ", nickname http://host:port/, etc'",
            'processes' : "Command names of the uWSGI processes running"
            + " as a comma separated string",
        })
        return config_help

    def get_default_config(self):
        """
        Returns the default collector settings
        """
        config = super(UwsgiCollector, self).get_default_config()
        config.update({'urls': []})
        return config

    def read_pure_tcp(self, service_host, service_port):
        s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        s.connect((service_host, service_port))
        total_data = []
        while True:
            chunk = s.recv(8192)
            if not chunk:
                break
            total_data.append(chunk)
        data = ''.join(total_data)

    def collect(self):
        for nickname in self.urls.keys():
            url = self.urls[nickname]

            try:
                # Parse Url
                parts = urlparse.urlparse(url)

                # Parse host and port
                endpoint = parts.netloc.split(':')
                if len(endpoint) > 1:
                    service_host = endpoint[0]
                    service_port = int(endpoint[1])
                else:
                    service_host = endpoint[0]
                    service_port = 80

                if parts.scheme == 'http':

                    # Setup Connection
                    connection = httplib.HTTPConnection(service_host,
                                                        service_port)

                    if parts.params == '':
                        url = parts.path
                    else:
                        url = "%s?%s" % (parts.path, parts.params)

                    connection.request("GET", url)
                    response = connection.getresponse()
                    data = response.read()
                    connection.close()
                elif parts.scheme == 'tcp':
                    data = self.read_pure_tcp(service_host, service_port)
                else:
                    raise 'Unknown URL scheme'

            except Exception, e:
                print(e)
                self.log.error(
                    "Error retrieving uWSGI stats for host %s:%s, url '%s': %s",
                    service_host, str(service_port), url, e)
                continue

            stats = json.loads(data)
            counters = { 'IdleWorkers': 0, 'BusyWorkers': 0, 'SigWorkers': 0,
                         'PauseWorkers': 0, 'CheapWorkers': 0, 'UnknownStateWorkers': 0 }
            for worker in stats['workers']:
                status = worker['status']
                if status.find('sig') == 0:
                    status = 'sig'
                key = status.capitalize() + 'Workers'
                if key not in counters:
                    key = 'UnknownStateWorkers'
                counters[key] += 1

            for key in counters:
                self._publish(nickname, key, counters[key])

    def _publish(self, nickname, key, value):

        metrics = ['BusyWorkers', 'IdleWorkers', 'SigWorkers',
                   'PauseWorkers', 'CheapWorkers', 'UnknownStateWorkers']

        if key in metrics:
            # Get Metric Name
            metric_name = "%s" % re.sub('\s+', '', key)

            # Prefix with the nickname?
            if len(nickname) > 0:
                metric_name = nickname + '.' + metric_name

            # Get Metric Value
            metric_value = "%d" % float(value)

            # Publish Metric
            self.publish(metric_name, metric_value)

