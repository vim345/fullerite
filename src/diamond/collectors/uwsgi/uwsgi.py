# coding=utf-8

"""
Collect stats from uWSGI stats server
#### Dependencies
 * httplib
 * urlparse
"""

import collections
import httplib
import re
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
            # Handle the case where there is a trailing comman on the urls list
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
        config.update({
            'path':     'uwsgi',
            'processes': ['uwsgi'],
            'urls':     ['localhost http://127.0.0.1:2081/']
        })
        return config

    def collect(self):
        for nickname in self.urls.keys():
            url = self.urls[nickname]

            try:
                while True:

                    # Parse Url
                    parts = urlparse.urlparse(url)

                    # Parse host and port
                    endpoint = parts[1].split(':')
                    if len(endpoint) > 1:
                        service_host = endpoint[0]
                        service_port = int(endpoint[1])
                    else:
                        service_host = endpoint[0]
                        service_port = 80

                    # Setup Connection
                    connection = httplib.HTTPConnection(service_host,
                                                        service_port)

                    if parts[4] == '':
                        url = parts[2]
                    else:
                        url = "%s?%s" % (parts[2], parts[4])

                    connection.request("GET", url)
                    response = connection.getresponse()
                    data = response.read()
                    headers = dict(response.getheaders())
                    if ('location' not in headers
                            or headers['location'] == url):
                        connection.close()
                        break
                    url = headers['location']
                    connection.close()
            except Exception, e:
                print(e)
                self.log.error(
                    "Error retrieving uWSGI stats for host %s:%s, url '%s': %s",
                    service_host, str(service_port), url, e)
                continue

            counters = { 'IdleWorkers': 0, 'BusyWorkers': 0, 'SigWorkers': 0,
                         'PauseWorkers': 0, 'CheapWorkers': 0, 'UnknownStateWorkers': 0 }
            for line in data.split('\n'):
                if line:
                    line = line.strip()

                    if line.find('"status":"') == 0:
                        pieces = line.split('"')
                        status = pieces[3]
                        if status.find('sig') == 0:
                            status = 'sig'
                        key = status.capitalize() + "Workers"
                        if key in counters:
                            counters[key] = counters[key] + 1
                        else:
                            counters['UnknownStateWorkers'] = counters['UnknownStateWorkers'] + 1

            for key in counters:
                self._publish(nickname, key, counters[key])

        try:
            p = Popen('ps ax -o rss=,vsz=,comm='.split(), stdout=PIPE, stderr=PIPE)
            output, errors = p.communicate()

            if errors:
                self.log.error(
                    "Failed to open process: {0!s}".format(errors)
                )
            else:
                resident_memory = collections.defaultdict(list)
                virtual_memory = collections.defaultdict(list)
                for line in output.split('\n'):
                    if not line:
                        continue
                    (rss, vsz, proc) = line.strip('\n').split(None,2)
                    if proc in self.config['processes']:
                        resident_memory[proc].append(int(rss))
                        virtual_memory[proc].append(int(vsz))

                for proc in self.config['processes']:
                    metric_name = '.'.join([proc, 'WorkersResidentMemory'])
                    memory_rss = resident_memory.get(proc, [0])
                    metric_value = sum(memory_rss) / len(memory_rss)

                    self.publish(metric_name, metric_value)
                    metric_name = '.'.join([proc, 'WorkersVirtualMemory'])
                    memory_vsz = virtual_memory.get(proc, [0])
                    metric_value = sum(memory_vsz) / len(memory_vsz)

                    self.publish(metric_name, metric_value)
        except Exception as e:
            self.log.error(
                "Failed because: {0!s}".format(e)
            )

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

